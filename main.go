package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/go-telegram/bot"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"

	"github.com/mbicl/mbicf_bot/adminlog"
	"github.com/mbicl/mbicf_bot/cf"
	"github.com/mbicl/mbicf_bot/config"
	"github.com/mbicl/mbicf_bot/db"
	"github.com/mbicl/mbicf_bot/models"
)

func main() {
	var cancel context.CancelFunc
	config.Ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	var err error
	config.B, err = bot.New(config.BotToken, []bot.Option{}...)
	if nil != err {
		log.Fatal(err.Error())
	}

	config.B.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypePrefix, startHandler)
	config.B.RegisterHandler(bot.HandlerTypeMessageText, "/handle", bot.MatchTypePrefix, userRegisterHandler)
	config.B.RegisterHandler(bot.HandlerTypeMessageText, "/gimme", bot.MatchTypePrefix, gimmeHandler)
	config.B.RegisterHandler(bot.HandlerTypeMessageText, "/standings", bot.MatchTypePrefix, standingsHandler)
	config.B.RegisterHandler(bot.HandlerTypeMessageText, "/iamdone", bot.MatchTypePrefix, iAmDoneHandler)
	config.B.RegisterHandler(bot.HandlerTypeMessageText, "/dailytasks", bot.MatchTypePrefix, dailyTasksHandler)

	config.B.RegisterHandler(bot.HandlerTypeMessageText, "/givemedatabasefile", bot.MatchTypePrefix, databaseBackupHandler)
	config.B.RegisterHandler(bot.HandlerTypeMessageText, "/updateusersdata", bot.MatchTypePrefix, updateUsersDataHandler)

	db.Connect()
	cf.GetAllProblems()

	err = config.DB.
		Model(&models.DailyTasks{}).
		Preload("Easy").
		Preload("Medium").
		Preload("Advanced").
		Preload("Hard").
		First(config.TodaysTasks).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Println(err.Error())
	}
	log.Println(config.TodaysTasks)

	location, err := time.LoadLocation("Asia/Tashkent")
	if err != nil {
		log.Fatal(err.Error())
	}
	crn := cron.New(cron.WithLocation(location))
	_, err = crn.AddFunc("0 8 * * *", func() {
		dailyTaskSender(config.Ctx, config.B)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	config.DB.First(&config.LastCheckedTime)
	if config.LastCheckedTime.UnixTime == 0 {
		config.LastCheckedTime.UnixTime = time.Now().Unix()
		config.DB.Save(&config.LastCheckedTime)
	}
	_, err = crn.AddFunc("@every 5m", func() {
		_ = statsUpdater2()
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	_, err = crn.AddFunc("@daily", updateUsersData)
	if err != nil {
		log.Fatal(err.Error())
	}
	crn.Start()

	adminlog.SendMessage("Bot started", config.Ctx, config.B)

	config.B.Start(config.Ctx)
}
