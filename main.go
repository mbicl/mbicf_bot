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

	db.Connect()
	cf.GetAllProblems()

	err = config.DB.
		Preload("Easy").
		Preload("Medium").
		Preload("Advanced").
		Preload("Hard").
		First(&config.TodaysTasks).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Println(err.Error())
	}

	crn := cron.New()
	_, err = crn.AddFunc("@every 10m", func() {
		dailyTaskSender(config.Ctx, config.B)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	config.DB.First(&config.LastCheckedTime)
	config.LastCheckedTime.UnixTime = time.Now().Unix()
	config.DB.Save(&config.LastCheckedTime)
	_, err = crn.AddFunc("@every 30s", statsUpdater)
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
