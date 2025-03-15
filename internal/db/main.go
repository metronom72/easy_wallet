package db

import (
	"easy-wallet/internal/models"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()
	dsn := composeDSN()
	createDB()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&models.User{})
	extractSchema(dsn, "internal/db/schema.sql")
}

func composeDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		getEnv("DATABASE_USER", "postgres"),
		getEnv("DATABASE_PASSWORD", ""),
		getEnv("DATABASE_HOST", "localhost"),
		getEnv("DATABASE_PORT", "5432"),
		getEnv("DATABASE_NAME", "postgres"),
		getEnv("DATABASE_SSLMODE", "disable"),
	)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func createDB() {
	adminDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable",
		getEnv("DATABASE_USER", "postgres"),
		getEnv("DATABASE_PASSWORD", ""),
		getEnv("DATABASE_HOST", "localhost"),
		getEnv("DATABASE_PORT", "5432"),
	)
	db, err := gorm.Open(postgres.Open(adminDSN), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db.Exec(fmt.Sprintf("CREATE DATABASE %s;", getEnv("DATABASE_NAME", "postgres")))
}

func extractSchema(dsn, outputFile string) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf(`pg_dump --schema-only --no-owner --no-privileges -d "%s" | sed '/^--/ d'`, dsn))
	sqlSchema, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	_ = os.WriteFile(outputFile, sqlSchema, 0644)
}
