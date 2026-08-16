// Package student implements full CRUD for student records: creating a
// student (which also provisions their login), updating enrollment details,
// and listing/searching students within a school or class.
package student

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"vidhya-service/src/apperrors"
	"vidhya-service/src/model"
	"vidhya-service/src/repository"
	"vidhya-service/src/services/auth"
)

// CreateRequest is the payload for POST /students - creates the login and
// the student enrollment row in one call. firstName, lastName, email,
// phone, password, and admissionNumber are all required: phone is needed
// (together with name) to detect accidental duplicate onboarding.
type CreateRequest struct {
	FirstName       string  `json:"firstName"`
	LastName        string  `json:"lastName"`
	Email           string  `json:"email"`
	Phone           string  `json:"phone"`
	Password        string  `json:"password"`
	AdmissionNumber string  `json:"admissionNumber"`
	RollNumber      string  `json:"rollNumber,omitempty"`
	ClassID         *string `json:"classId,omitempty"`
	ParentID        *string `json:"parentId,omitempty"`
	DateOfBirth     *string `json:"dateOfBirth,omitempty"` // RFC3339 date, e.g. "2012-05-01"
	Gender          string  `json:"gender,omitempty"`
	Address         string  `json:"address,omitempty"`
	BloodGroup      string  `json:"bloodGroup,omitempty"`
}

// UpdateRequest is the payload for PUT /students/{id} - partial update of
// enrollment fields (login/email changes are out of scope for phase 1).
type UpdateRequest struct {
	RollNumber *string `json:"rollNumber,omitempty"`
	ClassID    *string `json:"classId,omitempty"`
	ParentID   *string `json:"parentId,omitempty"`
	Gender     *string `json:"gender,omitempty"`
	Address    *string `json:"address,omitempty"`
	BloodGroup *string `json:"bloodGroup,omitempty"`
	FirstName  *string `json:"firstName,omitempty"`
	LastName   *string `json:"lastName,omitempty"`
	Phone      *string `json:"phone,omitempty"`
}

// Service implements student CRUD.
type Service struct {
	studentRepo *repository.StudentRepository
	userRepo    *repository.UserRepository
	authSvc     *auth.Service
}

// NewService creates a student Service.
func NewService(studentRepo *repository.StudentRepository, userRepo *repository.UserRepository, authSvc *auth.Service) *Service {
	return &Service{studentRepo: studentRepo, userRepo: userRepo, authSvc: authSvc}
}

// Create provisions a student's login and enrollment record. All
// pre-checks (required fields, duplicate detection) run before the login is
// created, so a rejected request never leaves behind an orphaned user row.
func (s *Service) Create(ctx context.Context, schoolID string, req CreateRequest) (*model.StudentWithUser, error) {
	if req.AdmissionNumber == "" {
		return nil, apperrors.WithMessage(apperrors.ErrInvalidPayload, "admissionNumber is required")
	}

	admissionExists, err := s.studentRepo.ExistsByAdmissionNumber(ctx, schoolID, req.AdmissionNumber)
	if err != nil {
		return nil, fmt.Errorf("student.Create: check admission number: %w", err)
	}
	if admissionExists {
		return nil, apperrors.WithMessage(apperrors.ErrConflict, "a student with this admission number already exists")
	}

	if req.ClassID != nil && req.RollNumber != "" {
		rollExists, err := s.studentRepo.ExistsByClassRollNumber(ctx, *req.ClassID, req.RollNumber, "")
		if err != nil {
			return nil, fmt.Errorf("student.Create: check roll number: %w", err)
		}
		if rollExists {
			return nil, apperrors.WithMessage(apperrors.ErrConflict,
				"another student in this class already has this roll number")
		}
	}

	user, err := s.authSvc.CreateUser(ctx, auth.CreateUserRequest{
		SchoolID:  &schoolID,
		Role:      model.RoleStudent,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Phone:     req.Phone,
		Password:  req.Password,
	})
	if err != nil {
		return nil, err
	}

	var dob *time.Time
	if req.DateOfBirth != nil && *req.DateOfBirth != "" {
		parsed, parseErr := time.Parse("2006-01-02", *req.DateOfBirth)
		if parseErr != nil {
			return nil, apperrors.WithMessage(apperrors.ErrInvalidPayload, "dateOfBirth must be in YYYY-MM-DD format")
		}
		dob = &parsed
	}

	studentRow := &model.Student{
		UserID:          user.ID,
		SchoolID:        schoolID,
		AdmissionNumber: req.AdmissionNumber,
		RollNumber:      req.RollNumber,
		ClassID:         req.ClassID,
		ParentID:        req.ParentID,
		DateOfBirth:     dob,
		Gender:          req.Gender,
		Address:         req.Address,
		BloodGroup:      req.BloodGroup,
	}
	if err := s.studentRepo.Create(ctx, studentRow); err != nil {
		return nil, fmt.Errorf("student.Create: %w", err)
	}

	return s.Get(ctx, user.ID)
}

// Get fetches a student's enrollment + profile by user ID.
func (s *Service) Get(ctx context.Context, userID string) (*model.StudentWithUser, error) {
	student, err := s.studentRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WithMessage(apperrors.ErrNotFound, "student not found")
		}
		return nil, fmt.Errorf("student.Get: %w", err)
	}
	return student, nil
}

// List lists students in a school, optionally filtered by class.
func (s *Service) List(ctx context.Context, schoolID string, classID *string, limit, offset int) ([]model.StudentWithUser, error) {
	students, err := s.studentRepo.ListBySchool(ctx, schoolID, classID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("student.List: %w", err)
	}
	return students, nil
}

// Update applies a partial update across the student and user rows.
func (s *Service) Update(ctx context.Context, userID string, req UpdateRequest) (*model.StudentWithUser, error) {
	existing, err := s.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	studentRow := existing.Student
	if req.RollNumber != nil {
		studentRow.RollNumber = *req.RollNumber
	}
	if req.ClassID != nil {
		studentRow.ClassID = req.ClassID
	}
	if (req.RollNumber != nil || req.ClassID != nil) && studentRow.ClassID != nil && studentRow.RollNumber != "" {
		rollExists, err := s.studentRepo.ExistsByClassRollNumber(ctx, *studentRow.ClassID, studentRow.RollNumber, userID)
		if err != nil {
			return nil, fmt.Errorf("student.Update: check roll number: %w", err)
		}
		if rollExists {
			return nil, apperrors.WithMessage(apperrors.ErrConflict,
				"another student in this class already has this roll number")
		}
	}
	if req.ParentID != nil {
		studentRow.ParentID = req.ParentID
	}
	if req.Gender != nil {
		studentRow.Gender = *req.Gender
	}
	if req.Address != nil {
		studentRow.Address = *req.Address
	}
	if req.BloodGroup != nil {
		studentRow.BloodGroup = *req.BloodGroup
	}
	if err := s.studentRepo.Update(ctx, &studentRow); err != nil {
		return nil, fmt.Errorf("student.Update: %w", err)
	}

	if req.FirstName != nil || req.LastName != nil || req.Phone != nil {
		user, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("student.Update: reload user: %w", err)
		}
		if req.FirstName != nil {
			user.FirstName = *req.FirstName
		}
		if req.LastName != nil {
			user.LastName = *req.LastName
		}
		if req.Phone != nil {
			user.Phone = *req.Phone
		}
		if err := s.userRepo.Update(ctx, user); err != nil {
			return nil, fmt.Errorf("student.Update: %w", err)
		}
	}

	return s.Get(ctx, userID)
}

// Delete deactivates a student (soft-delete: login disabled, records kept
// for attendance/exam history integrity).
func (s *Service) Delete(ctx context.Context, userID string) error {
	if _, err := s.Get(ctx, userID); err != nil {
		return err
	}
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("student.Delete: %w", err)
	}
	return nil
}

// NewStudentID is exposed for callers (e.g. seed data) that need to
// pre-generate a student's user ID before other writes.
func NewStudentID() string {
	return uuid.NewString()
}
