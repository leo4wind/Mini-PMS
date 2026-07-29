package applog

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type Level int

const (
	LevelInfo Level = iota
	LevelDebug
)

var (
	current = LevelDebug
	out     = log.New(os.Stdout, "", log.LstdFlags)
)

func SetLevel(level string) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "info":
		current = LevelInfo
	default:
		current = LevelDebug
	}
}

func LevelName() string {
	if current == LevelInfo {
		return "info"
	}
	return "debug"
}

// Info 打印 INFO 日志（HTTP 等）
func Info(format string, args ...interface{}) {
	if current == LevelInfo || current == LevelDebug {
		out.Printf("[INFO] "+format, args...)
	}
}

// Debug 打印 DEBUG 日志（SQL 等）；仅 debug 级别输出
func Debug(format string, args ...interface{}) {
	if current == LevelDebug {
		out.Printf("[DEBUG] "+format, args...)
	}
}

func FormatDuration(d time.Duration) string {
	return fmt.Sprintf("%.3fms", float64(d.Microseconds())/1000.0)
}
