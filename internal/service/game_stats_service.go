package service

import (
	"context"
	"errors"
	"slices"
	"sort"

	"github.com/diabolusgx/guess-the-number/internal/domain"
	"github.com/diabolusgx/guess-the-number/internal/types"
	ierr "github.com/diabolusgx/guess-the-number/pkg/errors"
	"github.com/diabolusgx/guess-the-number/pkg/logger"
	"github.com/go-playground/validator/v10"
)

// StatsService handles all statistics queries with caching
type StatsService interface {
	GetClosestGuesses(ctx context.Context, req *types.GetClosestGuessesRequest) (*types.GetClosestGuessesResponse, error)
	GetTopGuessedNumbers(ctx context.Context, req *types.GetTopGuessedNumbersRequest) (*types.GetTopGuessedNumbersResponse, error)
	GetTopGuessers(ctx context.Context, req *types.GetTopGuessersRequest) (*types.GetTopGuessersResponse, error)
	GetTopWinners(ctx context.Context, req *types.GetTopWinnersRequest) (*types.GetTopWinnersResponse, error)
}

type statsService struct {
	logger *logger.Logger

	redisStatsRepo domain.RedisStatsRepository
	gameStatsRepo  domain.GameStatsRepository
	gameRepo       domain.GameRepository
	guildDataRepo  domain.GuildDataRepository

	gameService GameService
}

const (
	topLimit = 10
)

func NewStatsService(params ServiceParams, gameService GameService) StatsService {
	return &statsService{
		logger: params.Logger,

		redisStatsRepo: params.RedisStatsRepo,
		gameStatsRepo:  params.GameStatsRepo,
		gameRepo:       params.GameRepo,
		guildDataRepo:  params.GuildDataRepo,

		gameService: gameService,
	}
}

// GetClosestGuesses retrieves closest guesses for a game with permission checks
func (s *statsService) GetClosestGuesses(ctx context.Context, req *types.GetClosestGuessesRequest) (*types.GetClosestGuessesResponse, error) {
	// Validate inputs
	if err := validator.New().Struct(req); err != nil {
		return nil, ierr.NewErrorWithContext(ctx, ierr.ErrCodeValidation, err).WithMessage("Make sure GameID & GuildID are set properly!")
	}

	// Get game info to check if finished
	gameResponse, err := s.gameService.GetGameInfo(ctx, &types.GetGameInfoRequest{GameID: req.GameID})
	if err != nil {
		return nil, err
	}
	if gameResponse == nil || gameResponse.Game == nil {
		return nil, ierr.New(ierr.ErrCodeNotFound, "Game not found")
	}

	game := gameResponse.Game
	guesses := make([]*domain.GuessAttempt, 0)

	// Check permissions - closest guesses only available after game completion OR for mods
	if !game.Finished && !req.IsUserMod {
		return nil, ierr.New(ierr.ErrCodePermissionDenied, "Only moderators or bot-managers can view closest guesses for games that are not finished yet!")
	}

	// Try Redis first (for active/recent games)
	redisGuesses, err := s.redisStatsRepo.GetGameClosestGuesses(ctx, game, topLimit)
	if err != nil {
		s.logger.FromContext(ctx).Errorw("failed to get game closest guesses from redis", "error", err)
	} else if len(redisGuesses) > 0 {
		guesses = append(guesses, redisGuesses...)
	}

	// Fall back to MongoDB for completed games
	mongoGuesses, err := s.gameStatsRepo.GetClosestGuesses(ctx, req.GuildID, req.GameID, topLimit)
	if err != nil {
		return nil, err
	}

	guesses = append(guesses, mongoGuesses...)

	// deduplicate guesses by userID & guess
	deduplicatedGuesses := make([]*domain.GuessAttempt, 0, len(guesses))
	for _, guess := range guesses {
		if !slices.ContainsFunc(deduplicatedGuesses, func(g *domain.GuessAttempt) bool {
			return g.UserID == guess.UserID && g.Guess == guess.Guess
		}) {
			deduplicatedGuesses = append(deduplicatedGuesses, guess)
		}
	}

	// sort guesses by distance and send topLimit
	sort.Slice(guesses, func(i, j int) bool {
		return guesses[i].Distance < guesses[j].Distance
	})
	if len(guesses) > topLimit {
		guesses = guesses[:topLimit]
	}

	return &types.GetClosestGuessesResponse{
		Game:    game,
		Guesses: deduplicatedGuesses,
	}, nil
}

// GetTopGuessedNumbers retrieves most frequently guessed numbers
func (s *statsService) GetTopGuessedNumbers(ctx context.Context, req *types.GetTopGuessedNumbersRequest) (*types.GetTopGuessedNumbersResponse, error) {
	// Validate inputs
	if req.GameID == "" && req.GuildID == "" {
		return nil, ierr.NewErrorWithContext(ctx, ierr.ErrCodeValidation, errors.New("game_id or guild_id is required"))
	}

	redisStats := make([]*domain.NumberFrequencyStats, 0)
	var game *domain.Game
	var err error

	if req.GameID != "" {
		// Game-specific stats from Redis - need to get game object first
		game, err = s.gameRepo.GetByID(ctx, req.GameID)
		if err != nil {
			return nil, ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, err).WithMessage("Failed to get game. Please try again later")
		}
		if game == nil {
			return nil, ierr.New(ierr.ErrCodeNotFound, "game not found")
		}

		redisStats, err = s.redisStatsRepo.GetGameTopNumbers(ctx, game, topLimit)
		if err != nil {
			return nil, ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, err).WithMessage("Failed to get game top numbers. Please try again later")
		}
	}

	mongoStats, err := s.gameStatsRepo.GetTopGuessedNumbers(ctx, req.GuildID, req.GameID, req.TimeRange, topLimit)
	if err != nil {
		return nil, ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, err).WithMessage("Failed to get top guessed numbers. Please try again later")
	}

	stats := make([]*domain.NumberFrequencyStats, 0, len(redisStats)+len(mongoStats))
	stats = append(stats, redisStats...)
	stats = append(stats, mongoStats...)

	// deduplicate stats by number
	deduplicatedStats := make([]*domain.NumberFrequencyStats, 0, len(stats))
	for _, stat := range stats {
		if !slices.ContainsFunc(deduplicatedStats, func(s *domain.NumberFrequencyStats) bool {
			return s.Number == stat.Number
		}) {
			deduplicatedStats = append(deduplicatedStats, stat)
		}
	}

	// sort stats by count
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})
	if len(stats) > topLimit {
		stats = stats[:topLimit]
	}

	result := &types.GetTopGuessedNumbersResponse{
		Game:    game,
		Numbers: stats,
	}

	return result, nil
}

// GetTopGuessers retrieves top guessers who have guessed the most unique numbers
func (s *statsService) GetTopGuessers(ctx context.Context, req *types.GetTopGuessersRequest) (*types.GetTopGuessersResponse, error) {
	// Validate inputs
	if err := validator.New().Struct(req); err != nil {
		return nil, ierr.NewErrorWithContext(ctx, ierr.ErrCodeValidation, err).WithMessage("Make sure GameID & GuildID are set properly!")
	}

	redisStats := make([]*domain.UserGuessStats, 0)
	var game *domain.Game
	var err error

	if req.GameID != "" {
		// Game-specific stats from Redis - need to get game object first
		game, err = s.gameRepo.GetByID(ctx, req.GameID)
		if err != nil {
			return nil, ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, err).WithMessage("Failed to get game. Please try again later")
		}
		if game == nil {
			return nil, ierr.New(ierr.ErrCodeNotFound, "game not found")
		}

		redisStats, err = s.redisStatsRepo.GetGameTopGuessers(ctx, game, topLimit)
		if err != nil {
			return nil, ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, err).WithMessage("Failed to get game top numbers. Please try again later")
		}
	}

	mongoStats, err := s.gameStatsRepo.GetTopGuessers(ctx, req.GuildID, req.GameID, req.TimeRange, topLimit)
	if err != nil {
		return nil, ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, err).WithMessage("Failed to get top guessed numbers. Please try again later")
	}

	stats := make([]*domain.UserGuessStats, 0, len(redisStats)+len(mongoStats))
	stats = append(stats, redisStats...)
	stats = append(stats, mongoStats...)

	// deduplicate stats by userID
	deduplicatedStats := make([]*domain.UserGuessStats, 0, len(stats))
	for _, stat := range stats {
		if !slices.ContainsFunc(deduplicatedStats, func(s *domain.UserGuessStats) bool {
			return s.UserID == stat.UserID
		}) {
			deduplicatedStats = append(deduplicatedStats, stat)
		}
	}

	// sort stats by count
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].UniqueGuesses > stats[j].UniqueGuesses
	})
	if len(stats) > topLimit {
		stats = stats[:topLimit]
	}

	result := &types.GetTopGuessersResponse{
		Game:        game,
		TopGuessers: stats,
	}

	return result, nil
}

// GetTopWinners retrieves top winners (migrated from leaderboard)
// TODO: migrate to new stats system
func (s *statsService) GetTopWinners(ctx context.Context, req *types.GetTopWinnersRequest) (*types.GetTopWinnersResponse, error) {
	// Validate inputs
	if err := validator.New().Struct(req); err != nil {
		return nil, ierr.NewErrorWithContext(ctx, ierr.ErrCodeValidation, err).WithMessage("Make sure GuildID is set properly!")
	}
	if !req.ByGame && !req.ByPoints {
		return nil, ierr.New(ierr.ErrCodeValidation, "Either ByGame or ByPoints must be true")
	}

	guildData, err := s.guildDataRepo.Get(ctx, req.GuildID)
	if err != nil {
		return nil, ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, err).WithMessage("Failed to get top game winners. Please try again later")
	}

	if guildData == nil {
		return nil, ierr.New(ierr.ErrCodeNotFound, "No winners data found for this guild")
	}

	stats := make([]*domain.WinnerStats, 0, len(guildData.Users))
	for userID, user := range guildData.Users {
		stats = append(stats, &domain.WinnerStats{
			UserID: userID,
			Wins:   user.Wins,
			Points: user.Points,
		})
	}

	// sort stats by wins
	sort.Slice(stats, func(i, j int) bool {
		if req.ByGame {
			return stats[i].Wins > stats[j].Wins
		}
		return stats[i].Points > stats[j].Points
	})
	if len(stats) > topLimit {
		stats = stats[:topLimit]
	}

	result := &types.GetTopWinnersResponse{
		Winners: stats,
	}

	return result, nil
}
