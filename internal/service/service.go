package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"org-structure-api/internal/models"
	"org-structure-api/internal/repository"
)

const maxDepartmentDepth = 5

type CreateDepartmentInput struct {
	Name     string
	ParentID *int
}

type CreateEmployeeInput struct {
	FullName string
	Position string
	HiredAt  *time.Time
}

type UpdateDepartmentInput struct {
	Name      string
	NameSet   bool
	ParentID  *int
	ParentSet bool
}

type DeleteDepartmentInput struct {
	Mode                   string
	ReassignToDepartmentID *int
}

type DepartmentService interface {
	CreateDepartment(ctx context.Context, input CreateDepartmentInput) (*models.Department, error)
	CreateEmployee(ctx context.Context, departmentID int, input CreateEmployeeInput) (*models.Employee, error)
	GetDepartmentTree(ctx context.Context, departmentID, depth int, includeEmployees bool) (*models.DepartmentNode, error)
	UpdateDepartment(ctx context.Context, departmentID int, input UpdateDepartmentInput) (*models.Department, error)
	DeleteDepartment(ctx context.Context, departmentID int, input DeleteDepartmentInput) error
}

type departmentService struct {
	repo repository.DepartmentRepository
}

func NewDepartmentService(repo repository.DepartmentRepository) DepartmentService {
	return &departmentService{repo: repo}
}

func (s *departmentService) CreateDepartment(ctx context.Context, input CreateDepartmentInput) (*models.Department, error) {
	name, err := normalizeRequired(input.Name, "name")
	if err != nil {
		return nil, err
	}

	if input.ParentID != nil {
		parent, err := s.repo.GetDepartmentByID(ctx, *input.ParentID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, ErrNotFound
		}
	}

	exists, err := s.repo.DepartmentExistsByName(ctx, input.ParentID, name, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrConflict
	}

	department := &models.Department{
		Name:     name,
		ParentID: input.ParentID,
	}
	if err := s.repo.CreateDepartment(ctx, department); err != nil {
		return nil, err
	}

	return department, nil
}

func (s *departmentService) CreateEmployee(ctx context.Context, departmentID int, input CreateEmployeeInput) (*models.Employee, error) {
	department, err := s.repo.GetDepartmentByID(ctx, departmentID)
	if err != nil {
		return nil, err
	}
	if department == nil {
		return nil, ErrNotFound
	}

	fullName, err := normalizeRequired(input.FullName, "full_name")
	if err != nil {
		return nil, err
	}
	position, err := normalizeRequired(input.Position, "position")
	if err != nil {
		return nil, err
	}

	employee := &models.Employee{
		DepartmentID: departmentID,
		FullName:     fullName,
		Position:     position,
		HiredAt:      input.HiredAt,
	}
	if err := s.repo.CreateEmployee(ctx, employee); err != nil {
		return nil, err
	}

	return employee, nil
}

func (s *departmentService) GetDepartmentTree(ctx context.Context, departmentID, depth int, includeEmployees bool) (*models.DepartmentNode, error) {
	department, err := s.repo.GetDepartmentByID(ctx, departmentID)
	if err != nil {
		return nil, err
	}
	if department == nil {
		return nil, ErrNotFound
	}

	if depth < 0 {
		depth = 0
	}
	if depth > maxDepartmentDepth {
		depth = maxDepartmentDepth
	}

	return s.buildDepartmentNode(ctx, *department, depth, includeEmployees)
}

func (s *departmentService) UpdateDepartment(ctx context.Context, departmentID int, input UpdateDepartmentInput) (*models.Department, error) {
	department, err := s.repo.GetDepartmentByID(ctx, departmentID)
	if err != nil {
		return nil, err
	}
	if department == nil {
		return nil, ErrNotFound
	}

	if input.NameSet {
		name, err := normalizeRequired(input.Name, "name")
		if err != nil {
			return nil, err
		}
		department.Name = name
	}

	if input.ParentSet {
		parentID := input.ParentID
		if parentID != nil && *parentID == department.ID {
			return nil, ErrSelfParent
		}
		if parentID != nil {
			parent, err := s.repo.GetDepartmentByID(ctx, *parentID)
			if err != nil {
				return nil, err
			}
			if parent == nil {
				return nil, ErrNotFound
			}
			if err := s.ensureNoCycle(ctx, department.ID, *parentID); err != nil {
				return nil, err
			}
		}
		department.ParentID = parentID
	}

	exists, err := s.repo.DepartmentExistsByName(ctx, department.ParentID, department.Name, &department.ID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrConflict
	}

	if err := s.repo.UpdateDepartment(ctx, department); err != nil {
		return nil, err
	}

	return department, nil
}

func (s *departmentService) DeleteDepartment(ctx context.Context, departmentID int, input DeleteDepartmentInput) error {
	department, err := s.repo.GetDepartmentByID(ctx, departmentID)
	if err != nil {
		return err
	}
	if department == nil {
		return ErrNotFound
	}

	switch input.Mode {
	case "cascade":
		return s.repo.DeleteDepartmentCascade(ctx, departmentID)
	case "reassign":
		if input.ReassignToDepartmentID == nil {
			return ErrMissingReassign
		}
		targetID := *input.ReassignToDepartmentID
		if targetID == departmentID {
			return ErrBadRequest
		}
		target, err := s.repo.GetDepartmentByID(ctx, targetID)
		if err != nil {
			return err
		}
		if target == nil {
			return ErrNotFound
		}
		if err := s.ensureNoCycle(ctx, departmentID, targetID); err != nil {
			if errors.Is(err, ErrCycleDetected) {
				return ErrBadRequest
			}
			return err
		}

		subtreeIDs, err := s.collectSubtreeDepartmentIDs(ctx, departmentID)
		if err != nil {
			return err
		}

		return s.repo.ReassignEmployeesAndDeleteSubtree(ctx, departmentID, subtreeIDs, targetID)
	default:
		return ErrInvalidDeleteMode
	}
}

func (s *departmentService) buildDepartmentNode(ctx context.Context, department models.Department, depth int, includeEmployees bool) (*models.DepartmentNode, error) {
	node := &models.DepartmentNode{
		Department: department,
		Children:   make([]models.DepartmentNode, 0),
	}

	if includeEmployees {
		employees, err := s.repo.ListEmployeesByDepartmentID(ctx, department.ID)
		if err != nil {
			return nil, err
		}
		node.Employees = employees
	}

	if depth == 0 {
		return node, nil
	}

	children, err := s.repo.ListChildren(ctx, department.ID)
	if err != nil {
		return nil, err
	}

	for _, child := range children {
		childNode, err := s.buildDepartmentNode(ctx, child, depth-1, includeEmployees)
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, *childNode)
	}

	return node, nil
}

func (s *departmentService) collectSubtreeDepartmentIDs(ctx context.Context, rootID int) ([]int, error) {
	ids := []int{rootID}

	children, err := s.repo.ListChildren(ctx, rootID)
	if err != nil {
		return nil, err
	}

	for _, child := range children {
		childIDs, err := s.collectSubtreeDepartmentIDs(ctx, child.ID)
		if err != nil {
			return nil, err
		}
		ids = append(ids, childIDs...)
	}

	return ids, nil
}

func (s *departmentService) ensureNoCycle(ctx context.Context, departmentID, newParentID int) error {
	currentID := newParentID
	for {
		if currentID == departmentID {
			return ErrCycleDetected
		}

		current, err := s.repo.GetDepartmentByID(ctx, currentID)
		if err != nil {
			return err
		}
		if current == nil || current.ParentID == nil {
			return nil
		}

		currentID = *current.ParentID
	}
}

func normalizeRequired(value, field string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", validationError(field+" must not be empty")
	}
	if len([]rune(trimmed)) > 200 {
		return "", validationError(field+" must be between 1 and 200 characters")
	}

	return trimmed, nil
}

func validationError(message string) error {
	return ValidationError{Message: message}
}
