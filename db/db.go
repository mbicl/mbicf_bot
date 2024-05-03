package db

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"mbicf_bot/config"
	"mbicf_bot/models"
)

func Connect() {
	var err error
	config.DB, err = gorm.Open(sqlite.Open("sqlite.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Could not open database file")
	}
	log.Println("Connected to database")

	err = config.DB.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal("Could not migrate User struct")
	}
	err = config.DB.AutoMigrate(&models.Problem{})
	if err != nil {
		log.Fatal("Could not migrate Problem struct")
	}
	err = config.DB.AutoMigrate(&models.UsedProblem{})
	if err != nil {
		log.Fatal("Could not migrate UsedProblem struct")
	}
	err = config.DB.AutoMigrate(&models.LastCheckedTime{})
	if err != nil {
		log.Fatal("Could not migrate LastCheckedTime")
	}

	log.Println("Database migrated")
}
