// Package app wires together configuration, the database connection, and
// every repository/service into a single App object used by the HTTP router
// and by main.go.
package app

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"vidhya-service/src/cache"
	"vidhya-service/src/config"
	"vidhya-service/src/db"
	"vidhya-service/src/logger"
	"vidhya-service/src/messaging"
	"vidhya-service/src/repository"
	"vidhya-service/src/services/academic"
	"vidhya-service/src/services/attendance"
	"vidhya-service/src/services/auth"
	"vidhya-service/src/services/exam"
	"vidhya-service/src/services/school"
	"vidhya-service/src/services/student"
	"vidhya-service/src/services/teacher"
	"vidhya-service/src/util"
)

// App holds every dependency the HTTP layer needs.
type App struct {
	Config *config.Configuration
	DB     *gorm.DB

	JWTIssuer *util.JWTIssuer

	AuthService       *auth.Service
	SchoolService     *school.Service
	AcademicService   *academic.Service
	StudentService    *student.Service
	TeacherService    *teacher.Service
	AttendanceService *attendance.Service
	ExamService       *exam.Service

	shutdowns []func() error
}

// InitApp reads configuration, opens the database (and optionally
// Redis/Kafka), and constructs every repository and service.
func InitApp(ctx context.Context) (*App, error) {
	cfg, err := config.Read()
	if err != nil {
		return nil, fmt.Errorf("app.InitApp: read configuration: %w", err)
	}

	logger.Configure(cfg.Log.Level)
	logger.Log.Info("", "Configuration loaded: server=%s db=%s@%s:%d/%s",
		cfg.Server.Address, cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)

	a := &App{Config: cfg}

	gormDB, err := db.Open(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("app.InitApp: open database: %w", err)
	}
	a.DB = gormDB
	logger.Log.Info("", "Connected to Postgres at %s:%d/%s", cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)
	a.shutdowns = append(a.shutdowns, func() error {
		sqlDB, err := gormDB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	})

	// Redis and Kafka are optional and best-effort: if they're disabled (the
	// local default) or unreachable, the service logs a warning and keeps
	// running on Postgres alone. This is intentional for phase 1 - see
	// devops/local/.env.example and aws.cfn.app.yml for how to enable them.
	redisClient, err := cache.Connect(ctx, cfg.Redis)
	if err != nil {
		logger.Log.Warn("", "Redis unavailable, continuing without cache: %s", err)
	}
	if redisClient != nil {
		a.shutdowns = append(a.shutdowns, redisClient.Close)
	}

	kafkaProducer, err := messaging.Connect(ctx, cfg.Kafka, "vidhya-events")
	if err != nil {
		logger.Log.Warn("", "Kafka unavailable, continuing without messaging: %s", err)
	}
	if kafkaProducer != nil {
		a.shutdowns = append(a.shutdowns, kafkaProducer.Close)
	}

	// Repositories
	userRepo := repository.NewUserRepository(gormDB)
	schoolRepo := repository.NewSchoolRepository(gormDB)
	academicRepo := repository.NewAcademicRepository(gormDB)
	studentRepo := repository.NewStudentRepository(gormDB)
	teacherRepo := repository.NewTeacherRepository(gormDB)
	attendanceRepo := repository.NewAttendanceRepository(gormDB)
	examRepo := repository.NewExamRepository(gormDB)

	// Services
	a.JWTIssuer = util.NewJWTIssuer(cfg.JWT.Secret, cfg.JWT.Issuer, cfg.JWT.AccessTokenDuration)
	a.AuthService = auth.NewService(userRepo, a.JWTIssuer)
	a.SchoolService = school.NewService(schoolRepo)
	a.AcademicService = academic.NewService(academicRepo)
	a.StudentService = student.NewService(studentRepo, userRepo, a.AuthService)
	a.TeacherService = teacher.NewService(teacherRepo, userRepo, a.AuthService)
	a.AttendanceService = attendance.NewService(attendanceRepo)
	a.ExamService = exam.NewService(examRepo)

	if cfg.Seed.Enabled {
		if err := a.seedAdmin(ctx); err != nil {
			logger.Log.Warn("", "Failed to seed default admin user: %s", err)
		}
	}

	logger.Log.Info("", "Application initialized successfully")
	return a, nil
}

// Cleanup releases every resource opened during InitApp (DB, Redis, Kafka).
func (a *App) Cleanup() {
	for _, shutdown := range a.shutdowns {
		if err := shutdown(); err != nil {
			logger.Log.Error("", "SHUTDOWN_ERROR", "%s", err)
		}
	}
}
