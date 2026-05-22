package database

import (
	"blogapp/backend/internal/config"
	"blogapp/backend/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg config.Config) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Article{},
		&model.Comment{},
		&model.Reaction{},
		&model.Notification{},
	)
}

func SeedDefaults(db *gorm.DB) error {
	var count int64
	if err := db.Model(&model.Category{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	categories := []model.Category{
		{Name: "Go", Slug: "go"},
		{Name: "Vue", Slug: "vue"},
		{Name: "Database", Slug: "database"},
		{Name: "Frontend", Slug: "frontend"},
	}

	for _, item := range categories {
		if err := db.Create(&item).Error; err != nil {
			return err
		}
	}
	return nil
}
