package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"

	"github.com/mbicl/mbicf_bot/adminlog"
	"github.com/mbicl/mbicf_bot/cf"
	"github.com/mbicl/mbicf_bot/config"
	"github.com/mbicl/mbicf_bot/db"
)

func main() {
	db.Connect()
	cf.GetAllProblems()

	var cancel context.CancelFunc
	config.Ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	var err error
	config.B, err = bot.New(config.BotToken, []bot.Option{}...)
	if nil != err {
		adminlog.Fatal(err.Error(), config.Ctx, config.B)
	}

	config.B.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypePrefix, startHandler)
	config.B.RegisterHandler(bot.HandlerTypeMessageText, "/handle", bot.MatchTypePrefix, userRegisterHandler)
	config.B.RegisterHandler(bot.HandlerTypeMessageText, "/gimme", bot.MatchTypePrefix, gimmeHandler)
	config.B.RegisterHandler(bot.HandlerTypeMessageText, "/standings", bot.MatchTypePrefix, standingsHandler)

	go dailyTaskSender(config.Ctx, config.B)
	go statsUpdater()

	adminlog.SendMessage("Bot started", config.Ctx, config.B)
	config.B.Start(config.Ctx)
}
