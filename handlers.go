package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

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
		Text:   "Assalomu alaykum",
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
	body := utils.HTTPGet(cfmodels.BaseURL + "user.info?handles=" + handle)
	userInfo := cfmodels.UserInfo{}
	err := json.Unmarshal(body, &userInfo)
	if err != nil {
		adminlog.SendMessage("Error unmarshalling user info: "+err.Error(), ctx, b)
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
		body := utils.HTTPGet(cfmodels.BaseURL + "user.status?handle=" + user.Handle + "&from=1&count=1")
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
		rating = rating / 100 * 100
		if rating < 800 || rating > 3500 {
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				ReplyParameters: &botModels.ReplyParameters{
					ChatID:    update.Message.Chat.ID,
					MessageID: update.Message.ID,
				},
				Text: "Reyting [800,3500] oraliqda bo'lishi kerak",
			})
			if err != nil {
				adminlog.SendMessage("Error sending message: "+err.Error(), ctx, b)
			}
			return
		}
		problem := cf.GetRandomProblemWithRating(rating)
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			ReplyParameters: &botModels.ReplyParameters{
				ChatID:    update.Message.Chat.ID,
				MessageID: update.Message.ID,
			},
			Text: "@" + update.Message.Chat.Username + " uchun masala: " + problem.Link,
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
		rating = rating / 100 * 100
		if rating < 800 || rating > 3500 {
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				ReplyParameters: &botModels.ReplyParameters{
					ChatID:    update.Message.Chat.ID,
					MessageID: update.Message.ID,
				},
				Text: "Reyting [800,3500] oraliqda bo'lishi kerak",
			})
			if err != nil {
				adminlog.SendMessage("Error sending message: "+err.Error(), ctx, b)
			}
			return
		}
		problem := cf.GetRandomProblemWithRating(rating)
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			ReplyParameters: &botModels.ReplyParameters{
				ChatID:    update.Message.Chat.ID,
				MessageID: update.Message.ID,
			},
			Text: "@" + update.Message.Chat.Username + " uchun masala: " + problem.Link,
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
		rating = rating / 100 * 100
		if rating < 800 || rating > 3500 {
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				ReplyParameters: &botModels.ReplyParameters{
					ChatID:    update.Message.Chat.ID,
					MessageID: update.Message.ID,
				},
				Text: "Reyting [800,3500] oraliqda bo'lishi kerak",
			})
			if err != nil {
				adminlog.SendMessage("Error sending message: "+err.Error(), ctx, b)
			}
			return
		}
		problem := cf.GetRandomProblemWithRating(rating)
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			ReplyParameters: &botModels.ReplyParameters{
				ChatID:    update.Message.Chat.ID,
				MessageID: update.Message.ID,
			},
			Text: "@" + update.Message.Chat.Username + " uchun masala: " + problem.Link,
		})
		if err != nil {
			adminlog.SendMessage("Error sending message: "+err.Error(), ctx, b)
		}
	}
}

func standingsHandler(ctx context.Context, b *bot.Bot, update *botModels.Update) {
	users := make([]models.User, 0)
	config.DB.Order("rating desc, cf_rating desc").Limit(20).Find(&users)
	msg := "Standings:\n"

	for i, user := range users {
		msg += fmt.Sprintf("%2d. %10s (%10s) - %4d (%4d)\n", i, user.FirstName, user.CFHandle, user.Rating, user.CFRating)
	}
	if msg == "Standings:\n" {
		msg = "Hozircha bo'sh."
	}
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   msg,
	})
	if err != nil {
		adminlog.SendMessage("Error sending message: "+err.Error(), ctx, b)
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
		config.TodaysTasks.Medium.Link,
		config.TodaysTasks.Medium.Name,
		config.TodaysTasks.Advanced.Link,
		config.TodaysTasks.Advanced.Name,
		config.TodaysTasks.Hard.Link,
		config.TodaysTasks.Hard.Name,
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
	//config.UserStatusMap = make(map[string]cfmodels.UserStatus)

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
		user.Rating += userRatingDelta
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
		body := utils.HTTPGet(cfmodels.BaseURL + "user.info?handles=" + user.CFHandle)
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
}
