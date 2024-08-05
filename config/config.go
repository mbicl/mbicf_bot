package config

import (
	"context"
	"os"

	"github.com/go-telegram/bot"
	"gorm.io/gorm"

	"github.com/mbicl/mbicf_bot/models"
)

var (
	BotToken    string
	DB          *gorm.DB
	Month       = []string{"", "yanvar", "fevral", "mart", "aprel", "may", "iyun", "iyul", "avgust", "sentyabr", "oktyabr", "noyabr", "dekabr"}
	TodaysTasks = &models.DailyTasks{}
	CFGroupID   = -1002120642025 // this is test grup ID (t.odo change to cf group id when releasing)
	GroupID     = -1001524140542 // this is Codeforces group ID
	FMessage    = "#dailytask #%d%s\n" +
		"%d-%s uchun kunlik masalalar.\n" +
		"🟢<a href=\"%s\">%s</a>(%d)\n" +
		"🟡<a href=\"%s\">%s</a>(%d)\n" +
		"🟠<a href=\"%s\">%s</a>(%d)\n" +
		"🔴<a href=\"%s\">%s</a>(%d)"
	LastCheckedTime = &models.LastCheckedTime{}
	B               *bot.Bot
	Ctx             context.Context
)

func init() {
	//err := godotenv.Load(".env")
	//if err != nil {
	//	log.Fatal("Error loading .env file")
	//}
	BotToken = os.Getenv("BOT_TOKEN")
}
