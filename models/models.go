package models

import (
	"fmt"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	FirstName  string
	LastName   string
	CFHandle   string
	CFRating   int
	TGUserName string
	TGUserID   int64 `gorm:"unique"`

	AttemptsCount int     `gorm:"default:0"`
	OKCount       int     `gorm:"default:0"`
	SolvedCount   int     `gorm:"default:0"`
	Rating        float64 `gorm:"default:0"`
	CountEasy     int     `gorm:"default:0"`
	CountMedium   int     `gorm:"default:0"`
	CountAdvanced int     `gorm:"default:0"`
	CountHard     int     `gorm:"default:0"`
}

func (u User) String() string {
	res := fmt.Sprintf(
		"User{\n\tID: %d\n\tName: %s\n\tCF Handle: %s\n\tCF Rating: %d\n\tTG Username: %s\n\tAttempts: %d\n\tAccepted: %d\n\tSolved: %d\n\tRating: %f\n}",
		u.ID,
		u.FirstName+" "+u.LastName,
		u.CFHandle,
		u.CFRating,
		u.TGUserName,
		u.AttemptsCount,
		u.OKCount,
		u.SolvedCount,
		u.Rating,
	)
	return res
}

type Attempt struct {
	gorm.Model
	User          User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:UserID"`
	UserID        uint
	UsedProblem   UsedProblem `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:UsedProblemID"`
	UsedProblemID uint
	Verdict       string
	CreationTime  int64
}

func (a Attempt) String() string {
	res := fmt.Sprintf(
		"Attempt{\n\tUserID: %d\n\tCF Handle: %s\n\tUsedProblemID: %d\n\tUsedProblem CFID: %s\n\tVerdict: %s\n}",
		a.UserID,
		a.User.CFHandle,
		a.UsedProblemID,
		a.UsedProblem.CFID,
		a.Verdict,
	)
	return res
}

type Problem struct {
	gorm.Model
	CFID   string `gorm:"unique"`
	Link   string
	Name   string
	Tags   pq.StringArray `gorm:"type:text[]"`
	Rating int
	Points float32
}

func (p Problem) String() string {
	res := fmt.Sprintf(
		"Problem{\n\tID: %d\n\tCF ID: %s\n\tLink: %s\n\tName: %s\n\tRating: %d\n\tPoints: %f\n}",
		p.ID,
		p.CFID,
		p.Link,
		p.Name,
		p.Rating,
		p.Points,
	)
	return res
}

type UsedProblem struct {
	gorm.Model
	CFID          string `gorm:"unique"`
	Link          string
	Name          string
	Tags          pq.StringArray `gorm:"type:text[]"`
	Rating        int
	Points        float32
	AttemptsCount int
	SolvedCount   int
	OKCount       int
}

func (u UsedProblem) String() string {
	res := fmt.Sprintf(
		"UsedProblem{\n\tID: %d\n\tCF ID: %s\n\tLink: %s\n\tName: %s\n\tRating: %d\n\tPoints: %f\n\tAttempts: %d\n\tAccepted: %d\n\tSolved: %d\n}",
		u.ID,
		u.CFID,
		u.Link,
		u.Name,
		u.Rating,
		u.Points,
		u.AttemptsCount,
		u.OKCount,
		u.SolvedCount,
	)
	return res
}

type DailyTasks struct {
	gorm.Model
	Easy          Problem `gorm:"foreignKey:EasyID"`
	EasyID        uint
	EasyPoint     float64
	Medium        Problem `gorm:"foreignKey:MediumID"`
	MediumID      uint
	MediumPoint   float64
	Advanced      Problem `gorm:"foreignKey:AdvancedID"`
	AdvancedID    uint
	AdvancedPoint float64
	Hard          Problem `gorm:"foreignKey:HardID"`
	HardID        uint
	HardPoint     float64
}

func (d DailyTasks) String() string {
	res := fmt.Sprintf(
		"Daily Tasks {\n\tEasy: %s(%d)\n\tMedium: %s(%d)\n\tAdvanced: %s(%d)\n\tHard: %s(%d)\n}\n",
		d.Easy.CFID, d.EasyID,
		d.Medium.CFID, d.MediumID,
		d.Advanced.CFID, d.AdvancedID,
		d.Hard.CFID, d.HardID,
	)
	return res
}

type LastCheckedTime struct {
	gorm.Model
	UnixTime int64 `gorm:"default:0"`
}
