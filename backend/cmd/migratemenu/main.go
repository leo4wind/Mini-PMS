package main

import (
	"fmt"

	"minipms/internal/config"
	"minipms/internal/database"
)

// 一次性迁移：将「项目」一级目录并入「产品」，适配顶栏二级导航。
func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		panic(err)
	}
	db, err := database.Connect(cfg)
	if err != nil {
		panic(err)
	}

	steps := []string{
		"UPDATE menu SET parent_id = 2, sort = 30 WHERE id = 20",
		"UPDATE menu SET parent_id = 2, sort = 40 WHERE id = 21",
		"UPDATE menu SET name = '产品' WHERE id = 2",
		"DELETE FROM role_menu WHERE menu_id = 3",
		"DELETE FROM menu WHERE id = 3",
	}
	for _, sql := range steps {
		if err := db.Exec(sql).Error; err != nil {
			panic(fmt.Errorf("%s: %w", sql, err))
		}
		fmt.Println("ok:", sql)
	}
	fmt.Println("menu migrate done")
}
