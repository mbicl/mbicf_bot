package models

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	FirstName  string
	LastName   string
	CFHandle   string `gorm:"unique"`
	CFRating   int    `gorm:"default:0"`
	TGUserName string `gorm:"unique"`
	TGUserID   int64  `gorm:"unique"`

	AttemptsCount int `gorm:"default:0"`
	OKCount       int `gorm:"default:0"`
	SolvedCount   int `gorm:"default:0"`
	Rating        int `gorm:"default:0"`
	CountEasy     int `gorm:"default:0"`
	CountMedium   int `gorm:"default:0"`
	CountAdvanced int `gorm:"default:0"`
	CountHard     int `gorm:"default:0"`
}

type Problem struct {
	gorm.Model
	ProblemID string `gorm:"unique"`
	Link      string
	Name      string
	Tags      pq.StringArray `gorm:"type:text[]"`
	Rating    int
	Points    float32
}

type UsedProblem struct {
	gorm.Model
	ProblemID      string `gorm:"unique"`
	Link           string
	Name           string
	Tags           pq.StringArray `gorm:"type:text[]"`
	Rating         int
	Points         float32
	AttemptsCount  int
	SolvedCount    int
	OKCount        int
	AttemptedUsers []*User `gorm:"many2many:user_attempted_problems;"`
	SolvedUsers    []*User `gorm:"many2many:user_solved_problems;"`
}

type DailyTasks struct {
	Easy     Problem
	Medium   Problem
	Advanced Problem
	Hard     Problem
}

type LastCheckedTime struct {
	gorm.Model
	UnixTime int64 `gorm:"default:0"`
}
