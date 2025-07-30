package types

import (
	"github.com/diabolusgx/guess-the-number-go/internal/domain"
)

type CreateGameRequest struct {
	GuildID           string `json:"guild_id" validate:"required"`
	ChannelID         string `json:"channel_id" validate:"required"`
	CreatedBy         string `json:"created_by" validate:"required"`
	LowerBound        int64  `json:"lower_bound"`
	UpperBound        int64  `json:"upper_bound" validate:"required,min=1"`
	AutoReactionHints bool   `json:"auto_reaction_hints,omitempty"`
}

type CreateGameResponse struct {
	Game *domain.Game `json:"game"`
}

type HandleAttemptRequest struct {
	ChannelID string `json:"channel_id" validate:"required"`
	MessageID string `json:"message_id" validate:"required"`
	UserID    string `json:"user_id" validate:"required"`
	Guess     int64  `json:"guess" validate:"required,min=0"`
}

type HandleAttemptResponse struct {
	Correct bool         `json:"correct"`
	Game    *domain.Game `json:"game"`
}

type FinishGameRequest struct {
	ChannelID string `json:"channel_id" validate:"required"`
	Guesses   int64  `json:"guesses"`
	MessageID string `json:"message_id"`
	WonBy     string `json:"won_by"`
}

type FinishGameResponse struct {
	Game *domain.Game `json:"game"`
}

type HintType string

func (h HintType) String() string {
	return string(h)
}

const (
	HintTypeNumber     HintType = "compare-number"
	HintTypeFirstDigit HintType = "first-digit"
	HintTypeLastDigit  HintType = "last-digit"
)

type GetHintRequest struct {
	ChannelID string   `json:"channel_id" validate:"required"`
	HintType  HintType `json:"hint_type" validate:"required"`
	CompareTo int64    `json:"compare_to"`
}

type GetHintResponse struct {
	Hint string `json:"hint"`
}

type GetGameInfoRequest struct {
	ChannelID string `json:"channel_id" validate:"required"`
}

type GetGameInfoResponse struct {
	Game *domain.Game `json:"game"`
}
