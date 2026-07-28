package main

import (
	"fmt"

	"minipms/internal/config"
	"minipms/internal/database"
	"minipms/internal/model"
)

func main() {
	cfg, _ := config.Load("configs/config.yaml")
	db, _ := database.Connect(cfg)
	var ps []model.Product
	db.Where("deleted = 0").Order("id").Limit(3).Find(&ps)
	for _, p := range ps {
		fmt.Printf("product %d %s\n", p.ID, p.Name)
	}
	var ss []model.Story
	db.Where("deleted = 0").Order("id").Limit(3).Find(&ss)
	for _, s := range ss {
		fmt.Printf("story %d %s\n", s.ID, s.Title)
	}
	var bs []model.Bug
	db.Where("deleted = 0").Order("id").Limit(3).Find(&bs)
	for _, b := range bs {
		fmt.Printf("bug %d %s\n", b.ID, b.Title)
	}
	var us []model.User
	db.Where("deleted = 0 AND id > 1").Order("id").Limit(3).Find(&us)
	for _, u := range us {
		fmt.Printf("user %d %s %s\n", u.ID, u.Account, u.Realname)
	}
}
