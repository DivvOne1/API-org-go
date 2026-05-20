package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"org-structure-api/internal/models"
)

func (r *departmentRepository) CreateDepartment(ctx context.Context, department *models.Department) error {
	return r.db.WithContext(ctx).Create(department).Error
}

func (r *departmentRepository) GetDepartmentByID(ctx context.Context, id int) (*models.Department, error) {
	var department models.Department
	err := r.db.WithContext(ctx).First(&department, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &department, nil
}

func (r *departmentRepository) DepartmentExistsByName(ctx context.Context, parentID *int, name string, excludeID *int) (bool, error) {
	query := r.db.WithContext(ctx).Model(&models.Department{}).Where("name = ?", name)
	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}
	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *departmentRepository) ListChildren(ctx context.Context, parentID int) ([]models.Department, error) {
	var departments []models.Department
	err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Order("name ASC").
		Find(&departments).Error
	if err != nil {
		return nil, err
	}

	return departments, nil
}

func (r *departmentRepository) ListEmployeesByDepartmentID(ctx context.Context, departmentID int) ([]models.Employee, error) {
	var employees []models.Employee
	err := r.db.WithContext(ctx).
		Where("department_id = ?", departmentID).
		Order("full_name ASC").
		Find(&employees).Error
	if err != nil {
		return nil, err
	}

	return employees, nil
}

func (r *departmentRepository) ReassignEmployeesAndDeleteSubtree(ctx context.Context, subtreeRootID int, departmentIDs []int, reassignToDepartmentID int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Employee{}).
			Where("department_id IN ?", departmentIDs).
			Update("department_id", reassignToDepartmentID).Error; err != nil {
			return err
		}

		return tx.Delete(&models.Department{}, subtreeRootID).Error
	})
}

func (r *departmentRepository) CreateEmployee(ctx context.Context, employee *models.Employee) error {
	return r.db.WithContext(ctx).Create(employee).Error
}

func (r *departmentRepository) UpdateDepartment(ctx context.Context, department *models.Department) error {
	return r.db.WithContext(ctx).
		Model(&models.Department{}).
		Where("id = ?", department.ID).
		Updates(map[string]any{
			"name":      department.Name,
			"parent_id": department.ParentID,
		}).Error
}

func (r *departmentRepository) DeleteDepartmentCascade(ctx context.Context, departmentID int) error {
	return r.db.WithContext(ctx).Delete(&models.Department{}, departmentID).Error
}
