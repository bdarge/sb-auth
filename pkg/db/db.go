package db

import (
	"gorm.io/driver/mysql"
	"log"

	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

func Init(url string) Handler {
	log.Printf("Initialize db")
	db, err := gorm.Open(mysql.Open(url), &gorm.Config{})

	if err != nil {
		log.Fatalln(err)
	}

	return Handler{db}
}
