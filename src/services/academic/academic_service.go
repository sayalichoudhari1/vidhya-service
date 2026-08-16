// Package academic implements CRUD for classes, subjects, and the
// class<->subject (with teacher) assignments that attendance/exams hang off.
package academic

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

// Service implements class/subject/class-subject CRUD.
type Service struct {
	repo *repository.AcademicRepository
}

// NewService creates an academic Service.
func NewService(repo *repository.AcademicRepository) *Service {
	return &Service{repo: repo}
}

// --------------------------------------------------------------- classes ---

// CreateClassRequest is the payload for creating a class.
type CreateClassRequest struct {
	Name           string  `json:"name"`
	Section        string  `json:"section"`
	AcademicYear   string  `json:"academicYear"`
	ClassTeacherID *string `json:"classTeacherId,omitempty"`
}

// CreateClass inserts a new class for a school.
func (s *Service) CreateClass(ctx context.Context, schoolID string, req CreateClassRequest) (*model.Class, error) {
	if util.IsStrEmpty(req.Name) || util.IsStrEmpty(req.AcademicYear) {
		return nil, apperrors.WithMessage(apperrors.ErrInvalidPayload, "name and academicYear are required")
	}
	exists, err := s.repo.ExistsClass(ctx, schoolID, req.Name, req.Section, req.AcademicYear, "")
	if err != nil {
		return nil, fmt.Errorf("academic.CreateClass: check existing: %w", err)
	}
	if exists {
		return nil, apperrors.WithMessage(apperrors.ErrConflict,
			"a class with this name, section, and academic year already exists")
	}
	class := &model.Class{
		ID:             uuid.NewString(),
		SchoolID:       schoolID,
		Name:           req.Name,
		Section:        req.Section,
		AcademicYear:   req.AcademicYear,
		ClassTeacherID: req.ClassTeacherID,
	}
	if err := s.repo.CreateClass(ctx, class); err != nil {
		return nil, fmt.Errorf("academic.CreateClass: %w", err)
	}
	return class, nil
}

// GetClass fetches a class by ID, scoped to the caller's school.
func (s *Service) GetClass(ctx context.Context, schoolID, id string) (*model.Class, error) {
	class, err := s.repo.GetClassByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("academic.GetClass: %w", err)
	}
	if class.SchoolID != schoolID {
		return nil, apperrors.ErrNotFound
	}
	return class, nil
}

// ListClasses lists classes for a school.
func (s *Service) ListClasses(ctx context.Context, schoolID string, limit, offset int) ([]model.Class, error) {
	classes, err := s.repo.ListClassesBySchool(ctx, schoolID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("academic.ListClasses: %w", err)
	}
	return classes, nil
}

// UpdateClassRequest is the payload for updating a class.
type UpdateClassRequest struct {
	Name           *string `json:"name,omitempty"`
	Section        *string `json:"section,omitempty"`
	ClassTeacherID *string `json:"classTeacherId,omitempty"`
}

// UpdateClass applies a partial update to a class.
func (s *Service) UpdateClass(ctx context.Context, schoolID, id string, req UpdateClassRequest) (*model.Class, error) {
	class, err := s.GetClass(ctx, schoolID, id)
	if err != nil {
		return nil, err
	}
	name, section := class.Name, class.Section
	if req.Name != nil {
		name = *req.Name
	}
	if req.Section != nil {
		section = *req.Section
	}
	if req.Name != nil || req.Section != nil {
		exists, err := s.repo.ExistsClass(ctx, schoolID, name, section, class.AcademicYear, id)
		if err != nil {
			return nil, fmt.Errorf("academic.UpdateClass: check existing: %w", err)
		}
		if exists {
			return nil, apperrors.WithMessage(apperrors.ErrConflict,
				"a class with this name, section, and academic year already exists")
		}
	}
	class.Name = name
	class.Section = section
	if req.ClassTeacherID != nil {
		class.ClassTeacherID = req.ClassTeacherID
	}
	if err := s.repo.UpdateClass(ctx, class); err != nil {
		return nil, fmt.Errorf("academic.UpdateClass: %w", err)
	}
	return class, nil
}

// DeleteClass removes a class by ID.
func (s *Service) DeleteClass(ctx context.Context, schoolID, id string) error {
	if _, err := s.GetClass(ctx, schoolID, id); err != nil {
		return err
	}
	if err := s.repo.DeleteClass(ctx, id); err != nil {
		return fmt.Errorf("academic.DeleteClass: %w", err)
	}
	return nil
}

// -------------------------------------------------------------- subjects ---

// CreateSubjectRequest is the payload for creating a subject.
type CreateSubjectRequest struct {
	Name string `json:"name"`
	Code string `json:"code,omitempty"`
}

// CreateSubject inserts a new subject for a school.
func (s *Service) CreateSubject(ctx context.Context, schoolID string, req CreateSubjectRequest) (*model.Subject, error) {
	if util.IsStrEmpty(req.Name) {
		return nil, apperrors.WithMessage(apperrors.ErrInvalidPayload, "name is required")
	}
	exists, err := s.repo.ExistsSubject(ctx, schoolID, req.Name, "")
	if err != nil {
		return nil, fmt.Errorf("academic.CreateSubject: check existing: %w", err)
	}
	if exists {
		return nil, apperrors.WithMessage(apperrors.ErrConflict, "a subject with this name already exists")
	}
	subject := &model.Subject{
		ID:       uuid.NewString(),
		SchoolID: schoolID,
		Name:     req.Name,
		Code:     req.Code,
	}
	if err := s.repo.CreateSubject(ctx, subject); err != nil {
		return nil, fmt.Errorf("academic.CreateSubject: %w", err)
	}
	return subject, nil
}

// ListSubjects lists subjects for a school.
func (s *Service) ListSubjects(ctx context.Context, schoolID string, limit, offset int) ([]model.Subject, error) {
	subjects, err := s.repo.ListSubjectsBySchool(ctx, schoolID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("academic.ListSubjects: %w", err)
	}
	return subjects, nil
}

// GetSubject fetches a subject by ID, scoped to the caller's school.
func (s *Service) GetSubject(ctx context.Context, schoolID, id string) (*model.Subject, error) {
	subject, err := s.repo.GetSubjectByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("academic.GetSubject: %w", err)
	}
	if subject.SchoolID != schoolID {
		return nil, apperrors.ErrNotFound
	}
	return subject, nil
}

// UpdateSubjectRequest is the payload for updating a subject.
type UpdateSubjectRequest struct {
	Name *string `json:"name,omitempty"`
	Code *string `json:"code,omitempty"`
}

// UpdateSubject applies a partial update to a subject.
func (s *Service) UpdateSubject(ctx context.Context, schoolID, id string, req UpdateSubjectRequest) (*model.Subject, error) {
	subject, err := s.GetSubject(ctx, schoolID, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		exists, err := s.repo.ExistsSubject(ctx, schoolID, *req.Name, id)
		if err != nil {
			return nil, fmt.Errorf("academic.UpdateSubject: check existing: %w", err)
		}
		if exists {
			return nil, apperrors.WithMessage(apperrors.ErrConflict, "a subject with this name already exists")
		}
		subject.Name = *req.Name
	}
	if req.Code != nil {
		subject.Code = *req.Code
	}
	if err := s.repo.UpdateSubject(ctx, subject); err != nil {
		return nil, fmt.Errorf("academic.UpdateSubject: %w", err)
	}
	return subject, nil
}

// DeleteSubject removes a subject by ID.
func (s *Service) DeleteSubject(ctx context.Context, schoolID, id string) error {
	if _, err := s.GetSubject(ctx, schoolID, id); err != nil {
		return err
	}
	if err := s.repo.DeleteSubject(ctx, id); err != nil {
		return fmt.Errorf("academic.DeleteSubject: %w", err)
	}
	return nil
}

// --------------------------------------------------------- class-subjects ---

// AssignSubjectRequest is the payload for assigning a subject to a class.
type AssignSubjectRequest struct {
	SubjectID string  `json:"subjectId"`
	TeacherID *string `json:"teacherId,omitempty"`
}

// AssignSubjectToClass links a subject (and optional teacher) to a class.
func (s *Service) AssignSubjectToClass(ctx context.Context, classID string, req AssignSubjectRequest) (*model.ClassSubject, error) {
	if util.IsStrEmpty(req.SubjectID) {
		return nil, apperrors.WithMessage(apperrors.ErrInvalidPayload, "subjectId is required")
	}
	exists, err := s.repo.ExistsAssignment(ctx, classID, req.SubjectID)
	if err != nil {
		return nil, fmt.Errorf("academic.AssignSubjectToClass: check existing: %w", err)
	}
	if exists {
		return nil, apperrors.WithMessage(apperrors.ErrConflict, "this subject is already assigned to this class")
	}
	cs := &model.ClassSubject{
		ID:        uuid.NewString(),
		ClassID:   classID,
		SubjectID: req.SubjectID,
		TeacherID: req.TeacherID,
	}
	if err := s.repo.AssignSubjectToClass(ctx, cs); err != nil {
		return nil, fmt.Errorf("academic.AssignSubjectToClass: %w", err)
	}
	return cs, nil
}

// ListSubjectsForClass lists subject assignments for a class.
func (s *Service) ListSubjectsForClass(ctx context.Context, classID string) ([]model.ClassSubject, error) {
	rows, err := s.repo.ListSubjectsForClass(ctx, classID)
	if err != nil {
		return nil, fmt.Errorf("academic.ListSubjectsForClass: %w", err)
	}
	return rows, nil
}

// RemoveSubjectFromClass unlinks a subject from a class.
func (s *Service) RemoveSubjectFromClass(ctx context.Context, id string) error {
	if err := s.repo.RemoveSubjectFromClass(ctx, id); err != nil {
		return fmt.Errorf("academic.RemoveSubjectFromClass: %w", err)
	}
	return nil
}
