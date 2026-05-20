package repository

import (
	"context"

	"gorm.io/gorm"

	"org-structure-api/internal/models"
)

type DepartmentRepository interface {
	CreateDepartment(ctx context.Context, department *models.Department) error
	GetDepartmentByID(ctx context.Context, id int) (*models.Department, error)
	DepartmentExistsByName(ctx context.Context, parentID *int, name string, excludeID *int) (bool, error)
	ListChildren(ctx context.Context, parentID int) ([]models.Department, error)
	ListEmployeesByDepartmentID(ctx context.Context, departmentID int) ([]models.Employee, error)
	ReassignEmployeesAndDeleteSubtree(ctx context.Context, subtreeRootID int, departmentIDs []int, reassignToDepartmentID int) error
	CreateEmployee(ctx context.Context, employee *models.Employee) error
	UpdateDepartment(ctx context.Context, department *models.Department) error
	DeleteDepartmentCascade(ctx context.Context, departmentID int) error
}

type departmentRepository struct {
	db *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) DepartmentRepository {
	return &departmentRepository{db: db}
}
