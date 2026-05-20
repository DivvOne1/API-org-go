package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"org-structure-api/internal/models"
	"org-structure-api/internal/service"
)

type stubDepartmentService struct {
	createDepartmentFn func(ctx context.Context, input service.CreateDepartmentInput) (*models.Department, error)
	getDepartmentTreeFn func(ctx context.Context, departmentID, depth int, includeEmployees bool) (*models.DepartmentNode, error)
}

func (s stubDepartmentService) CreateDepartment(ctx context.Context, input service.CreateDepartmentInput) (*models.Department, error) {
	return s.createDepartmentFn(ctx, input)
}

func (s stubDepartmentService) CreateEmployee(context.Context, int, service.CreateEmployeeInput) (*models.Employee, error) {
	panic("not implemented")
}

func (s stubDepartmentService) GetDepartmentTree(ctx context.Context, departmentID, depth int, includeEmployees bool) (*models.DepartmentNode, error) {
	return s.getDepartmentTreeFn(ctx, departmentID, depth, includeEmployees)
}

func (s stubDepartmentService) UpdateDepartment(context.Context, int, service.UpdateDepartmentInput) (*models.Department, error) {
	panic("not implemented")
}

func (s stubDepartmentService) DeleteDepartment(context.Context, int, service.DeleteDepartmentInput) error {
	panic("not implemented")
}

func TestCreateDepartmentReturnsCreatedDepartment(t *testing.T) {
	handler := NewHandler(stubDepartmentService{
		createDepartmentFn: func(ctx context.Context, input service.CreateDepartmentInput) (*models.Department, error) {
			if input.Name != "Backend" {
				t.Fatalf("expected Backend, got %q", input.Name)
			}

			return &models.Department{
				ID:        1,
				Name:      "Backend",
				CreatedAt: time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC),
			}, nil
		},
	})

	body, _ := json.Marshal(map[string]any{"name": "Backend"})
	req := httptest.NewRequest(http.MethodPost, "/departments/", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}
}

func TestGetDepartmentRejectsInvalidDepth(t *testing.T) {
	handler := NewHandler(stubDepartmentService{
		createDepartmentFn: func(ctx context.Context, input service.CreateDepartmentInput) (*models.Department, error) {
			return nil, nil
		},
		getDepartmentTreeFn: func(ctx context.Context, departmentID, depth int, includeEmployees bool) (*models.DepartmentNode, error) {
			return nil, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/departments/1?depth=6", nil)
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestGetDepartmentReturnsTree(t *testing.T) {
	handler := NewHandler(stubDepartmentService{
		createDepartmentFn: func(ctx context.Context, input service.CreateDepartmentInput) (*models.Department, error) {
			return nil, nil
		},
		getDepartmentTreeFn: func(ctx context.Context, departmentID, depth int, includeEmployees bool) (*models.DepartmentNode, error) {
			if departmentID != 1 || depth != 2 || !includeEmployees {
				t.Fatalf("unexpected params: id=%d depth=%d include=%v", departmentID, depth, includeEmployees)
			}

			return &models.DepartmentNode{
				Department: models.Department{ID: 1, Name: "Root"},
				Children: []models.DepartmentNode{
					{Department: models.Department{ID: 2, Name: "Child"}},
				},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/departments/1?depth=2&include_employees=true", nil)
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}
