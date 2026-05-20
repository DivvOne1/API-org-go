package service

import (
	"context"
	"errors"
	"testing"

	"org-structure-api/internal/models"
)

type fakeDepartmentRepository struct {
	departments map[int]*models.Department
	children    map[int][]models.Department
}

func (f *fakeDepartmentRepository) CreateDepartment(context.Context, *models.Department) error {
	return nil
}

func (f *fakeDepartmentRepository) GetDepartmentByID(_ context.Context, id int) (*models.Department, error) {
	return f.departments[id], nil
}

func (f *fakeDepartmentRepository) DepartmentExistsByName(context.Context, *int, string, *int) (bool, error) {
	return false, nil
}

func (f *fakeDepartmentRepository) ListChildren(_ context.Context, parentID int) ([]models.Department, error) {
	return f.children[parentID], nil
}

func (f *fakeDepartmentRepository) ListEmployeesByDepartmentID(context.Context, int) ([]models.Employee, error) {
	return nil, nil
}

func (f *fakeDepartmentRepository) ReassignEmployeesAndDeleteSubtree(context.Context, int, []int, int) error {
	return nil
}

func (f *fakeDepartmentRepository) CreateEmployee(context.Context, *models.Employee) error {
	return nil
}

func (f *fakeDepartmentRepository) UpdateDepartment(context.Context, *models.Department) error {
	return nil
}

func (f *fakeDepartmentRepository) DeleteDepartmentCascade(context.Context, int) error {
	return nil
}

func TestUpdateDepartmentRejectsCycle(t *testing.T) {
	rootID := 1
	childID := 2
	grandchildID := 3

	repo := &fakeDepartmentRepository{
		departments: map[int]*models.Department{
			rootID:       {ID: rootID, Name: "Root"},
			childID:      {ID: childID, Name: "Child", ParentID: &rootID},
			grandchildID: {ID: grandchildID, Name: "Grandchild", ParentID: &childID},
		},
	}

	svc := NewDepartmentService(repo)
	newParentID := grandchildID

	_, err := svc.UpdateDepartment(context.Background(), rootID, UpdateDepartmentInput{
		ParentID:  &newParentID,
		ParentSet: true,
	})
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
	if !errors.Is(err, ErrCycleDetected) {
		t.Fatalf("expected ErrCycleDetected, got %v", err)
	}
}
