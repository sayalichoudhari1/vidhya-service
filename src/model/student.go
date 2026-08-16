package model

import "time"

// Student holds the academic-record fields for a user with Role=RoleStudent.
// UserID is a 1:1 reference back to the users table (name/email/login live
// there); this table only carries student-specific enrollment data.
type Student struct {
	UserID          string     `json:"userId" gorm:"column:user_id;primaryKey"`
	SchoolID        string     `json:"schoolId" gorm:"column:school_id"`
	AdmissionNumber string     `json:"admissionNumber" gorm:"column:admission_number"`
	RollNumber      string     `json:"rollNumber,omitempty" gorm:"column:roll_number"`
	ClassID         *string    `json:"classId,omitempty" gorm:"column:class_id"`
	ParentID        *string    `json:"parentId,omitempty" gorm:"column:parent_id"`
	DateOfBirth     *time.Time `json:"dateOfBirth,omitempty" gorm:"column:date_of_birth"`
	Gender          string     `json:"gender,omitempty" gorm:"column:gender"`
	Address         string     `json:"address,omitempty" gorm:"column:address"`
	BloodGroup      string     `json:"bloodGroup,omitempty" gorm:"column:blood_group"`
	CreatedAt       time.Time  `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt       time.Time  `json:"updatedAt" gorm:"column:updated_at"`
}

// TableName pins the GORM table name explicitly.
func (Student) TableName() string { return "students" }

// StudentWithUser is a read-friendly projection joining the student and the
// underlying user row, used for API responses.
type StudentWithUser struct {
	Student
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Phone     string `json:"phone,omitempty"`
	IsActive  bool   `json:"isActive"`
}

// Teacher holds the fields for a user with Role=RoleTeacher.
type Teacher struct {
	UserID         string    `json:"userId" gorm:"column:user_id;primaryKey"`
	SchoolID       string    `json:"schoolId" gorm:"column:school_id"`
	EmployeeNumber string    `json:"employeeNumber,omitempty" gorm:"column:employee_number"`
	CreatedAt      time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt      time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

// TableName pins the GORM table name explicitly.
func (Teacher) TableName() string { return "teachers" }
