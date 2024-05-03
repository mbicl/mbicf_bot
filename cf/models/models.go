package models

const (
	BaseURL     = "https://codeforces.com/api/"
	ProblemsURL = "https://codeforces.com/problemset/problem/"
)

type Problem struct {
	ContestID int      `json:"contestId"`
	Index     string   `json:"index"`
	Name      string   `json:"name"`
	Link      string   `gorm:"link"`
	Type      string   `json:"type"`
	Tags      []string `json:"tags"`
	Rating    int      `json:"rating,omitempty"`
	Points    float32  `json:"points,omitempty"`
}

type ProblemStatistics struct {
	ContestID   int    `json:"contestId"`
	Index       string `json:"index"`
	SolvedCount int    `json:"solvedCount"`
}

type ProblemSet struct {
	Status string `json:"status"`
	Result struct {
		Problems          []*Problem           `json:"problems"`
		ProblemStatistics []*ProblemStatistics `json:"problemStatistics"`
	} `json:"result"`
}

type UserStatus struct {
	Status string        `json:"status"`
	Result []*Submission `json:"result"`
}

type Submission struct {
	ID                  int   `json:"id"`
	ContestID           int   `json:"contestId"`
	CreationTimeSeconds int64 `json:"creationTimeSeconds"`
	RelativeTimeSeconds int64 `json:"relativeTimeSeconds"`
	Problem             struct {
		ContestID int      `json:"contestId"`
		Index     string   `json:"index"`
		Name      string   `json:"name"`
		Type      string   `json:"type"`
		Points    float32  `json:"points"`
		Rating    int      `json:"rating"`
		Tags      []string `json:"tags"`
	} `json:"problem"`
	Author struct {
		ContestID int `json:"contestId"`
		Members   []struct {
			Handle string `json:"handle"`
		} `json:"members"`
		ParticipantType  string `json:"participantType"`
		Ghost            bool   `json:"ghost"`
		StartTimeSeconds int64  `json:"startTimeSeconds"`
	} `json:"author"`
	ProgrammingLanguage string `json:"programmingLanguage"`
	Verdict             string `json:"verdict"`
	TestSet             string `json:"testset"`
	PassedTestCount     int    `json:"passedTestCount"`
	TimeConsumedMillis  int    `json:"timeConsumedMillis"`
	MemoryConsumedBytes int    `json:"memoryConsumedBytes"`
}

type UserInfo struct {
	Status string `json:"status"`
	Result []User `json:"result"`
}

type User struct {
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Handle     string `json:"handle"`
	TitlePhoto string `json:"titlePhoto"`
	Avatar     string `json:"avatar"`

	Country      string `json:"country"`
	City         string `json:"city"`
	Organization string `json:"organization"`

	LastOnlineTimeSeconds   int64 `json:"lastOnlineTimeSeconds"`
	RegistrationTimeSeconds int64 `json:"registrationTimeSeconds"`

	Rating        int    `json:"rating"`
	MaxRating     int    `json:"maxRating"`
	Rank          string `json:"rank"`
	MaxRank       string `json:"maxRank"`
	FriendOfCount int    `json:"friendOfCount"`
	Contribution  int    `json:"contribution"`
}
