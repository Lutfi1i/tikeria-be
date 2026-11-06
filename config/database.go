package config

import (
	"fmt"
	"log"
	"os"

	"tikeria-be/models"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDatabase() *gorm.DB {
	dbDriver := os.Getenv("DB_DRIVER")
	dbSource := os.Getenv("DB_SOURCE")

	var db *gorm.DB
	var err error

	if dbDriver == "postgres" {
		db, err = gorm.Open(postgres.Open(dbSource), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	} else {
		if dbSource == "" {
			dbSource = "tikeria.db"
		}
		db, err = gorm.Open(sqlite.Open(dbSource), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	}
	if err != nil {
		log.Fatalf("Database not connected: %v", err)
	}
	fmt.Println("Database Connected")

	err = db.AutoMigrate(
		&models.User{},
		&models.Event{},
		&models.Ticket{},
		&models.TicketHold{},
		&models.Order{},
	)

	if err != nil {
		log.Fatalf("Migration Failed, Error with Connection: %v", err)
	}

	fmt.Println("Migration successfully")

	DB = db
	return db
}
