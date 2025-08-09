package domain

import (
	"time"
)

// NumberFrequencyStats represents how often a number was guessed
type NumberFrequencyStats struct {
	Number int64
	Count  int64
}

// UserGuessStats represents guess statistics for a user
type UserGuessStats struct {
	UserID        string
	UniqueGuesses int64
	TotalGuesses  int64
}

// WinnerStats represents win statistics for users
type WinnerStats struct {
	UserID string
	Wins   int64
	Points int64
}

// TimeRange represents a time period for filtering stats
type TimeRange struct {
	Type      string
	StartDate *time.Time
	EndDate   *time.Time
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
