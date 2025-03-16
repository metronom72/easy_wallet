package db

import (
	"context"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

type DatabaseConnection struct {
	DB *gorm.DB
}

func (conn *DatabaseConnection) GetConnection(dsn string) (*gorm.DB, error) {
	if conn.DB == nil {
		connection, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, err
		}

		conn.DB = connection
	}

	return conn.DB, nil
}

func InjectDBToContext(ctx context.Context) context.Context {
	databaseName := os.Getenv("DATABASE_NAME")
	databaseUser := os.Getenv("DATABASE_USER")
	databasePassword := os.Getenv("DATABASE_PASSWORD")
	databaseHost := os.Getenv("DATABASE_HOST")
	databasePort := os.Getenv("DATABASE_PORT")
	databaseSSLMode := os.Getenv("DATABASE_SSLMODE")
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		databaseUser,
		databasePassword,
		databaseHost,
		databasePort,
		databaseName,
		databaseSSLMode,
	)

	conn := DatabaseConnection{}
	db, err := conn.GetConnection(dsn)
	if err != nil {
		log.Println("[ERROR] Failed to connect to database")
		return ctx
	}

	return context.WithValue(ctx, "db", db)
}

func GetDBFromContext(ctx context.Context) (*gorm.DB, bool) {
	db, ok := ctx.Value("db").(*gorm.DB)
	return db, ok
}
