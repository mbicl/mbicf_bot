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
	GroupID     = -1002120642025 // todo change to cf group id when releasing
	CFGroupID   = -1001524140542
	FMessage    = "#dailytask #%d%s\n" +
		"%d-%s uchun kunlik masalalar.\n" +
		"🟢Easy:         <a href=\"%s\">%s</a>\n" +
		"🟡Medium:    <a href=\"%s\">%s</a>\n" +
		"🟠Advanced: <a href=\"%s\">%s</a>\n" +
		"🔴Hard:         <a href=\"%s\">%s</a>"
	LastCheckedTime = &models.LastCheckedTime{}
	//UserStatusMap   = make(map[string]cfmodels.UserStatus)
	B   *bot.Bot
	Ctx context.Context
)

func init() {
	//err := godotenv.Load(".env")
	//if err != nil {
	//	log.Fatal("Error loading .env file")
	//}
	BotToken = os.Getenv("BOT_TOKEN")
}
