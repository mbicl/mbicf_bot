package adminlog

import (
	"context"
	"log"
	"os"

	"github.com/go-telegram/bot"
)

var (
	TGID = 947518051
)

func SendMessage(msg string, ctx context.Context, b *bot.Bot) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: TGID,
		Text:   msg,
	})
	if err != nil {
		log.Println("Error sending message to admin: ", err)
	}
}

func Fatal(msg string, ctx context.Context, b *bot.Bot) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: TGID,
		Text:   msg,
	})
	if err != nil {
		log.Println("Error sending message to admin: ", err)
	}
	os.Exit(1)
}
