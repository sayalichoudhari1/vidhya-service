// Package teacher implements CRUD for teacher records (login + role-specific
// row), used mainly so principals can onboard staff and assign them as
// class teachers / subject teachers.
package teacher

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"vidhya-service/src/apperrors"
	"vidhya-service/src/model"
	"vidhya-service/src/repository"
	"vidhya-service/src/services/auth"
	"vidhya-service/src/util"
)

// CreateRequest is the payload for POST /teachers. firstName, lastName,
// email, phone, password, and employeeNumber are all required: phone is
// needed (together with name) to detect accidental duplicate onboarding,
// and employeeNumber is the school's own unique staff identifier.
type CreateRequest struct {
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	Password       string `json:"password"`
	EmployeeNumber string `json:"employeeNumber"`
}

// TeacherView is the response shape combining login + teacher fields.
type TeacherView struct {
	UserID         string `json:"userId"`
	SchoolID       string `json:"schoolId"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	Email          string `json:"email"`
	Phone          string `json:"phone,omitempty"`
	EmployeeNumber string `json:"employeeNumber,omitempty"`
	IsActive       bool   `json:"isActive"`
}

// Service implements teacher CRUD.
type Service struct {
	teacherRepo *repository.TeacherRepository
	userRepo    *repository.UserRepository
	authSvc     *auth.Service
}

// NewService creates a teacher Service.
func NewService(teacherRepo *repository.TeacherRepository, userRepo *repository.UserRepository, authSvc *auth.Service) *Service {
	return &Service{teacherRepo: teacherRepo, userRepo: userRepo, authSvc: authSvc}
}

// Create provisions a teacher's login and role-specific row. All
// pre-checks (required fields, duplicate detection) run before the login is
// created, so a rejected request never leaves behind an orphaned user row.
func (s *Service) Create(ctx context.Context, schoolID string, req CreateRequest) (*TeacherView, error) {
	if util.IsStrEmpty(req.EmployeeNumber) {
		return nil, apperrors.WithMessage(apperrors.ErrInvalidPayload, "employeeNumber is required")
	}

	exists, err := s.teacherRepo.ExistsByEmployeeNumber(ctx, schoolID, req.EmployeeNumber)
	if err != nil {
		return nil, fmt.Errorf("teacher.Create: check existing employee number: %w", err)
	}
	if exists {
		return nil, apperrors.WithMessage(apperrors.ErrConflict, "a teacher with this employee number already exists")
	}

	user, err := s.authSvc.CreateUser(ctx, auth.CreateUserRequest{
		SchoolID:  &schoolID,
		Role:      model.RoleTeacher,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Phone:     req.Phone,
		Password:  req.Password,
	})
	if err != nil {
		return nil, err
	}

	teacherRow := &model.Teacher{
		UserID:         user.ID,
		SchoolID:       schoolID,
		EmployeeNumber: req.EmployeeNumber,
	}
	if err := s.teacherRepo.Create(ctx, teacherRow); err != nil {
		return nil, fmt.Errorf("teacher.Create: %w", err)
	}

	return s.toView(user, teacherRow), nil
}

// Get fetches a teacher by user ID.
func (s *Service) Get(ctx context.Context, userID string) (*TeacherView, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WithMessage(apperrors.ErrNotFound, "teacher not found")
		}
		return nil, fmt.Errorf("teacher.Get: %w", err)
	}
	if user.Role != model.RoleTeacher {
		return nil, apperrors.WithMessage(apperrors.ErrNotFound, "teacher not found")
	}
	teacherRow, err := s.teacherRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("teacher.Get: %w", err)
	}
	return s.toView(user, teacherRow), nil
}

// List lists teachers in a school.
func (s *Service) List(ctx context.Context, schoolID string, limit, offset int) ([]TeacherView, error) {
	users, err := s.teacherRepo.ListBySchool(ctx, schoolID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("teacher.List: %w", err)
	}
	views := make([]TeacherView, 0, len(users))
	for _, u := range users {
		schoolID := ""
		if u.SchoolID != nil {
			schoolID = *u.SchoolID
		}
		views = append(views, TeacherView{
			UserID: u.ID, SchoolID: schoolID, FirstName: u.FirstName,
			LastName: u.LastName, Email: u.Email, Phone: u.Phone, IsActive: u.IsActive,
		})
	}
	return views, nil
}

// Delete deactivates a teacher.
func (s *Service) Delete(ctx context.Context, userID string) error {
	if _, err := s.Get(ctx, userID); err != nil {
		return err
	}
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("teacher.Delete: %w", err)
	}
	return nil
}

func (s *Service) toView(user *model.User, teacherRow *model.Teacher) *TeacherView {
	schoolID := ""
	if user.SchoolID != nil {
		schoolID = *user.SchoolID
	}
	return &TeacherView{
		UserID: user.ID, SchoolID: schoolID, FirstName: user.FirstName, LastName: user.LastName,
		Email: user.Email, Phone: user.Phone, EmployeeNumber: teacherRow.EmployeeNumber, IsActive: user.IsActive,
	}
}
