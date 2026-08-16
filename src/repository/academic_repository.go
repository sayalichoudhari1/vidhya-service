package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"vidhya-service/src/model"
)

// AcademicRepository persists classes, subjects, and the class<->subject
// (with teacher) assignments.
type AcademicRepository struct {
	db *gorm.DB
}

// NewAcademicRepository creates an AcademicRepository.
func NewAcademicRepository(db *gorm.DB) *AcademicRepository {
	return &AcademicRepository{db: db}
}

// --------------------------------------------------------------- classes ---

// CreateClass inserts a new class.
func (r *AcademicRepository) CreateClass(ctx context.Context, c *model.Class) error {
	if err := r.db.WithContext(ctx).Create(c).Error; err != nil {
		return fmt.Errorf("AcademicRepository.CreateClass: %w", err)
	}
	return nil
}

// GetClassByID fetches a class by ID.
func (r *AcademicRepository) GetClassByID(ctx context.Context, id string) (*model.Class, error) {
	var c model.Class
	if err := r.db.WithContext(ctx).First(&c, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("AcademicRepository.GetClassByID: %w", err)
	}
	return &c, nil
}

// ListClassesBySchool lists classes for a school.
func (r *AcademicRepository) ListClassesBySchool(ctx context.Context, schoolID string, limit, offset int) ([]model.Class, error) {
	var classes []model.Class
	if err := r.db.WithContext(ctx).Where("school_id = ?", schoolID).
		Order("name, section").Limit(limit).Offset(offset).Find(&classes).Error; err != nil {
		return nil, fmt.Errorf("AcademicRepository.ListClassesBySchool: %w", err)
	}
	return classes, nil
}

// ExistsClass reports whether a class with the same name/section/academic
// year already exists in the school, optionally excluding one ID (used when
// checking on update, so a class doesn't collide with itself).
func (r *AcademicRepository) ExistsClass(ctx context.Context, schoolID, name, section, academicYear string, excludeID string) (bool, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&model.Class{}).
		Where("school_id = ? AND lower(name) = lower(?) AND lower(section) = lower(?) AND academic_year = ?",
			schoolID, name, section, academicYear)
	if excludeID != "" {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return false, fmt.Errorf("AcademicRepository.ExistsClass: %w", err)
	}
	return count > 0, nil
}

// UpdateClass persists changes to an existing class.
func (r *AcademicRepository) UpdateClass(ctx context.Context, c *model.Class) error {
	if err := r.db.WithContext(ctx).Model(&model.Class{}).Where("id = ?", c.ID).Updates(c).Error; err != nil {
		return fmt.Errorf("AcademicRepository.UpdateClass: %w", err)
	}
	return nil
}

// DeleteClass removes a class by ID.
func (r *AcademicRepository) DeleteClass(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&model.Class{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("AcademicRepository.DeleteClass: %w", err)
	}
	return nil
}

// -------------------------------------------------------------- subjects ---

// CreateSubject inserts a new subject.
func (r *AcademicRepository) CreateSubject(ctx context.Context, s *model.Subject) error {
	if err := r.db.WithContext(ctx).Create(s).Error; err != nil {
		return fmt.Errorf("AcademicRepository.CreateSubject: %w", err)
	}
	return nil
}

// GetSubjectByID fetches a subject by ID.
func (r *AcademicRepository) GetSubjectByID(ctx context.Context, id string) (*model.Subject, error) {
	var s model.Subject
	if err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("AcademicRepository.GetSubjectByID: %w", err)
	}
	return &s, nil
}

// ListSubjectsBySchool lists subjects for a school.
func (r *AcademicRepository) ListSubjectsBySchool(ctx context.Context, schoolID string, limit, offset int) ([]model.Subject, error) {
	var subjects []model.Subject
	if err := r.db.WithContext(ctx).Where("school_id = ?", schoolID).
		Order("name").Limit(limit).Offset(offset).Find(&subjects).Error; err != nil {
		return nil, fmt.Errorf("AcademicRepository.ListSubjectsBySchool: %w", err)
	}
	return subjects, nil
}

// ExistsSubject reports whether a subject with the same name already
// exists in the school, optionally excluding one ID (for update checks).
func (r *AcademicRepository) ExistsSubject(ctx context.Context, schoolID, name string, excludeID string) (bool, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&model.Subject{}).
		Where("school_id = ? AND lower(name) = lower(?)", schoolID, name)
	if excludeID != "" {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return false, fmt.Errorf("AcademicRepository.ExistsSubject: %w", err)
	}
	return count > 0, nil
}

// UpdateSubject persists changes to an existing subject.
func (r *AcademicRepository) UpdateSubject(ctx context.Context, s *model.Subject) error {
	if err := r.db.WithContext(ctx).Model(&model.Subject{}).Where("id = ?", s.ID).Updates(s).Error; err != nil {
		return fmt.Errorf("AcademicRepository.UpdateSubject: %w", err)
	}
	return nil
}

// DeleteSubject removes a subject by ID.
func (r *AcademicRepository) DeleteSubject(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&model.Subject{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("AcademicRepository.DeleteSubject: %w", err)
	}
	return nil
}

// --------------------------------------------------------- class-subjects ---

// AssignSubjectToClass links a subject (and optional teacher) to a class.
func (r *AcademicRepository) AssignSubjectToClass(ctx context.Context, cs *model.ClassSubject) error {
	if err := r.db.WithContext(ctx).Create(cs).Error; err != nil {
		return fmt.Errorf("AcademicRepository.AssignSubjectToClass: %w", err)
	}
	return nil
}

// ExistsAssignment reports whether this subject is already assigned to this
// class (each subject may only be assigned to a class once - a second
// assignment would just be a duplicate row with a different teacher).
func (r *AcademicRepository) ExistsAssignment(ctx context.Context, classID, subjectID string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.ClassSubject{}).
		Where("class_id = ? AND subject_id = ?", classID, subjectID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("AcademicRepository.ExistsAssignment: %w", err)
	}
	return count > 0, nil
}

// ListSubjectsForClass lists the subject assignments for a class.
func (r *AcademicRepository) ListSubjectsForClass(ctx context.Context, classID string) ([]model.ClassSubject, error) {
	var rows []model.ClassSubject
	if err := r.db.WithContext(ctx).Where("class_id = ?", classID).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("AcademicRepository.ListSubjectsForClass: %w", err)
	}
	return rows, nil
}

// RemoveSubjectFromClass unlinks a subject from a class.
func (r *AcademicRepository) RemoveSubjectFromClass(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&model.ClassSubject{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("AcademicRepository.RemoveSubjectFromClass: %w", err)
	}
	return nil
}
