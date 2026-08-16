// Package db opens and configures the GORM/Postgres connection used by every
// repository. Today it points at local Postgres (see devops/local); in AWS
// it points at the Aurora PostgreSQL cluster via the same DB_HOST/DB_* env
// vars (see aws.cfn.app.yml) - no code changes needed to switch.
package db

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"vidhya-service/src/config"
)

// Open connects to Postgres using the given configuration and configures the
// underlying connection pool.
func Open(cfg config.DatabaseConfig) (*gorm.DB, error) {
	gormDB, err := gorm.Open(postgres.Open(cfg.PostgresDSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("db.Open: connect to postgres: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("db.Open: get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConnections)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConnections)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("db.Open: ping postgres: %w", err)
	}

	return gormDB, nil
}
