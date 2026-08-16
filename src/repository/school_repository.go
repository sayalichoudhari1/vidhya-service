package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"vidhya-service/src/model"
)

// SchoolRepository persists the schools table (the multi-tenant root).
type SchoolRepository struct {
	db *gorm.DB
}

// NewSchoolRepository creates a SchoolRepository.
func NewSchoolRepository(db *gorm.DB) *SchoolRepository {
	return &SchoolRepository{db: db}
}

// Create inserts a new school.
func (r *SchoolRepository) Create(ctx context.Context, s *model.School) error {
	if err := r.db.WithContext(ctx).Create(s).Error; err != nil {
		return fmt.Errorf("SchoolRepository.Create: %w", err)
	}
	return nil
}

// GetByID fetches a school by ID.
func (r *SchoolRepository) GetByID(ctx context.Context, id string) (*model.School, error) {
	var s model.School
	if err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("SchoolRepository.GetByID: %w", err)
	}
	return &s, nil
}

// GetByCode fetches a school by its unique short code.
func (r *SchoolRepository) GetByCode(ctx context.Context, code string) (*model.School, error) {
	var s model.School
	if err := r.db.WithContext(ctx).First(&s, "code = ?", code).Error; err != nil {
		return nil, fmt.Errorf("SchoolRepository.GetByCode: %w", err)
	}
	return &s, nil
}

// ExistsByName reports whether an active school with this name (case
// insensitive) already exists - catches accidental duplicate onboarding of
// the same school under a different code.
func (r *SchoolRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.School{}).
		Where("lower(name) = lower(?) AND is_active = true", name).Count(&count).Error; err != nil {
		return false, fmt.Errorf("SchoolRepository.ExistsByName: %w", err)
	}
	return count > 0, nil
}

// List returns all schools (used by platform ADMIN).
func (r *SchoolRepository) List(ctx context.Context, limit, offset int) ([]model.School, error) {
	var schools []model.School
	if err := r.db.WithContext(ctx).Order("created_at desc").Limit(limit).Offset(offset).Find(&schools).Error; err != nil {
		return nil, fmt.Errorf("SchoolRepository.List: %w", err)
	}
	return schools, nil
}

// Update persists changes to an existing school.
func (r *SchoolRepository) Update(ctx context.Context, s *model.School) error {
	if err := r.db.WithContext(ctx).Model(&model.School{}).Where("id = ?", s.ID).Updates(s).Error; err != nil {
		return fmt.Errorf("SchoolRepository.Update: %w", err)
	}
	return nil
}

// Delete deactivates a school (soft-delete).
func (r *SchoolRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Model(&model.School{}).Where("id = ?", id).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("SchoolRepository.Delete: %w", err)
	}
	return nil
}
