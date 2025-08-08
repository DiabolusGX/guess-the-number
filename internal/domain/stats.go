package domain

import (
	"context"
	"time"
)

type StatsRepository interface {
	GetClosestGuesses(ctx context.Context, gameID string, limit int) ([]GuessAttempt, error)
}

// TimeRange represents a time period for filtering stats
type TimeRange struct {
	Type      string     `json:"type"` // "last-day", "last-week", "last-month", "all-time", "custom"
	StartDate *time.Time `json:"startDate,omitempty"`
	EndDate   *time.Time `json:"endDate,omitempty"`
}

// String returns a string representation of the time range for caching
func (tr TimeRange) String() string {
	if tr.Type == "custom" && tr.StartDate != nil && tr.EndDate != nil {
		return tr.Type + ":" + tr.StartDate.Format("2006-01-02") + ":" + tr.EndDate.Format("2006-01-02")
	}
	return tr.Type
}

// GetTimeRange returns the actual start and end dates for the time range
func (tr TimeRange) GetTimeRange() (time.Time, time.Time) {
	now := time.Now()
	var start, end time.Time

	switch tr.Type {
	case "last-day":
		start = now.AddDate(0, 0, -1)
		end = now
	case "last-week":
		start = now.AddDate(0, 0, -7)
		end = now
	case "last-month":
		start = now.AddDate(0, -1, 0)
		end = now
	case "custom":
		if tr.StartDate != nil {
			start = *tr.StartDate
		}
		if tr.EndDate != nil {
			end = *tr.EndDate
		}
	case "all-time":
		fallthrough
	default:
		// For all-time, use epoch as start
		start = time.Unix(0, 0)
		end = now
	}

	return start, end
}

// StatResult represents the response structure for stats queries
type StatResult struct {
	Type        string    `json:"type"`
	TimeRange   TimeRange `json:"timeRange"`
	Data        any       `json:"data"`
	UserRank    *UserRank `json:"userRank,omitempty"`
	GeneratedAt time.Time `json:"generatedAt"`
}

// UserRank represents a user's rank in a particular stat
type UserRank struct {
	Position int `json:"position"`
	Value    any `json:"value"`
	Total    int `json:"total"`
}

// NumberFrequencyStats represents how often a number was guessed
type NumberFrequencyStats struct {
	Number int64 `json:"number"`
	Count  int64 `json:"count"`
}

// UserGuessStats represents guess statistics for a user
type UserGuessStats struct {
	UserID        string `json:"userId"`
	UniqueGuesses int64  `json:"uniqueGuesses"`
	TotalGuesses  int64  `json:"totalGuesses"`
}

// WinnerStats represents win statistics for users
type WinnerStats struct {
	UserID string `json:"userId"`
	Wins   int64  `json:"wins"`
	Points int64  `json:"points"`
}
