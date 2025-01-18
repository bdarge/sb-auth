package db

import (
	"gorm.io/driver/mysql"
	"log"

	"gorm.io/gorm"
)

// Handler db handler
type Handler struct {
	DB *gorm.DB
}

// Init connect to database
func Init(url string) Handler {
	log.Printf("Connect to db")
	db, err := gorm.Open(mysql.Open(url), &gorm.Config{})

	if err != nil {
		log.Fatalln(err)
	}

	return Handler{db}
}
