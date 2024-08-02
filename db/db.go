package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/mbicl/mbicf_bot/adminlog"
	"github.com/mbicl/mbicf_bot/config"
	"github.com/mbicl/mbicf_bot/models"
)

func Connect() {
	var err error
	config.DB, err = gorm.Open(sqlite.Open("sqlite.db"), &gorm.Config{})
	if err != nil {
		adminlog.Fatal("Could not open database file", config.Ctx, config.B)
	}
	adminlog.SendMessage("Connected to database", config.Ctx, config.B)

	err = config.DB.AutoMigrate(&models.User{}, &models.Problem{}, &models.UsedProblem{}, &models.Attempt{}, &models.DailyTasks{}, &models.LastCheckedTime{})
	if err != nil {
		adminlog.Fatal("Could not migrate User struct", config.Ctx, config.B)
	}

	adminlog.SendMessage("Database migrated", config.Ctx, config.B)
}
