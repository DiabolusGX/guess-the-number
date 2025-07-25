package service

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/diabolusgx/guess-the-number-go/internal/domain"
	"github.com/diabolusgx/guess-the-number-go/internal/lib"
	ierr "github.com/diabolusgx/guess-the-number-go/pkg/errors"
)

type GameService interface {
	CreateGame(ctx context.Context, guildID, channelID, createdBy string, lowerBound, upperBound int64) (*domain.Game, error)
	GetGameByChannelID(ctx context.Context, channelID string) (*domain.Game, error)
	IncrementGuesses(ctx context.Context, channelID string) (int64, error)
	FinishGame(ctx context.Context, game *domain.Game) error
	GetDailyLeaderboard(ctx context.Context, guildID string) ([]domain.Game, error)
	GetAllTimeLeaderboard(ctx context.Context, guildID string) ([]domain.Game, error)
	GetWeeklyLeaderboard(ctx context.Context, guildID string) ([]domain.Game, error)
	CheckAnswer(ctx context.Context, channelID, userID string, guess int64) (bool, *domain.Game, error)
}

type gameService struct {
	ServiceParams
}

func NewGameService(params ServiceParams) GameService {
	return &gameService{
		ServiceParams: params,
	}
}

func (s *gameService) CreateGame(ctx context.Context, guildID, channelID, createdBy string, lowerBound, upperBound int64) (*domain.Game, error) {
	// Validate inputs
	if guildID == "" || channelID == "" || createdBy == "" || lowerBound >= upperBound {
		return nil, ierr.NewErrorWithContext(ctx, ierr.ErrCodeValidation, errors.New("invalid inputs"))
	}

	answer := rand.Int63n(upperBound-lowerBound+1) + lowerBound
	points := (upperBound - lowerBound) / 10

	game := &domain.Game{
		ID:        lib.NewULID(lib.ULIDGameIDPrefix),
		GuildID:   guildID,
		ChannelID: channelID,
		CreatedBy: createdBy,
		Answer:    answer,
		Points:    points,
	}

	err := s.ServiceParams.GameRepo.Create(ctx, game)
	if err != nil {
		return nil, err
	}

	s.ServiceParams.Metrics.GamesCreated.Inc()
	s.ServiceParams.Metrics.GamesRunning.Inc()

	return game, nil
}

func (s *gameService) GetGameByChannelID(ctx context.Context, channelID string) (*domain.Game, error) {
	return s.ServiceParams.GameRepo.GetGameByChannelID(ctx, channelID)
}

func (s *gameService) DisableGame(ctx context.Context, gameID string, guesses int) error {
	return s.ServiceParams.GameRepo.DisableGame(ctx, gameID, guesses)
}

func (s *gameService) GetDailyLeaderboard(ctx context.Context, guildID string) ([]domain.Game, error) {
	return s.ServiceParams.GameRepo.GetDailyLeaderboard(ctx, guildID)
}

func (s *gameService) GetAllTimeLeaderboard(ctx context.Context, guildID string) ([]domain.Game, error) {
	return s.ServiceParams.GameRepo.GetAllTimeLeaderboard(ctx, guildID)
}

func (s *gameService) GetWeeklyLeaderboard(ctx context.Context, guildID string) ([]domain.Game, error) {
	now := time.Now()
	startOfWeek := now.AddDate(0, 0, -int(now.Weekday()))
	endOfWeek := startOfWeek.AddDate(0, 0, 7)
	return s.ServiceParams.GameRepo.GetGamesBetweenDates(ctx, guildID, startOfWeek, endOfWeek)
}

func (s *gameService) IncrementGuesses(ctx context.Context, channelID string) (int64, error) {
	return 0, nil
}

func (s *gameService) FinishGame(ctx context.Context, game *domain.Game) error {
	if err := s.ServiceParams.GameRepo.Finish(ctx, game.ID, game.WonBy, int(game.Points), int(game.Guesses)); err != nil {
		return err
	}
	s.ServiceParams.Metrics.GamesRunning.Dec()
	return nil
}

func (s *gameService) CheckAnswer(ctx context.Context, channelID, userID string, guess int64) (bool, *domain.Game, error) {
	game, err := s.ServiceParams.GameRepo.GetGameByChannelID(ctx, channelID)
	if err != nil {
		return false, nil, err
	}
	if game == nil {
		return false, nil, nil
	}

	if guess == game.Answer {
		game.WonBy = userID
		guesses, err := s.ServiceParams.GameRepo.IncrementGuesses(ctx, channelID)
		if err != nil {
			return false, nil, err
		}
		game.Guesses = guesses
		if err := s.FinishGame(ctx, game); err != nil {
			return false, nil, err
		}
		s.ServiceParams.Metrics.Guesses.WithLabelValues("true").Inc()
		return true, game, nil
	}

	if _, err := s.ServiceParams.GameRepo.IncrementGuesses(ctx, channelID); err != nil {
		return false, nil, err
	}
	s.ServiceParams.Metrics.Guesses.WithLabelValues("false").Inc()
	return false, nil, nil
}
