package cf

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"

	"github.com/mbicl/mbicf_bot/adminlog"
	cfmodels "github.com/mbicl/mbicf_bot/cf/models"
	"github.com/mbicl/mbicf_bot/config"
	"github.com/mbicl/mbicf_bot/models"
	"github.com/mbicl/mbicf_bot/utils"
)

func UserAttemptStats(ProblemID, CFHandle string) (int, int) {
	userStatus := cfmodels.UserStatus{}
	if _, ok := config.UserStatusMap[CFHandle]; !ok {
		body := utils.HTTPGet(fmt.Sprintf("%suser.status?handle=%s&from=1&count=111", cfmodels.BaseURL, CFHandle))
		err := json.Unmarshal(body, &userStatus)
		if err != nil {
			adminlog.SendMessage("Error unmarshalling status for CF handle "+CFHandle, config.Ctx, config.B)
			return 0, 0
		}
		config.UserStatusMap[CFHandle] = userStatus
	} else {
		userStatus = config.UserStatusMap[CFHandle]
	}

	attemptCount, OK := 0, 0
	for _, i := range userStatus.Result {
		if i.CreationTimeSeconds < config.LastCheckedTime.UnixTime {
			break
		}
		if strconv.Itoa(i.Problem.ContestID)+i.Problem.Index == ProblemID {
			attemptCount++
			if i.Verdict == "OK" {
				OK++
			}
		}
	}
	return attemptCount, OK
}

func IsUsed(ProblemID string) bool {
	problem := models.UsedProblem{}
	config.DB.Where("problem_id = ?", ProblemID).First(&problem)
	if problem.Link != "" {
		return true
	}
	return false
}

func GetAllProblems() {
	body := utils.HTTPGet(cfmodels.BaseURL + "problemset.problems")
	problemSet := cfmodels.ProblemSet{}
	err := json.Unmarshal(body, &problemSet)
	if err != nil {
		adminlog.Fatal("Error unmarshalling problems (GetAllProblems)."+err.Error(), config.Ctx, config.B)
	}
	for _, i := range problemSet.Result.Problems {
		special := false
		for _, k := range i.Tags {
			if k == "*special" {
				special = true
			}
		}
		if special {
			continue
		}
		problem := models.Problem{}
		pid := strconv.Itoa(i.ContestID) + i.Index
		config.DB.Where("problem_id = ?", pid).First(&problem)
		if problem.Link != "" {
			continue
		}
		problem.ProblemID = pid
		problem.Link = cfmodels.ProblemsURL + strconv.Itoa(i.ContestID) + "/" + i.Index + "/"
		problem.Name = i.Name
		problem.Rating = i.Rating
		problem.Points = i.Points
		for _, t := range i.Tags {
			problem.Tags = append(problem.Tags, t)
		}
		log.Println(problem.ProblemID)
		config.DB.Save(&problem)
		log.Println("Problem added: " + problem.ProblemID)
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

	config.TodaysTasks.Easy = easyProblems[n]
	config.TodaysTasks.Medium = mediumProblems[n]
	config.TodaysTasks.Advanced = advancedProblems[n]
	config.TodaysTasks.Hard = hardProblems[n]

	config.DB.Delete(&config.TodaysTasks.Easy)
	config.DB.Delete(&config.TodaysTasks.Medium)
	config.DB.Delete(&config.TodaysTasks.Advanced)
	config.DB.Delete(&config.TodaysTasks.Hard)

	config.DB.Save(&models.UsedProblem{
		ProblemID:      config.TodaysTasks.Easy.ProblemID,
		Link:           config.TodaysTasks.Easy.Link,
		Name:           config.TodaysTasks.Easy.Name,
		Tags:           config.TodaysTasks.Easy.Tags,
		Rating:         config.TodaysTasks.Easy.Rating,
		Points:         config.TodaysTasks.Easy.Points,
		AttemptsCount:  0,
		SolvedCount:    0,
		AttemptedUsers: []*models.User{},
		SolvedUsers:    []*models.User{},
	})
	config.DB.Save(&models.UsedProblem{
		ProblemID:      config.TodaysTasks.Medium.ProblemID,
		Link:           config.TodaysTasks.Medium.Link,
		Name:           config.TodaysTasks.Medium.Name,
		Tags:           config.TodaysTasks.Medium.Tags,
		Rating:         config.TodaysTasks.Medium.Rating,
		Points:         config.TodaysTasks.Medium.Points,
		AttemptsCount:  0,
		SolvedCount:    0,
		AttemptedUsers: []*models.User{},
		SolvedUsers:    []*models.User{},
	})
	config.DB.Save(&models.UsedProblem{
		ProblemID:      config.TodaysTasks.Advanced.ProblemID,
		Link:           config.TodaysTasks.Advanced.Link,
		Name:           config.TodaysTasks.Advanced.Name,
		Tags:           config.TodaysTasks.Advanced.Tags,
		Rating:         config.TodaysTasks.Advanced.Rating,
		Points:         config.TodaysTasks.Advanced.Points,
		AttemptsCount:  0,
		SolvedCount:    0,
		AttemptedUsers: []*models.User{},
		SolvedUsers:    []*models.User{},
	})
	config.DB.Save(&models.UsedProblem{
		ProblemID:      config.TodaysTasks.Hard.ProblemID,
		Link:           config.TodaysTasks.Hard.Link,
		Name:           config.TodaysTasks.Hard.Name,
		Tags:           config.TodaysTasks.Hard.Tags,
		Rating:         config.TodaysTasks.Hard.Rating,
		Points:         config.TodaysTasks.Hard.Points,
		AttemptsCount:  0,
		SolvedCount:    0,
		AttemptedUsers: []*models.User{},
		SolvedUsers:    []*models.User{},
	})
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
