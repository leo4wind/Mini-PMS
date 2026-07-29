package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"minipms/internal/config"
	"minipms/internal/pkg/applog"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// sqlLogger 将 GORM SQL 以 [DEBUG] 输出
type sqlLogger struct {
	slow time.Duration
}

func (l sqlLogger) LogMode(level logger.LogLevel) logger.Interface { return l }

func (l sqlLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	applog.Debug("gorm: "+msg, data...)
}

func (l sqlLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	applog.Info("gorm-warn: "+msg, data...)
}

func (l sqlLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	applog.Info("gorm-error: "+msg, data...)
}

func (l sqlLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if applog.LevelName() != "debug" {
		// info 级别不打 SQL；错误仍提示
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			elapsed := time.Since(begin)
			sql, rows := fc()
			applog.Info("SQL err= %v | %s | rows=%d | %s", err, applog.FormatDuration(elapsed), rows, sql)
		}
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()
	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		applog.Debug("SQL err= %v | %s | rows=%d | %s", err, applog.FormatDuration(elapsed), rows, sql)
	case l.slow > 0 && elapsed > l.slow:
		applog.Debug("SQL SLOW | %s | rows=%d | %s", applog.FormatDuration(elapsed), rows, sql)
	default:
		applog.Debug("SQL | %s | rows=%d | %s", applog.FormatDuration(elapsed), rows, sql)
	}
}

func Connect(cfg *config.Config) (*gorm.DB, error) {
	slow := 500 * time.Millisecond
	db, err := gorm.Open(mysql.Open(cfg.MySQL.DSN), &gorm.Config{
		Logger:      sqlLogger{slow: slow},
		PrepareStmt: true,
	})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db, nil
}
