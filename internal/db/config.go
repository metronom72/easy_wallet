package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
