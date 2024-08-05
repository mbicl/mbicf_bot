package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fogleman/gg"
	"github.com/go-telegram/bot"
	botModels "github.com/go-telegram/bot/models"

	"github.com/mbicl/mbicf_bot/adminlog"
	"github.com/mbicl/mbicf_bot/cf"
	cfmodels "github.com/mbicl/mbicf_bot/cf/models"
	"github.com/mbicl/mbicf_bot/config"
	"github.com/mbicl/mbicf_bot/models"
	"github.com/mbicl/mbicf_bot/utils"
)

func defaultHandler(ctx context.Context, b *bot.Bot, update *botModels.Update) {
	_, err := b.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: update.Message.Chat.ID,
		Action: botModels.ChatActionTyping,
	})

	if err != nil {
		adminlog.SendMessage("Error sending chat action"+err.Error(), ctx, b)
		return
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Hello",
	})

	if err != nil {
		adminlog.SendMessage("Error sending message: "+err.Error(), ctx, b)
	}
}

func startHandler(ctx context.Context, b *bot.Bot, update *botModels.Update) {
	if update.Message.Chat.Type != "private" {
		return
	}
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text: "Assalomu alaykum, botdan foydalanish bo'yicha qo'llanma:\n" +
			"/handle <cfhandle> - ro'yxatdan o'tish.\n" +
			"/iamdone - urinish qilgandan so'ng statistikani yangilash\n" +
			"/standings - natijalar jadvalini ko'rish\n" +
			"/gimme - reytingingizga mos tasodifiy masala olish\n" +
			"/gimme <son> - ko'rsatilgan reytingda tasodifiy masala olish\n" +
			"/gimme <±delta> - reytingingizdan ±deltaga farq qiluvchi tasodifiy masala olish\n",
	})
	if err != nil {
		adminlog.SendMessage("Error sending message: "+err.Error(), ctx, b)
	}
}

func userRegisterHandler(ctx context.Context, b *bot.Bot, update *botModels.Update) {
	if update.Message.Chat.Type != "private" {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			ReplyParameters: &botModels.ReplyParameters{
				ChatID:    update.Message.Chat.ID,
				MessageID: update.Message.ID,
			},
			Text: "Botga shaxsiy xabar jo'natish orqali ro'yxatdan o'tishingiz mumkin.",
		})
		if err != nil {
			adminlog.SendMessage("Error sending message: "+err.Error(), ctx, b)
		}
		return
	}
	splittedMessage := strings.Split(update.Message.Text, " ")
	if len(splittedMessage) != 2 {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Noto'g'ri buyruq.\n/handle your_handle ko'rinishida yuboring",
		})
		if err != nil {
			adminlog.SendMessage("Error sending message: "+err.Error(), ctx, b)
		}
		return
	}
	handle := splittedMessage[1]
	body := utils.HTTPGet(cf.BaseURL + "user.info?handles=" + handle)
	userInfo := cfmodels.UserInfo{}
	err := json.Unmarshal(body, &userInfo)
	if err != nil {
		adminlog.SendMessage("Error unmarshalling user info: "+err.Error()+string(body), ctx, b)
		return
	}

	if len(userInfo.Result) == 0 {
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Bunday foydalanuvchi nomi codeforcesda mavjud emas.",
		})
		if err != nil {
			adminlog.SendMessage("Error sending message: "+err.Error(), ctx, b)
		}
		return
	}

	user := userInfo.Result[0]
	problem := cf.GetRandomProblem()

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   problem.Link + " masalaga 60 soniya ichida kompilyatsiya xatoligi beradigan kod jo'nating.",
	})
	if err != nil {
		adminlog.SendMessage("Error sending message: "+err.Error(), ctx, b)
		return
	}

	start := time.Now()
	for {
		time.Sleep(1 * time.Second)
		if time.Since(start) > time.Second*60 {
			_, err = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Vaqt tugadi, ulgurmadingiz.",
			})
			if err != nil {
				adminlog.SendMessage("Error sending message: "+err.Error(), ctx, b)
			}
			break
		}
		body := utils.HTTPGet(cf.BaseURL + "user.status?handle=" + user.Handle + "&from=1&count=1")
		if body == nil || len(body) == 0 {
			adminlog.SendMessage("Error on getting status for user "+user.Handle, ctx, b)
			continue
		}
		userStatus := cfmodels.UserStatus{}
		err = json.Unmarshal(body, &userStatus)
		if err != nil {
			adminlog.SendMessage("Error unmarshalling user status(userRegistration):"+err.Error(), ctx, b)
			continue
		}
		if len(userStatus.Result) == 0 {
			continue
		}
		submission := userStatus.Result[0]
		if strconv.Itoa(submission.Problem.ContestID)+submission.Problem.Index == problem.CFID &&
			submission.Verdict == "COMPILATION_ERROR" {
			newUser := models.User{}
			config.DB.Where("cf_handle = ?", user.Handle).First(&newUser)
			if newUser.TGUserID != 0 {
				_, err = b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Ro'yxatdan o'tgan.",
				})
				if err != nil {
					adminlog.SendMessage("Error sending message: "+err.Error(), ctx, b)
				}
				return
			}
			newUser = models.User{}
			config.DB.Where("tg_user_id = ?", update.Message.Chat.ID).First(&newUser)
			if newUser.TGUserID != 0 {
				config.DB.Model(&models.User{}).Where("tg_user_id = ?", update.Message.Chat.ID).Updates(models.User{
					FirstName:  user.FirstName,
					LastName:   user.LastName,
					CFHandle:   user.Handle,
					CFRating:   user.Rating,
					TGUserName: update.Message.Chat.Username,
				})
				_, err = b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Ma'lumotlaringiz yangilandi.",
				})
				if err != nil {
					adminlog.SendMessage("Error sending message: "+err.Error(), ctx, b)
				}
				break
			}
			newUser = models.User{
				FirstName:  user.FirstName,
				LastName:   user.LastName,
				CFHandle:   user.Handle,
				CFRating:   user.Rating,
				TGUserName: update.Message.Chat.Username,
				TGUserID:   update.Message.Chat.ID,
			}
			config.DB.Save(&newUser)
			_, err = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Muvaffaqiyatli ro'yxatdan o'tdingiz.",
			})
			if err != nil {
				adminlog.SendMessage("Error sending message: "+err.Error(), ctx, b)
			}
			adminlog.SendMessage("Yangi foydalanuvchi: "+newUser.String(), ctx, b)
			break
		}
	}
}

func gimmeHandler(ctx context.Context, b *bot.Bot, update *botModels.Update) {
	msgTokens := strings.Split(update.Message.Text, " ")
	if len(msgTokens) > 2 {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			ReplyParameters: &botModels.ReplyParameters{
				ChatID:    update.Message.Chat.ID,
				MessageID: update.Message.ID,
			},
			Text: "Noto'g'ri format.\n Foydalanish uchun misollar:\n/gimme\n/gimme 1500\n/gimme +100\n/gimme -200",
		})
		if err != nil {
			adminlog.SendMessage("Error sending message: "+err.Error(), ctx, b)
		}
		return
	}
	if len(msgTokens) == 1 {
		user := models.User{}
		if update.Message.Chat.Type != "private" {
			config.DB.Where("tg_user_id = ?", update.Message.From.ID).First(&user)
		} else {
			config.DB.Where("tg_user_id = ?", update.Message.Chat.ID).First(&user)
		}
		if len(user.CFHandle) == 0 {
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				ReplyParameters: &botModels.ReplyParameters{
					ChatID:    update.Message.Chat.ID,
					MessageID: update.Message.ID,
				},
				Text: "Botdan ro'yxatdan o'tmagansiz.",
			})
			if err != nil {
				adminlog.SendMessage("Error sending message: "+err.Error(), ctx, b)
			}
			return
		}

		rating := user.CFRating
		rating = int(math.Round(float64(rating)/100.0)) * 100
		rating = max(rating, 800)
		rating = min(rating, 3500)

		problem := cf.GetRandomProblemWithRating(rating)
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			ReplyParameters: &botModels.ReplyParameters{
				ChatID:    update.Message.Chat.ID,
				MessageID: update.Message.ID,
			},
			Text:      fmt.Sprintf("<a href=\"%s\">%s</a> (%d)", problem.Link, problem.Name, problem.Rating),
			ParseMode: botModels.ParseModeHTML,
		})
		if err != nil {
			adminlog.SendMessage("Error sending message: "+err.Error(), ctx, b)
		}
		return
	}
	ratingStr := msgTokens[1]
	if ratingStr[0] == '+' || ratingStr[0] == '-' {
		ratingDelta, err := strconv.Atoi(ratingStr)
		if err != nil {
			adminlog.SendMessage("Error converting gimme rating to int:"+err.Error(), ctx, b)
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				ReplyParameters: &botModels.ReplyParameters{
					ChatID:    update.Message.Chat.ID,
					MessageID: update.Message.ID,
				},
				Text: "Noto'g'ri format.\n Foydalanish uchun misollar:\n/gimme\n/gimme 1500\n/gimme +100\n/gimme -200",
			})
			if err != nil {
				adminlog.SendMessage("Error sending message: "+err.Error(), ctx, b)
			}
			return
		}
		user := models.User{}
		if update.Message.Chat.Type != "private" {
			config.DB.Where("tg_user_id = ?", update.Message.From.ID).First(&user)
		} else {
			config.DB.Where("tg_user_id = ?", update.Message.Chat.ID).First(&user)
		}
		if len(user.CFHandle) == 0 {
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				ReplyParameters: &botModels.ReplyParameters{
					ChatID:    update.Message.Chat.ID,
					MessageID: update.Message.ID,
				},
				Text: "Botdan ro'yxatdan o'tmagansiz.",
			})
			if err != nil {
				adminlog.SendMessage("Error sending message: "+err.Error(), ctx, b)
			}
			return
		}

		rating := user.CFRating + ratingDelta
		rating = int(math.Round(float64(rating)/100.0)) * 100
		rating = max(rating, 800)
		rating = min(rating, 3500)
		problem := cf.GetRandomProblemWithRating(rating)
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			ReplyParameters: &botModels.ReplyParameters{
				ChatID:    update.Message.Chat.ID,
				MessageID: update.Message.ID,
			},
			Text:      fmt.Sprintf("<a href=\"%s\">%s</a> (%d)", problem.Link, problem.Name, problem.Rating),
			ParseMode: botModels.ParseModeHTML,
		})
	} else {
		rating, err := strconv.Atoi(ratingStr)
		if err != nil {
			adminlog.SendMessage("Error converting gimme rating to int:"+err.Error(), ctx, b)
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				ReplyParameters: &botModels.ReplyParameters{
					ChatID:    update.Message.Chat.ID,
					MessageID: update.Message.ID,
				},
				Text: "Noto'g'ri format.\n Foydalanish uchun misollar:\n/gimme\n/gimme 1500\n/gimme +100\n/gimme -200",
			})
			if err != nil {
				adminlog.SendMessage("Error sending message: "+err.Error(), ctx, b)
			}
			return
		}
		rating = int(math.Round(float64(rating)/100.0)) * 100
		rating = max(rating, 800)
		rating = min(rating, 3500)
		problem := cf.GetRandomProblemWithRating(rating)
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			ReplyParameters: &botModels.ReplyParameters{
				ChatID:    update.Message.Chat.ID,
				MessageID: update.Message.ID,
			},
			Text:      fmt.Sprintf("<a href=\"%s\">%s</a> (%d)", problem.Link, problem.Name, problem.Rating),
			ParseMode: botModels.ParseModeHTML,
		})
		if err != nil {
			adminlog.SendMessage("Error sending message: "+err.Error(), ctx, b)
		}
	}
}

func standingsHandler(ctx context.Context, b *bot.Bot, update *botModels.Update) {
	err501 := func() {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			ReplyParameters: &botModels.ReplyParameters{
				ChatID:    update.Message.Chat.ID,
				MessageID: update.Message.ID,
			},
			Text: "501",
		})
		if err != nil {
			adminlog.SendMessage("Error sending message: "+err.Error(), ctx, b)
		}
	}

	users := make([]models.User, 0)
	err := config.DB.Order("rating desc, cf_rating desc").Limit(20).Find(&users).Error
	if err != nil {
		adminlog.SendMessage("Error on fetching users for standings: "+err.Error(), ctx, b)
		err501()
		return
	}

	const (
		HEIGHT = 670
		WIDTH  = 1440
		font   = "JetBrainsMono-Regular.ttf"
	)

	dc := gg.NewContext(WIDTH, HEIGHT)

	// Header and footer background
	dc.SetHexColor("363E3F")
	dc.DrawRectangle(0, 0, WIDTH, 60)
	dc.DrawRectangle(0, 660, WIDTH, 10)
	dc.Fill()

	// Header
	err = dc.LoadFontFace(font, 25)
	if err != nil {
		adminlog.SendMessage("Error on loading font: "+err.Error(), ctx, b)
		err501()
		return
	}

	dc.SetHexColor("F3F3F3")
	dc.DrawString("#", 14, 50)
	dc.DrawString("Name", 60, 50)
	dc.DrawString("Handle", 460, 50)
	dc.DrawString("Points|Rating", 825, 50)

	dc.SetHexColor("77B058")
	dc.DrawCircle(1090, 30, 17)
	dc.Fill()
	dc.SetHexColor("FDCB58")
	dc.DrawCircle(1190, 30, 17)
	dc.Fill()
	dc.SetHexColor("F3900D")
	dc.DrawCircle(1290, 30, 17)
	dc.Fill()
	dc.SetHexColor("DC2D44")
	dc.DrawCircle(1390, 30, 17)
	dc.Fill()

	// Table
	for y := 60.0; y < HEIGHT-10; y += 30 {
		if int((y-40)/30)%2 == 0 {
			dc.SetHexColor("FFFFFF")
			dc.DrawRectangle(0, y, WIDTH, 30)
			dc.Fill()
		} else {
			dc.SetHexColor("F8F8F8")
			dc.DrawRectangle(0, y, WIDTH, 30)
			dc.Fill()
		}
	}

	dc.SetHexColor("AAAAAA")
	dc.DrawLine(40.0, 0.0, 40.0, HEIGHT)
	dc.DrawLine(440.0, 0.0, 440.0, HEIGHT)
	dc.DrawLine(800.0, 0.0, 800.0, HEIGHT)
	dc.DrawLine(1040.0, 0.0, 1040.0, HEIGHT)
	dc.DrawLine(1140.0, 0.0, 1140.0, HEIGHT)
	dc.DrawLine(1240.0, 0.0, 1240.0, HEIGHT)
	dc.DrawLine(1340.0, 0.0, 1340.0, HEIGHT)
	dc.Stroke()

	for i := 0; i < 20; i++ {
		dc.DrawStringAnchored(fmt.Sprintf("%d", i), 20, float64(75+i*30), 0.5, 0.5)
	}

	for i, user := range users {
		color := cf.ColorNewbie
		if user.CFRating < 1200 {
			color = cf.ColorNewbie
		} else if user.CFRating < 1400 {
			color = cf.ColorPupil
		} else if user.CFRating < 1600 {
			color = cf.ColorSpec
		} else if user.CFRating < 1900 {
			color = cf.ColorExpert
		} else if user.CFRating < 2100 {
			color = cf.ColorCM
		} else if user.CFRating < 2300 {
			color = cf.ColorMaster
		} else if user.CFRating < 2400 {
			color = cf.ColorIMaster
		} else if user.CFRating < 2600 {
			color = cf.ColorGM
		} else if user.CFRating < 3000 {
			color = cf.ColorIGM
		} else {
			color = cf.ColorLGM
		}
		dc.SetHexColor(color)
		y := float64(75 + i*30)
		dc.DrawString(fmt.Sprintf("%s %s", user.FirstName, user.LastName), 60, y+8)
		dc.DrawString(fmt.Sprintf("%s", user.CFHandle), 460, y+8)
		dc.DrawString(fmt.Sprintf("%7.2f|%d", user.Rating, user.CFRating), 825, y+8)
		dc.DrawStringAnchored(fmt.Sprintf("%d", user.CountEasy), 1090, y, 0.5, 0.5)
		dc.DrawStringAnchored(fmt.Sprintf("%d", user.CountMedium), 1190, y, 0.5, 0.5)
		dc.DrawStringAnchored(fmt.Sprintf("%d", user.CountAdvanced), 1290, y, 0.5, 0.5)
		dc.DrawStringAnchored(fmt.Sprintf("%d", user.CountHard), 1390, y, 0.5, 0.5)
	}

	err = dc.SavePNG("standings.png")
	if err != nil {
		adminlog.SendMessage("Error while saving image: "+err.Error(), ctx, b)
		err501()
		return
	}

	standingsPhoto, err := os.ReadFile("standings.png")
	if err != nil {
		adminlog.SendMessage("Error while reading standings photo: "+err.Error(), ctx, b)
		err501()
		return
	}

	_, err = b.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID: update.Message.Chat.ID,
		Photo:  &botModels.InputFileUpload{Filename: "standings.png", Data: bytes.NewReader(standingsPhoto)},
		ReplyParameters: &botModels.ReplyParameters{
			ChatID:    update.Message.Chat.ID,
			MessageID: update.Message.ID,
		},
	})
	if err != nil {
		adminlog.SendMessage("Error while sending standings photo: "+err.Error(), ctx, b)
	}
}

func dailyTaskSender(ctx context.Context, b *bot.Bot) {
	cf.UpdateTodaysTasks()
	err := config.DB.Save(&config.TodaysTasks).Error
	if err != nil {
		adminlog.SendMessage("Error while updating daily tasks: "+err.Error(), ctx, b)
		return
	}
	adminlog.SendMessage("Daily tasks updated: \n"+config.TodaysTasks.String(), ctx, b)

	_, err = b.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: config.GroupID,
		Action: botModels.ChatActionTyping,
	})
	if err != nil {
		adminlog.SendMessage("Error sending chat action: "+err.Error(), ctx, b)
	}
	msg := fmt.Sprintf(
		config.FMessage,
		time.Now().Day(),
		config.Month[time.Now().Month()],
		time.Now().Day(),
		config.Month[time.Now().Month()],
		config.TodaysTasks.Easy.Link,
		config.TodaysTasks.Easy.Name,
		config.TodaysTasks.Easy.Rating,
		config.TodaysTasks.Medium.Link,
		config.TodaysTasks.Medium.Name,
		config.TodaysTasks.Medium.Rating,
		config.TodaysTasks.Advanced.Link,
		config.TodaysTasks.Advanced.Name,
		config.TodaysTasks.Advanced.Rating,
		config.TodaysTasks.Hard.Link,
		config.TodaysTasks.Hard.Name,
		config.TodaysTasks.Hard.Rating,
	)

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    config.GroupID,
		Text:      msg,
		ParseMode: botModels.ParseModeHTML,
	})
	if err != nil {
		adminlog.SendMessage("Error sending daily task: "+err.Error(), ctx, b)
		return
	}
}

func statsUpdater() {
	users := make([]models.User, 0)
	config.DB.Find(&users)
	usedProblems := make([]models.UsedProblem, 0)

	config.DB.Find(&usedProblems)

	updateCheckedTime := true
	for _, user := range users {
		newAttempts := cf.GetLatestAttempts(user.CFHandle)
		if len(newAttempts) == 0 {
			continue
		}

		userRatingDelta := 0
		countEasyDelta := 0
		countMediumDelta := 0
		countAdvancedDelta := 0
		countHardDelta := 0

		attemptCountDelta := 0
		OKCountDelta := 0
		SolvedCountDelta := 0

		userAttempts := make([]models.Attempt, 0)
		err := config.DB.
			Model(&models.Attempt{}).
			Preload("User").
			Preload("UsedProblem").
			Where("user_id = ?", user.ID).
			Find(&userAttempts).
			Error
		if err != nil {
			adminlog.SendMessage(fmt.Sprintf("Error getting user %s attempts from DB: %s", user.CFHandle, err.Error()), config.Ctx, config.B)
		}
		for _, usedProblem := range usedProblems {
			isSolved := false
			for _, userAttempt := range userAttempts {
				if userAttempt.UsedProblem.CFID == usedProblem.CFID && userAttempt.Verdict == "OK" {
					isSolved = true
					break
				}
			}
			for _, newAttempt := range newAttempts {
				if newAttempt.Verdict == "TESTING" {
					updateCheckedTime = false
					continue
				}
				if newAttempt.UsedProblem.CFID == usedProblem.CFID {
					usedProblem.AttemptsCount++
					attemptCountDelta++
					if newAttempt.Verdict == "OK" {
						usedProblem.OKCount++
						OKCountDelta++
						if !isSolved {
							SolvedCountDelta++
							if usedProblem.Rating > 700 && usedProblem.Rating < 1100 {
								countEasyDelta++
							}
							if usedProblem.Rating > 1000 && usedProblem.Rating < 1600 {
								countMediumDelta++
							}
							if usedProblem.Rating > 1500 && usedProblem.Rating < 2100 {
								countAdvancedDelta++
							}
							if usedProblem.Rating > 2000 {
								countHardDelta++
							}
							if usedProblem.CFID == config.TodaysTasks.Easy.CFID ||
								usedProblem.CFID == config.TodaysTasks.Medium.CFID ||
								usedProblem.CFID == config.TodaysTasks.Advanced.CFID ||
								usedProblem.CFID == config.TodaysTasks.Hard.CFID {
								delta := max(1, (usedProblem.Rating-user.CFRating+50)/100*2)
								fmt.Printf("delta of user %s for problem %s is %d\n", user.CFHandle, usedProblem.CFID, delta)
								userRatingDelta += delta
							}
							isSolved = true
						}
					}
				}
			}
			err := config.DB.Save(&usedProblem).Error
			if err != nil {
				adminlog.SendMessage("Error on updating used problem stats: "+err.Error(), config.Ctx, config.B)
			}
		}

		for _, newAttempt := range newAttempts {
			if newAttempt.Verdict == "TESTING" {
				continue
			}
			for _, usedProblem := range usedProblems {
				if newAttempt.UsedProblem.CFID != usedProblem.CFID {
					continue
				}
				newAttempt.User = user
				newAttempt.UserID = user.ID
				newAttempt.UsedProblem = usedProblem
				newAttempt.UsedProblemID = usedProblem.ID
				err = config.DB.Create(&newAttempt).Error
				if err != nil {
					adminlog.SendMessage(fmt.Sprintf("Error on creating new attempt for user %s: %s", user.CFHandle, err.Error()), config.Ctx, config.B)
				}
				break
			}
		}

		if attemptCountDelta == 0 {
			continue
		}
		user.Rating += float64(userRatingDelta)
		user.SolvedCount += SolvedCountDelta
		user.OKCount += OKCountDelta
		user.AttemptsCount += attemptCountDelta
		user.CountEasy += countEasyDelta
		user.CountMedium += countMediumDelta
		user.CountAdvanced += countAdvancedDelta
		user.CountHard += countHardDelta
		fmt.Printf("Rating delta for %s: %d\n", user.CFHandle, userRatingDelta)
		config.DB.Save(&user)
		if updateCheckedTime {
			config.DB.First(&config.LastCheckedTime)
			config.LastCheckedTime.UnixTime = time.Now().Unix()
			config.DB.Save(&config.LastCheckedTime)
		}
	}
	log.Println("Stats updated successfully")
}

func statsUpdater2() error {
	users := make([]models.User, 0)
	usedProblems := make([]models.UsedProblem, 0)
	newAttempts := make([]models.Attempt, 0)

	err := config.DB.Find(&users).Error
	if err != nil {
		adminlog.SendMessage("Error on fetching all users (stats updater): "+err.Error(), config.Ctx, config.B)
		return errors.New("database error")
	}
	err = config.DB.Find(&usedProblems).Error
	if err != nil {
		adminlog.SendMessage("Error on fetching all used problems (stats updater): "+err.Error(), config.Ctx, config.B)
		return errors.New("database error")
	}

	updateCheckedTime := true
	checkedTime := time.Now().Unix()
	for _, user := range users {
		userNewAttempts := cf.GetLatestAttempts(user.CFHandle)
		for _, attempt := range userNewAttempts {
			ok := false
			for _, usedProblem := range usedProblems {
				if attempt.UsedProblem.CFID == usedProblem.CFID {
					ok = true
					attempt.UsedProblem = usedProblem
					attempt.UsedProblemID = usedProblem.ID
					break
				}
			}
			if !ok {
				continue
			}
			if attempt.Verdict == "TESTING" {
				updateCheckedTime = false
				break
			}

			attempt.User = user
			attempt.UserID = user.ID
			newAttempts = append(newAttempts, attempt)
		}
		if !updateCheckedTime {
			break
		}
	}
	if !updateCheckedTime {
		return errors.New("TESTING")
	}
	config.DB.First(&config.LastCheckedTime)
	config.LastCheckedTime.UnixTime = checkedTime
	config.DB.Save(&config.LastCheckedTime)

	sort.Slice(newAttempts, func(i, j int) bool {
		return newAttempts[i].CreationTime < newAttempts[j].CreationTime
	})

	for _, newAttempt := range newAttempts {
		usedProblemIndex := -1
		for i := 0; i < len(usedProblems); i++ {
			if newAttempt.UsedProblemID == usedProblems[i].ID {
				usedProblemIndex = i
				break
			}
		}
		userIndex := -1
		for i := 0; i < len(users); i++ {
			if newAttempt.UserID == users[i].ID {
				userIndex = i
				break
			}
		}

		usedProblems[usedProblemIndex].AttemptsCount++
		users[userIndex].AttemptsCount++
		if newAttempt.Verdict == "OK" {
			users[userIndex].OKCount++
			usedProblems[usedProblemIndex].OKCount++
		}
		oldAttempts := make([]models.Attempt, 0)
		config.DB.
			Model(&models.Attempt{}).
			Where("user_id = ?", newAttempt.UserID).
			Where("used_problem_id = ?", newAttempt.UsedProblemID).
			Find(&oldAttempts)
		isSolved := false
		for _, oldAttempt := range oldAttempts {
			//log.Println(oldAttempt)
			if oldAttempt.Verdict == "OK" {
				isSolved = true
				//adminlog.SendMessage(fmt.Sprintf("%s solved %s problem before.", users[userIndex].CFHandle, usedProblems[usedProblemIndex].CFID), config.Ctx, config.B)
			}
		}
		if !isSolved && newAttempt.Verdict == "OK" {
			users[userIndex].SolvedCount++
			usedProblems[usedProblemIndex].SolvedCount++
			cfRating := usedProblems[usedProblemIndex].Rating
			if cfRating > 700 && cfRating < 1100 {
				users[userIndex].CountEasy++
			}
			if cfRating > 1000 && cfRating < 1600 {
				users[userIndex].CountMedium++
			}
			if cfRating > 1500 && cfRating < 2100 {
				users[userIndex].CountAdvanced++
			}
			if cfRating > 2000 {
				users[userIndex].CountHard++
			}
		}

		err := config.DB.Create(&newAttempt).Error
		if err != nil {
			adminlog.SendMessage("Error on creating new attempt: "+err.Error(), config.Ctx, config.B)
		}

		if newAttempt.Verdict != "OK" {
			continue
		}
		if newAttempt.UsedProblem.CFID == config.TodaysTasks.Easy.CFID {
			users[userIndex].Rating += config.TodaysTasks.EasyPoint
			config.TodaysTasks.EasyPoint = math.Max(20, config.TodaysTasks.EasyPoint*0.97)
			config.DB.Save(&config.TodaysTasks)
			continue
		}
		if newAttempt.UsedProblem.CFID == config.TodaysTasks.Medium.CFID {
			users[userIndex].Rating += config.TodaysTasks.MediumPoint
			config.TodaysTasks.MediumPoint = math.Max(20, config.TodaysTasks.MediumPoint*0.95)
			config.DB.Save(&config.TodaysTasks)
			continue
		}
		if newAttempt.UsedProblem.CFID == config.TodaysTasks.Advanced.CFID {
			users[userIndex].Rating += config.TodaysTasks.AdvancedPoint
			config.TodaysTasks.AdvancedPoint = math.Max(20, config.TodaysTasks.AdvancedPoint*0.93)
			config.DB.Save(&config.TodaysTasks)
			continue
		}
		if newAttempt.UsedProblem.CFID == config.TodaysTasks.Hard.CFID {
			users[userIndex].Rating += config.TodaysTasks.HardPoint
			config.TodaysTasks.HardPoint = math.Max(20, config.TodaysTasks.HardPoint*0.9)
			config.DB.Save(&config.TodaysTasks)
			continue
		}
		users[userIndex].Rating += 10
	}

	for _, user := range users {
		err := config.DB.Save(&user).Error
		if err != nil {
			adminlog.SendMessage(fmt.Sprintf("Error on updating user (%s) data: %s", user.CFHandle, err.Error()), config.Ctx, config.B)
		}
	}
	for _, problem := range usedProblems {
		err := config.DB.Save(&problem).Error
		if err != nil {
			adminlog.SendMessage(fmt.Sprintf("Error on updating problem (%s) data: %s", problem.CFID, err.Error()), config.Ctx, config.B)
		}
	}

	return nil
}

func updateUsersData() {
	users := make([]models.User, 0)
	err := config.DB.Find(&users).Error
	if err != nil {
		adminlog.SendMessage("Error on finding all users while updating users data: "+err.Error(), config.Ctx, config.B)
		return
	}
	updated := make([]string, 0)
	notUpdated := make([]string, 0)
	for _, user := range users {
		body := utils.HTTPGet(cf.BaseURL + "user.info?handles=" + user.CFHandle)
		if body == nil || len(body) == 0 {
			adminlog.SendMessage("Error getting user info while updating users data: "+user.CFHandle, config.Ctx, config.B)
			notUpdated = append(notUpdated, user.CFHandle)
			continue
		}
		res := cfmodels.UserInfo{}
		err = json.Unmarshal(body, &res)
		if err != nil {
			adminlog.SendMessage("Error while unmarshaling user info: "+user.CFHandle+" "+err.Error(), config.Ctx, config.B)
			notUpdated = append(notUpdated, user.CFHandle)
			continue
		}
		if len(res.Result) == 0 {
			adminlog.SendMessage("No such codeforces user: "+user.CFHandle, config.Ctx, config.B)
			notUpdated = append(notUpdated, user.CFHandle)
			continue
		}
		userInfo := res.Result[0]
		user.CFRating = userInfo.Rating
		user.FirstName = userInfo.FirstName
		user.LastName = userInfo.LastName

		err = config.DB.Save(&user).Error
		if err != nil {
			adminlog.SendMessage("Error while updating users data: "+user.CFHandle+" "+err.Error(), config.Ctx, config.B)
			notUpdated = append(notUpdated, user.CFHandle)
			continue
		}
		updated = append(updated, user.CFHandle)
	}
	updatedStr := strings.Join(updated, ",")
	notUpdatedStr := strings.Join(notUpdated, ",")
	adminlog.SendMessage(fmt.Sprintf("Updated: %s\nNot updated: %s", updatedStr, notUpdatedStr), config.Ctx, config.B)

	data, err := os.ReadFile("sqlite.db")
	if err != nil {
		adminlog.SendMessage("Cannot send database file: "+err.Error(), config.Ctx, config.B)
		return
	}

	_, err = config.B.SendDocument(config.Ctx, &bot.SendDocumentParams{
		ChatID:   adminlog.TGID,
		Document: &botModels.InputFileUpload{Filename: "sqlite.db", Data: bytes.NewReader(data)},
		Caption:  "Database for backup",
	})
	if err != nil {
		adminlog.SendMessage("Cannot send database file: "+err.Error(), config.Ctx, config.B)
	}
}

func iAmDoneHandler(ctx context.Context, b *bot.Bot, update *botModels.Update) {
	err := statsUpdater2()
	if err != nil && err.Error() == "TESTING" {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			ReplyParameters: &botModels.ReplyParameters{
				ChatID:    update.Message.Chat.ID,
				MessageID: update.Message.ID,
			},
			Text: "Ba'zi foydalanuvchilarning yechimlari testlash jarayonida, keyinroq urinib ko'ring.",
		})
		if err != nil {
			adminlog.SendMessage("Error sending message: "+err.Error(), ctx, b)
		}
		return
	}
	if err != nil {
		//adminlog.SendMessage(err.Error(), ctx, b)
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		ReplyParameters: &botModels.ReplyParameters{
			ChatID:    update.Message.Chat.ID,
			MessageID: update.Message.ID,
		},
		Text: "👍",
	})
	if err != nil {
		adminlog.SendMessage("Error sending message: "+err.Error(), ctx, b)
	}
}
