package model

import "time"

// School is the top-level tenant. Vidhya supports multiple schools in one
// deployment; every academic/RBAC record is scoped by SchoolID.
type School struct {
	ID          string    `json:"id" gorm:"column:id;primaryKey"`
	Name        string    `json:"name" gorm:"column:name"`
	Code        string    `json:"code" gorm:"column:code"` // short unique code, e.g. "VID-BLR-01"
	Address     string    `json:"address,omitempty" gorm:"column:address"`
	PrincipalID *string   `json:"principalId,omitempty" gorm:"column:principal_id"`
	IsActive    bool      `json:"isActive" gorm:"column:is_active"`
	CreatedAt   time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

// TableName pins the GORM table name explicitly.
func (School) TableName() string { return "schools" }
