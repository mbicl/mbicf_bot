package cf

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"

	"github.com/mbicl/mbicf_bot/adminlog"
	cfmodels "github.com/mbicl/mbicf_bot/cf/models"
	"github.com/mbicl/mbicf_bot/config"
	"github.com/mbicl/mbicf_bot/models"
	"github.com/mbicl/mbicf_bot/utils"
)

func GetLatestAttempts(CFHandle string) []models.Attempt {
	userStatus := cfmodels.UserStatus{}
	body := utils.HTTPGet(fmt.Sprintf("%suser.status?handle=%s&from=1&count=111", BaseURL, CFHandle))
	if len(body) == 0 || body == nil {
		adminlog.SendMessage(fmt.Sprintf("Cannot get user %s's status.", CFHandle), config.Ctx, config.B)
		return []models.Attempt{}
	}
	err := json.Unmarshal(body, &userStatus)
	if err != nil {
		adminlog.SendMessage("Error unmarshalling status for CF handle "+CFHandle, config.Ctx, config.B)
		return []models.Attempt{}
	}

	newAttempts := make([]models.Attempt, 0)
	for _, i := range userStatus.Result {
		if i.CreationTimeSeconds < config.LastCheckedTime.UnixTime {
			break
		}
		newAttempts = append(newAttempts, models.Attempt{
			Verdict:      i.Verdict,
			User:         models.User{CFHandle: CFHandle},
			UsedProblem:  models.UsedProblem{CFID: strconv.Itoa(i.Problem.ContestID) + i.Problem.Index},
			CreationTime: i.CreationTimeSeconds,
		})
	}
	return newAttempts
}

func IsUsed(ProblemID string) bool {
	problem := models.UsedProblem{}
	config.DB.Where("cf_id = ?", ProblemID).First(&problem)
	if problem.Link != "" {
		return true
	}
	return false
}

func GetAllProblems() {
	body := utils.HTTPGet(BaseURL + "problemset.problems")
	problemSet := cfmodels.ProblemSet{}
	err := json.Unmarshal(body, &problemSet)
	if err != nil {
		log.Println(string(body))
		adminlog.Fatal("Error unmarshalling problems (GetAllProblems)."+err.Error(), config.Ctx, config.B)
	}
	for _, i := range problemSet.Result.Problems {
		special := false
		for _, k := range i.Tags {
			if strings.Contains(k, "special") {
				special = true
			}
		}
		if special {
			continue
		}
		problem := models.Problem{}
		pid := strconv.Itoa(i.ContestID) + i.Index
		config.DB.Where("cf_id = ?", pid).First(&problem)
		if problem.Link != "" {
			continue
		}
		usedProblem := models.UsedProblem{}
		config.DB.Where("cf_id = ?", pid).First(&usedProblem)
		if usedProblem.Link != "" {
			continue
		}
		problem.CFID = pid
		problem.Link = fmt.Sprintf("https://codeforces.com/contest/%d/problem/%s", i.ContestID, i.Index)
		problem.Name = i.Name
		problem.Rating = i.Rating
		problem.Points = i.Points
		for _, t := range i.Tags {
			problem.Tags = append(problem.Tags, t)
		}
		log.Println(problem.CFID)
		config.DB.Save(&problem)
		log.Println("Problem added: " + problem.CFID)
	}
}

func UpdateTodaysTasks() {
	easyProblems := make([]models.Problem, 0)
	mediumProblems := make([]models.Problem, 0)
	advancedProblems := make([]models.Problem, 0)
	hardProblems := make([]models.Problem, 0)

	config.DB.Where("rating >  700").Where("rating < 1100").Find(&easyProblems)
	config.DB.Where("rating > 1000").Where("rating < 1600").Find(&mediumProblems)
	config.DB.Where("rating > 1500").Where("rating < 2100").Find(&advancedProblems)
	config.DB.Where("rating > 2000").Where("rating < 2700").Find(&hardProblems)
	log.Println("Count problems (e,m,a,h): ", len(easyProblems), len(mediumProblems), len(advancedProblems), len(hardProblems))

	rand.Shuffle(len(easyProblems), func(i, j int) {
		easyProblems[i], easyProblems[j] = easyProblems[j], easyProblems[i]
	})
	rand.Shuffle(len(mediumProblems), func(i, j int) {
		mediumProblems[i], mediumProblems[j] = mediumProblems[j], mediumProblems[i]
	})
	rand.Shuffle(len(advancedProblems), func(i, j int) {
		advancedProblems[i], advancedProblems[j] = advancedProblems[j], advancedProblems[i]
	})
	rand.Shuffle(len(hardProblems), func(i, j int) {
		hardProblems[i], hardProblems[j] = hardProblems[j], hardProblems[i]
	})

	n := rand.Intn(min(len(easyProblems), len(mediumProblems), len(advancedProblems), len(hardProblems)))

	config.TodaysTasks.Easy = models.UsedProblem{
		CFID:   easyProblems[n].CFID,
		Link:   easyProblems[n].Link,
		Name:   easyProblems[n].Name,
		Rating: easyProblems[n].Rating,
		Points: easyProblems[n].Points,
		Tags:   easyProblems[n].Tags,
	}
	config.TodaysTasks.Medium = models.UsedProblem{
		CFID:   mediumProblems[n].CFID,
		Link:   mediumProblems[n].Link,
		Name:   mediumProblems[n].Name,
		Rating: mediumProblems[n].Rating,
		Points: mediumProblems[n].Points,
		Tags:   mediumProblems[n].Tags,
	}
	config.TodaysTasks.Advanced = models.UsedProblem{
		CFID:   advancedProblems[n].CFID,
		Link:   advancedProblems[n].Link,
		Name:   advancedProblems[n].Name,
		Rating: advancedProblems[n].Rating,
		Points: advancedProblems[n].Points,
		Tags:   advancedProblems[n].Tags,
	}
	config.TodaysTasks.Hard = models.UsedProblem{
		CFID:   hardProblems[n].CFID,
		Link:   hardProblems[n].Link,
		Name:   hardProblems[n].Name,
		Rating: hardProblems[n].Rating,
		Points: hardProblems[n].Points,
		Tags:   hardProblems[n].Tags,
	}
	config.TodaysTasks.EasyPoint = 100
	config.TodaysTasks.MediumPoint = 100
	config.TodaysTasks.AdvancedPoint = 100
	config.TodaysTasks.HardPoint = 100

	config.DB.Delete(&easyProblems[n])
	config.DB.Delete(&mediumProblems[n])
	config.DB.Delete(&advancedProblems[n])
	config.DB.Delete(&hardProblems[n])

	config.DB.Save(&config.TodaysTasks.Easy)
	config.DB.Save(&config.TodaysTasks.Medium)
	config.DB.Save(&config.TodaysTasks.Advanced)
	config.DB.Save(&config.TodaysTasks.Hard)
}

func GetRandomProblem() *models.Problem {
	problems := make([]models.Problem, 0)
	config.DB.Find(&problems)
	n := rand.Intn(len(problems))
	return &problems[n]
}

func GetRandomProblemWithRating(rating int) *models.Problem {
	problems := make([]models.Problem, 0)
	config.DB.Where("rating = ?", rating).Find(&problems)
	n := rand.Intn(len(problems))
	return &problems[n]
}
