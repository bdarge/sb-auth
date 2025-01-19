package db

import (
	"gorm.io/driver/mysql"
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Handler db handler
type Handler struct {
	DB *gorm.DB
}

// Init database connection
func Init(url string) Handler {
	log.Printf("Initialize db connection")
	db, err := gorm.Open(mysql.Open(url), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalln(err)
	}

	return Handler{db}
}
