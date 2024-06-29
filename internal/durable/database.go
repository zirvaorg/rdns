package durable

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func ConnectDB(c string) error {
	var err error

	db, err = gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{
		SkipDefaultTransaction: false,
		Logger:                 logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		return err
	}
	return nil
}

func Connection() *gorm.DB {
	return db
}
