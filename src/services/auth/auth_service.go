// Package auth implements local login (email + password) and JWT issuance.
// There is no external identity provider in Vidhya - every user (admin,
// principal, teacher, student, parent) has a row in the users table with a
// bcrypt password hash.
package auth

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
	"vidhya-service/src/util"
)

// LoginRequest is the login API payload.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse is the login API response: a bearer token plus enough user
// context for the frontend to render the right dashboard immediately.
type LoginResponse struct {
	AccessToken string    `json:"accessToken"`
	ExpiresAt   time.Time `json:"expiresAt"`
	User        UserView  `json:"user"`
}

// UserView is the safe-to-return subset of a user's profile.
type UserView struct {
	ID        string     `json:"id"`
	SchoolID  string     `json:"schoolId,omitempty"`
	Role      model.Role `json:"role"`
	FirstName string     `json:"firstName"`
	LastName  string     `json:"lastName"`
	Email     string     `json:"email"`
}

// Service implements login and (self-service) password change.
type Service struct {
	userRepo *repository.UserRepository
	issuer   *util.JWTIssuer
}

// NewService creates an auth Service.
func NewService(userRepo *repository.UserRepository, issuer *util.JWTIssuer) *Service {
	return &Service{userRepo: userRepo, issuer: issuer}
}

// Login validates credentials and, on success, issues an access token.
func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	if util.IsStrEmpty(req.Email) || util.IsStrEmpty(req.Password) {
		return nil, apperrors.WithMessage(apperrors.ErrInvalidPayload, "email and password are required")
	}

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WithMessage(apperrors.ErrUnauthorized, "invalid email or password")
		}
		return nil, fmt.Errorf("auth.Login: lookup user: %w", err)
	}

	if !user.IsActive {
		return nil, apperrors.WithMessage(apperrors.ErrForbidden, "this account has been deactivated")
	}

	if !util.ComparePassword(user.PasswordHash, req.Password) {
		return nil, apperrors.WithMessage(apperrors.ErrUnauthorized, "invalid email or password")
	}

	schoolID := ""
	if user.SchoolID != nil {
		schoolID = *user.SchoolID
	}

	token, expiresAt, err := s.issuer.GenerateAccessToken(user.ID, schoolID, user.Email, user.Role)
	if err != nil {
		return nil, fmt.Errorf("auth.Login: generate token: %w", err)
	}

	return &LoginResponse{
		AccessToken: token,
		ExpiresAt:   expiresAt,
		User: UserView{
			ID:        user.ID,
			SchoolID:  schoolID,
			Role:      user.Role,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
		},
	}, nil
}

// Me returns the profile for the currently authenticated user.
func (s *Service) Me(ctx context.Context, userID string) (*UserView, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("auth.Me: %w", err)
	}
	schoolID := ""
	if user.SchoolID != nil {
		schoolID = *user.SchoolID
	}
	return &UserView{
		ID: user.ID, SchoolID: schoolID, Role: user.Role,
		FirstName: user.FirstName, LastName: user.LastName, Email: user.Email,
	}, nil
}

// CreateUserRequest is used internally by other services (school, student,
// teacher onboarding) to create a login for a new principal/teacher/etc.
type CreateUserRequest struct {
	SchoolID  *string
	Role      model.Role
	FirstName string
	LastName  string
	Email     string
	Phone     string
	Password  string
}

// CreateUser creates a new login for any role. Used by the school service
// (principal), and by the student/teacher services (to create the
// underlying login before writing the role-specific enrollment row).
func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) (*model.User, error) {
	if util.IsStrEmpty(req.FirstName) || util.IsStrEmpty(req.LastName) {
		return nil, apperrors.WithMessage(apperrors.ErrInvalidPayload, "firstName and lastName are required")
	}
	if !util.IsValidEmail(req.Email) {
		return nil, apperrors.WithMessage(apperrors.ErrInvalidPayload, "a valid email is required")
	}
	if !req.Role.IsValid() {
		return nil, apperrors.WithMessage(apperrors.ErrInvalidPayload, "role is invalid")
	}
	if util.IsStrEmpty(req.Password) {
		return nil, apperrors.WithMessage(apperrors.ErrInvalidPayload, "password is required")
	}
	if !util.IsValidPhone(req.Phone) {
		return nil, apperrors.WithMessage(apperrors.ErrInvalidPayload, "a valid phone number is required")
	}
	phone := util.NormalizePhone(req.Phone)

	emailExists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("auth.CreateUser: check existing email: %w", err)
	}
	if emailExists {
		return nil, apperrors.WithMessage(apperrors.ErrConflict, "a user with this email already exists")
	}

	dupExists, err := s.userRepo.ExistsByNamePhone(ctx, req.SchoolID, req.FirstName, req.LastName, phone)
	if err != nil {
		return nil, fmt.Errorf("auth.CreateUser: check existing name/phone: %w", err)
	}
	if dupExists {
		return nil, apperrors.WithMessage(apperrors.ErrConflict,
			"a user with the same first name, last name, and phone number already exists")
	}

	hash, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("auth.CreateUser: hash password: %w", err)
	}

	user := &model.User{
		ID:           uuid.NewString(),
		SchoolID:     req.SchoolID,
		Role:         req.Role,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		Phone:        phone,
		PasswordHash: hash,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("auth.CreateUser: %w", err)
	}
	return user, nil
}
