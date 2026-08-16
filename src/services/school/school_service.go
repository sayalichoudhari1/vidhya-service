// Package school implements CRUD for the schools table - the multi-tenant
// root entity. Only platform ADMIN can create/list/delete schools; a
// PRINCIPAL can view/update their own school.
package school

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"vidhya-service/src/apperrors"
	"vidhya-service/src/model"
	"vidhya-service/src/repository"
	"vidhya-service/src/util"
)

// CreateRequest is the payload for creating a school.
type CreateRequest struct {
	Name    string `json:"name"`
	Code    string `json:"code"`
	Address string `json:"address,omitempty"`
}

// UpdateRequest is the payload for updating a school.
type UpdateRequest struct {
	Name    *string `json:"name,omitempty"`
	Address *string `json:"address,omitempty"`
}

// Service implements school CRUD.
type Service struct {
	schoolRepo *repository.SchoolRepository
}

// NewService creates a school Service.
func NewService(schoolRepo *repository.SchoolRepository) *Service {
	return &Service{schoolRepo: schoolRepo}
}

// Create inserts a new school.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*model.School, error) {
	if util.IsStrEmpty(req.Name) || util.IsStrEmpty(req.Code) {
		return nil, apperrors.WithMessage(apperrors.ErrInvalidPayload, "name and code are required")
	}

	if _, err := s.schoolRepo.GetByCode(ctx, req.Code); err == nil {
		return nil, apperrors.WithMessage(apperrors.ErrConflict, "a school with this code already exists")
	}

	nameExists, err := s.schoolRepo.ExistsByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("school.Create: check existing name: %w", err)
	}
	if nameExists {
		return nil, apperrors.WithMessage(apperrors.ErrConflict, "a school with this name already exists")
	}

	school := &model.School{
		ID:       uuid.NewString(),
		Name:     req.Name,
		Code:     req.Code,
		Address:  req.Address,
		IsActive: true,
	}
	if err := s.schoolRepo.Create(ctx, school); err != nil {
		return nil, fmt.Errorf("school.Create: %w", err)
	}
	return school, nil
}

// Get fetches a school by ID.
func (s *Service) Get(ctx context.Context, id string) (*model.School, error) {
	school, err := s.schoolRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("school.Get: %w", err)
	}
	return school, nil
}

// List returns all schools.
func (s *Service) List(ctx context.Context, limit, offset int) ([]model.School, error) {
	schools, err := s.schoolRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("school.List: %w", err)
	}
	return schools, nil
}

// Update applies a partial update to a school.
func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*model.School, error) {
	school, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		school.Name = *req.Name
	}
	if req.Address != nil {
		school.Address = *req.Address
	}
	if err := s.schoolRepo.Update(ctx, school); err != nil {
		return nil, fmt.Errorf("school.Update: %w", err)
	}
	return school, nil
}

// Delete deactivates a school.
func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}
	if err := s.schoolRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("school.Delete: %w", err)
	}
	return nil
}

// SetPrincipal links a school to its principal user.
func (s *Service) SetPrincipal(ctx context.Context, schoolID, principalUserID string) error {
	school, err := s.Get(ctx, schoolID)
	if err != nil {
		return err
	}
	school.PrincipalID = &principalUserID
	if err := s.schoolRepo.Update(ctx, school); err != nil {
		return fmt.Errorf("school.SetPrincipal: %w", err)
	}
	return nil
}
