package repository

import (
	"fmt"
	"llm-wiki-server/internal/config"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(c config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s", c.Host, c.Port, c.Username, c.Password, c.DatabaseName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
		return nil, err
	}

	return db, nil
}
