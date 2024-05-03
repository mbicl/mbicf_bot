package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"

	"mbicf_bot/cf"
	"mbicf_bot/config"
	"mbicf_bot/db"
)

func main() {
	db.Connect()
	cf.GetAllProblems()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	b, err := bot.New(config.BotToken, []bot.Option{}...)
	if nil != err {
		log.Fatal(err.Error())
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypePrefix, startHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/handle", bot.MatchTypePrefix, userRegisterHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/gimme", bot.MatchTypePrefix, gimmeHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/standings", bot.MatchTypePrefix, standingsHandler)

	go dailyTaskSender(ctx, b)
	go statsUpdater()

	log.Println("Bot started")
	b.Start(ctx)
}
