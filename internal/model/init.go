package model

import (
	"errors"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/vincenty1ung/lastfm-scrobbler/common"
	"github.com/vincenty1ung/lastfm-scrobbler/config"
	"github.com/vincenty1ung/lastfm-scrobbler/core/db"
)

var GlobalDBForSqlLite *gorm.DB
var GlobalDBForMysql *gorm.DB

func GetDB() *gorm.DB {
	if config.ConfigObj.Database.Type == string(common.DatabaseTypeMySQL) {
		return GlobalDBForMysql
	}
	return GlobalDBForSqlLite
}

func InitDB(dataSourceName string, l *zap.Logger) error {
	var err error

	// Create custom logger with OpenTelemetry
	customLogger := db.NewCustomLogger(l)
	switch config.ConfigObj.Database.Type {
	case string(common.DatabaseTypeSQLite):
		// Open SQLite database with custom logger
		GlobalDBForSqlLite, err = gorm.Open(
			sqlite.Open(dataSourceName), &gorm.Config{
				Logger: customLogger,
			},
		)
		if err != nil {
			return err
		}

		// Auto migrate the schema for TrackPlayRecord
		err = GlobalDBForSqlLite.AutoMigrate(&TrackPlayRecord{})
		if err != nil {
			return err
		}

		// Auto migrate the schema for Track
		err = GlobalDBForSqlLite.AutoMigrate(&Track{})
		if err != nil {
			return err
		}

		// Auto migrate the schema for Genre
		err = GlobalDBForSqlLite.AutoMigrate(&Genre{})
		if err != nil {
			return err
		}
	case string(common.DatabaseTypeMySQL):
		// Open MySQL database with custom logger
		GlobalDBForMysql, err = gorm.Open(
			mysql.Open(db.MysqlDSN(config.ConfigObj.Database.Mysql.GetMysqlDSN())), &gorm.Config{
				Logger: customLogger,
			},
		)
		if err != nil {
			return err
		}
		if config.ConfigObj.IsDev {
			// Auto migrate the schema for TrackPlayRecord
			err = GlobalDBForMysql.AutoMigrate(&TrackPlayRecord{})
			if err != nil {
				return err
			}

			// Auto migrate the schema for Track
			err = GlobalDBForMysql.AutoMigrate(&Track{})
			if err != nil {
				return err
			}

			// Auto migrate the schema for Genre
			err = GlobalDBForMysql.AutoMigrate(&Genre{})
			if err != nil {
				return err
			}
		}
	default:
		return errors.New("unsupported database type" + config.ConfigObj.Database.Type)
	}

	return nil
}
