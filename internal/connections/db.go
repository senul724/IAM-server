package connections

import (
	"log"
	"os"

	"IAM-server/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() {
	conn := os.Getenv("DATABASE_URL")
	if conn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := gorm.Open(postgres.Open(conn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // logs SQL in dev
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	DB = db
	log.Println("database connected")
}

func Migrate() {
	err := DB.AutoMigrate(
		&models.Customer{},
		&models.Note{},
		&models.Session{},
	)
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("migration completed")
}