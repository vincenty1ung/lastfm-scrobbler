package model

import (
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/vincenty1ung/lastfm-scrobbler/core/db"
)

var GlobalDB *gorm.DB

func GetDB() *gorm.DB {
	return GlobalDB
}

func InitDB(dataSourceName string, l *zap.Logger) error {
	var err error

	// Create custom logger with OpenTelemetry
	customLogger := db.NewCustomLogger(l)

	// Open database with custom logger
	GlobalDB, err = gorm.Open(
		sqlite.Open(dataSourceName), &gorm.Config{
			Logger: customLogger,
		},
	)
	if err != nil {
		return err
	}

	// Auto migrate the schema for TrackPlayRecord
	err = GlobalDB.AutoMigrate(&TrackPlayRecord{})
	if err != nil {
		return err
	}

	// Auto migrate the schema for Track
	err = GlobalDB.AutoMigrate(&Track{})
	if err != nil {
		return err
	}

	return nil
}
