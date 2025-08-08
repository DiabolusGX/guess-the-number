package types

import "github.com/diabolusgx/guess-the-number/internal/domain"

type GetClosestGuessesRequest struct {
	RequesterUserID string `json:"requester_user_id" validate:"required"`
	GuildID         string `json:"guild_id" validate:"required"`
	GameID          string `json:"game_id" validate:"required"`
	IsUserMod       bool   `json:"is_user_mod"`
}

type GetClosestGuessesResponse struct {
	Game    *domain.Game           `json:"game"`
	Guesses []*domain.GuessAttempt `json:"guesses"`
}

type GetTopGuessedNumbersRequest struct {
	RequesterUserID string           `json:"requester_user_id"`
	GameID          string           `json:"game_id"`
	GuildID         string           `json:"guild_id"`
	TimeRange       domain.TimeRange `json:"time_range"`
}

type GetTopGuessedNumbersResponse struct {
	Game    *domain.Game                   `json:"game"`
	Numbers []*domain.NumberFrequencyStats `json:"numbers"`
}

type GetTopGuessersRequest struct {
	RequesterUserID string           `json:"requester_user_id"`
	GuildID         string           `json:"guild_id"`
	GameID          string           `json:"game_id"`
	TimeRange       domain.TimeRange `json:"time_range"`
}

type GetTopGuessersResponse struct {
	Game        *domain.Game             `json:"game"`
	TopGuessers []*domain.UserGuessStats `json:"top_guessers"`
}

type GetTopWinnersRequest struct {
	RequesterUserID string `json:"requester_user_id"`
	GuildID         string `json:"guild_id"`
	ByGame          bool   `json:"by_game"`
	ByPoints        bool   `json:"by_points"`
}

type GetTopWinnersResponse struct {
	Winners []*domain.WinnerStats `json:"winners"`
}
