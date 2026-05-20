package models

import "time"

type Department struct {
	ID        int          `json:"id" gorm:"primaryKey"`
	Name      string       `json:"name" gorm:"size:200;not null"`
	ParentID  *int         `json:"parent_id" gorm:"index"`
	Parent    *Department  `json:"-" gorm:"foreignKey:ParentID"`
	Children  []Department `json:"-" gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE;"`
	Employees []Employee   `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	CreatedAt time.Time    `json:"created_at"`
}

type Employee struct {
	ID           int        `json:"id" gorm:"primaryKey"`
	DepartmentID int        `json:"department_id" gorm:"not null;index"`
	FullName     string     `json:"full_name" gorm:"size:200;not null"`
	Position     string     `json:"position" gorm:"size:200;not null"`
	HiredAt      *time.Time `json:"hired_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

type DepartmentNode struct {
	Department Department       `json:"department"`
	Employees  []Employee       `json:"employees,omitempty"`
	Children   []DepartmentNode `json:"children"`
}
