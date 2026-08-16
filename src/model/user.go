package model

import "time"

// Role is a platform-wide RBAC role. Every user has exactly one role in
// phase 1; permissions are derived purely from this field (no separate
// permissions table yet - see docs for the phase-2 fine-grained plan).
type Role string

// Supported roles for phase 1.
const (
	RoleAdmin     Role = "ADMIN"     // full access across all schools
	RolePrincipal Role = "PRINCIPAL" // full access within their own school
	RoleTeacher   Role = "TEACHER"   // manage attendance/marks for classes they teach
	RoleStudent   Role = "STUDENT"   // read-only access to their own records
	RoleParent    Role = "PARENT"    // read-only access to their children's records
)

// IsValid reports whether r is one of the known roles.
func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RolePrincipal, RoleTeacher, RoleStudent, RoleParent:
		return true
	default:
		return false
	}
}

// User is any authenticated principal in the system: admin, principal,
// teacher, student, or parent. Role-specific data (e.g. which class a
// student is in) lives in the Student/Teacher tables, joined by UserID.
type User struct {
	ID           string     `json:"id" gorm:"column:id;primaryKey"`
	SchoolID     *string    `json:"schoolId,omitempty" gorm:"column:school_id"` // null for platform ADMIN
	Role         Role       `json:"role" gorm:"column:role"`
	FirstName    string     `json:"firstName" gorm:"column:first_name"`
	LastName     string     `json:"lastName" gorm:"column:last_name"`
	Email        string     `json:"email" gorm:"column:email"`
	Phone        string     `json:"phone,omitempty" gorm:"column:phone"`
	PasswordHash string     `json:"-" gorm:"column:password_hash"`
	IsActive     bool       `json:"isActive" gorm:"column:is_active"`
	CreatedAt    time.Time  `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt    time.Time  `json:"updatedAt" gorm:"column:updated_at"`
	DeletedAt    *time.Time `json:"-" gorm:"column:deleted_at"`
}

// TableName pins the GORM table name explicitly.
func (User) TableName() string { return "users" }

// FullName returns the user's display name.
func (u User) FullName() string {
	return u.FirstName + " " + u.LastName
}
