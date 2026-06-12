package database

import (
	"3za-digital/pkg/logger"
	"3za-digital/utils"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const postgresTimeZone = "Asia/Jakarta"

func ConnDb() (db *gorm.DB, sqlDB *sql.DB, err error) {
	dsn := PostgresDSN()

	logger.WriteLog(logger.LogLevelDebug, "ConnDb; Initialize db connection...")

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		logger.WriteLog(logger.LogLevelError, fmt.Sprintf("ConnDb; %s Error: %s", dsn, err.Error()))
		return
	}

	maxIdle := 10
	maxIdleTime := 5 * time.Minute
	maxConn := 100
	maxLifeTime := time.Hour

	sqlDB, err = db.DB()
	if err != nil {
		logger.WriteLog(logger.LogLevelError, fmt.Sprintf("ConnDb.sqlDB; %s Error: %s", dsn, err.Error()))
		return
	}

	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxIdleTime(maxIdleTime)
	sqlDB.SetMaxOpenConns(maxConn)
	sqlDB.SetConnMaxLifetime(maxLifeTime)

	db.Debug()

	return
}

func PostgresDSN() string {
	dsn := strings.TrimSpace(utils.GetEnv("DATABASE_URL", ""))
	if dsn == "" {
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
			utils.GetEnv("DB_HOST", ""),
			utils.GetEnv("DB_PORT", ""),
			utils.GetEnv("DB_USERNAME", ""),
			utils.GetEnv("DB_PASS", ""),
			utils.GetEnv("DB_NAME", ""),
			utils.GetEnv("DB_SSLMODE", "disable"),
			postgresTimeZone)
	}

	return withPostgresTimeZone(dsn)
}

func withPostgresTimeZone(dsn string) string {
	lowerDSN := strings.ToLower(dsn)
	if strings.Contains(lowerDSN, "timezone=") || strings.Contains(lowerDSN, "time zone=") {
		return dsn
	}

	if strings.Contains(dsn, "://") {
		separator := "?"
		if strings.Contains(dsn, "?") {
			separator = "&"
		}
		return dsn + separator + "TimeZone=" + url.QueryEscape(postgresTimeZone)
	}

	return dsn + " TimeZone=" + postgresTimeZone
}
