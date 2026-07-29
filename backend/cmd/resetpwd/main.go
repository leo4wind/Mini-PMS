package main

import (
	"fmt"
	"log"
	"path/filepath"

	"minipms/internal/config"
	"minipms/internal/database"
	"minipms/internal/model"
	"minipms/internal/service"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg, err := config.Load(filepath.Join("configs", "config.yaml"))
	if err != nil {
		log.Fatal(err)
	}
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(service.DefaultUserPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	res := db.Model(&model.User{}).Where("deleted = 0").Update("password_hash", string(hash))
	if res.Error != nil {
		log.Fatal(res.Error)
	}
	fmt.Printf("updated %d users to password %s\n", res.RowsAffected, service.DefaultUserPassword)
}
