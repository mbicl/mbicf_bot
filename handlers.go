package main

import (
	"context"
	"encoding/json"
	"fmt"
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
		userStatus := cfmodels.UserStatus{}
		err = json.Unmarshal(body, &userStatus)
		if err != nil {
			adminlog.SendMessage("Error unmarshalling user status(userRegistration):"+err.Error(), ctx, b)
			continue
		}
		submission := userStatus.Result[0]
		if strconv.Itoa(submission.Problem.ContestID)+submission.Problem.Index == problem.ProblemID &&
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
	if len(msg) == 0 {
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
	sent := false

	for {
		time.Sleep(60 * time.Second)
		if sent || time.Now().Hour() != 8 {
			continue
		}
		if time.Now().Hour() == 0 {
			sent = false
		}

		cf.UpdateTodaysTasks()
		_, err := b.SendChatAction(ctx, &bot.SendChatActionParams{
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
		sent = true
	}
}

func statsUpdater() {
	for {
		time.Sleep(100 * time.Second)

		users := make([]models.User, 0)
		config.DB.Find(&users)
		usedProblems := make([]models.UsedProblem, 0)

		config.DB.Find(&usedProblems)
		config.DB.First(&config.LastCheckedTime)
		config.LastCheckedTime.UnixTime = time.Now().Unix()
		config.DB.Save(&config.LastCheckedTime)

		config.UserStatusMap = make(map[string]cfmodels.UserStatus)

		for _, user := range users {
			userRatingDelta := 0
			countEasyDelta := 0
			countMediumDelta := 0
			countAdvancedDelta := 0
			countHardDelta := 0
			attemptCountDelta := 0
			OKCountDelta := 0
			SolvedCountDelta := 0

			for _, usedProblem := range usedProblems {
				attemptCount, OK := cf.UserAttemptStats(usedProblem.ProblemID, user.CFHandle)
				if attemptCount == 0 {
					continue
				}

				attemptCountDelta += attemptCount
				OKCountDelta += OK

				usedProblem.AttemptsCount += attemptCount
				usedProblem.OKCount += OK
				isAttempted := false
				isSolved := false
				for _, i := range usedProblem.AttemptedUsers {
					if i.CFHandle == user.CFHandle {
						isAttempted = true
					}
				}
				for _, i := range usedProblem.SolvedUsers {
					if i.CFHandle == user.CFHandle {
						isSolved = true
					}
				}
				if !isAttempted {
					usedProblem.AttemptedUsers = append(usedProblem.AttemptedUsers, &user)
				}
				if OK > 0 && !isSolved {
					usedProblem.SolvedUsers = append(usedProblem.SolvedUsers, &user)
					SolvedCountDelta++
				}

				config.DB.Save(&usedProblem)

				if OK == 0 {
					continue
				}
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
				if usedProblem.ProblemID == config.TodaysTasks.Easy.ProblemID ||
					usedProblem.ProblemID == config.TodaysTasks.Medium.ProblemID ||
					usedProblem.ProblemID == config.TodaysTasks.Advanced.ProblemID ||
					usedProblem.ProblemID == config.TodaysTasks.Hard.ProblemID {
					userRatingDelta += max(1, (usedProblem.Rating-user.Rating+50)/100*2)
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
			config.DB.Save(&user)
		}
	}
}
