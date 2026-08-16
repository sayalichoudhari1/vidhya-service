package app

import (
	"context"
	"errors"
	"fmt"

	"vidhya-service/src/apperrors"
	"vidhya-service/src/logger"
	"vidhya-service/src/model"
	"vidhya-service/src/services/auth"
)

// seedAdmin creates a single platform ADMIN user on a fresh database so the
// API is immediately usable after `docker compose up` + `go run`, without
// any manual SQL. It is idempotent: if a user with SEED_ADMIN_EMAIL already
// exists, nothing happens. Controlled by SEED_ADMIN_ENABLED (default true).
func (a *App) seedAdmin(ctx context.Context) error {
	_, err := a.AuthService.CreateUser(ctx, auth.CreateUserRequest{
		Role:      model.RoleAdmin,
		FirstName: "Vidhya",
		LastName:  "Admin",
		Email:     a.Config.Seed.AdminEmail,
		Phone:     a.Config.Seed.AdminPhone,
		Password:  a.Config.Seed.AdminPassword,
	})
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) && appErr.Code == apperrors.ErrConflict.Code {
			logger.Log.Debug("", "Seed admin user already exists (%s) - skipping", a.Config.Seed.AdminEmail)
			return nil
		}
		return fmt.Errorf("app.seedAdmin: %w", err)
	}

	logger.Log.Info("", "Seeded default admin user: email=%s password=%s (change this immediately outside local dev!)",
		a.Config.Seed.AdminEmail, a.Config.Seed.AdminPassword)
	return nil
}
