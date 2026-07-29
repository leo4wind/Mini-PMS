package main

import (
	"log"
	"os"
	"path/filepath"

	"minipms/internal/config"
	"minipms/internal/database"
	"minipms/internal/pkg/applog"
	"minipms/internal/router"
)

func main() {
	cfgPath := os.Getenv("MINIPMS_CONFIG")
	if cfgPath == "" {
		cfgPath = filepath.Join("configs", "config.yaml")
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	applog.SetLevel(cfg.Log.Level)
	applog.Info("log level=%s", applog.LevelName())
	if err := os.MkdirAll(cfg.Upload.Dir, 0o755); err != nil {
		log.Fatalf("mkdir upload: %v", err)
	}
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	engine := router.Setup(cfg, db)
	applog.Info("MiniPMS API listening on %s", cfg.Server.Addr)
	if err := engine.Run(cfg.Server.Addr); err != nil {
		log.Fatal(err)
	}
}
