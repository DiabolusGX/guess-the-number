package service

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strconv"

	"github.com/diabolusgx/guess-the-number/internal/domain"
	"github.com/diabolusgx/guess-the-number/internal/lib"
	"github.com/diabolusgx/guess-the-number/internal/repository"
	"github.com/diabolusgx/guess-the-number/internal/types"
	ierr "github.com/diabolusgx/guess-the-number/pkg/errors"
	"github.com/diabolusgx/guess-the-number/pkg/logger"
	"github.com/diabolusgx/guess-the-number/pkg/metrics"
	"github.com/go-playground/validator/v10"
)

type GameService interface {
	CreateGame(ctx context.Context, req *types.CreateGameRequest) (*types.CreateGameResponse, error)
	HandleAttempt(ctx context.Context, req *types.HandleAttemptRequest) (*types.HandleAttemptResponse, error)
	FinishGame(ctx context.Context, req *types.FinishGameRequest) (*types.FinishGameResponse, error)
	GetHint(ctx context.Context, req *types.GetHintRequest) (*types.GetHintResponse, error)
	GetGameInfo(ctx context.Context, req *types.GetGameInfoRequest) (*types.GetGameInfoResponse, error)
	GetAllActiveGames(ctx context.Context) ([]*domain.Game, error)
}

type gameService struct {
	logger             *logger.Logger
	metrics            *metrics.Metrics
	gameRepo           domain.GameRepository
	guildDataRepo      domain.GuildDataRepository
	gameStatsRepo      domain.GameStatsRepository
	redisStatsRepo     domain.RedisStatsRepository
	transactionManager repository.TransactionManager
	runningGames       map[string]*domain.Game
}

func NewGameService(params ServiceParams) GameService {
	return &gameService{
		logger:             params.Logger,
		metrics:            params.Metrics,
		gameRepo:           params.GameRepo,
		guildDataRepo:      params.GuildDataRepo,
		gameStatsRepo:      params.GameStatsRepo,
		redisStatsRepo:     params.RedisStatsRepo,
		transactionManager: params.TransactionManager,

		runningGames: make(map[string]*domain.Game),
	}
}

func (s *gameService) CreateGame(ctx context.Context, req *types.CreateGameRequest) (*types.CreateGameResponse, error) {
	// Validate inputs
	if err := validator.New().Struct(req); err != nil {
		return nil, ierr.NewErrorWithContext(ctx, ierr.ErrCodeValidation, err).WithMessage("Make sure Min and Max values are set properly")
	}
	if req.LowerBound >= req.UpperBound {
		return nil, ierr.New(ierr.ErrCodeValidation, "Make sure Min is less than Max value")
	}

	// check if game is already running in this channel
	runningGame, err := s.getRunningInChannel(ctx, req.ChannelID)
	if err != nil {
		return nil, err
	}
	if runningGame != nil && !runningGame.Finished {
		return nil, ierr.New(ierr.ErrCodeAlreadyExists, "Game already running in this channel")
	}

	answer := rand.Int63n(req.UpperBound-req.LowerBound+1) + req.LowerBound
	points := int64(math.Ceil(float64(req.UpperBound-req.LowerBound) / 10))

	game := &domain.Game{
		ID:                lib.NewULID(lib.ULIDGameIDPrefix),
		GuildID:           req.GuildID,
		ChannelID:         req.ChannelID,
		CreatedBy:         req.CreatedBy,
		Answer:            answer,
		Points:            points,
		AutoReactionHints: req.AutoReactionHints,
	}

	err = s.gameRepo.Create(ctx, game)
	if err != nil {
		return nil, err
	}

	s.runningGames[req.ChannelID] = game

	s.metrics.GamesCreated.Inc()
	s.metrics.GamesRunning.Inc()

	return &types.CreateGameResponse{
		Game: game,
	}, nil
}

func (s *gameService) FinishGame(ctx context.Context, req *types.FinishGameRequest) (*types.FinishGameResponse, error) {
	if err := validator.New().Struct(req); err != nil {
		return nil, ierr.NewErrorWithContext(ctx, ierr.ErrCodeValidation, err).WithMessage("Make sure ChannelID and MessageID are set properly")
	}

	game, err := s.getRunningInChannel(ctx, req.ChannelID)
	if err != nil {
		return nil, err
	}
	if game == nil {
		return nil, ierr.New(ierr.ErrCodeNotFound, "There's no active game in given channel!")
	}

	ctx = context.WithValue(ctx, lib.CtxGameID, game.ID)

	_, txnErr := s.transactionManager.WithTransaction(ctx, func(ctx context.Context) (any, error) {
		if err := s.gameRepo.Finish(ctx, game.ID, req.MessageID, req.WonBy, req.Guesses); err != nil {
			return nil, err
		}

		err := s.guildDataRepo.UpdateGameAndUserStats(ctx, game.GuildID, game.ChannelID, req.WonBy, game.Points, 1)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})
	if txnErr != nil {
		return nil, txnErr
	}

	// TODO: Produce event for syncing game data to MongoDB

	delete(s.runningGames, req.ChannelID)
	s.metrics.GamesRunning.Dec()

	return &types.FinishGameResponse{
		Game: game,
	}, nil
}

func (s *gameService) HandleAttempt(ctx context.Context, req *types.HandleAttemptRequest) (*types.HandleAttemptResponse, error) {
	if err := validator.New().Struct(req); err != nil {
		return nil, ierr.NewErrorWithContext(ctx, ierr.ErrCodeValidation, err).WithMessage("invalid inputs")
	}

	game, err := s.getRunningInChannel(ctx, req.ChannelID)
	if err != nil {
		return nil, err
	}
	if game == nil {
		return nil, ierr.New(ierr.ErrCodeNotFound, "There's no active game in given channel!")
	}

	ctx = context.WithValue(ctx, lib.CtxGameID, game.ID)

	guesses, err := s.gameRepo.IncrementGuesses(ctx, game.ID, game.ChannelID)
	if err != nil {
		return nil, err
	}
	game.Guesses = guesses

	// Record guess asynchronously (no latency impact)
	go func() {
		ctx := lib.CopyContextKeys(ctx)

		// Calculate distance and create lightweight record
		distance := math.Abs(float64(req.Guess - game.Answer))

		if err := s.redisStatsRepo.RecordGuessAsync(ctx, game, req.UserID, req.Guess, int64(distance), req.Timestamp); err != nil {
			s.logger.FromContext(ctx).Error("failed to record guess in redis", "error", err, "gameID", game.ID, "userID", req.UserID, "guess", req.Guess)
		}
	}()

	s.logger.FromContext(ctx).Debugw("guesses incremented", "guesses", guesses)

	if req.Guess == game.Answer {
		game.WonBy = req.UserID
		game.WinMessageID = req.MessageID

		_, err = s.FinishGame(ctx, &types.FinishGameRequest{
			ChannelID: game.ChannelID,
			Guesses:   guesses,
			MessageID: req.MessageID,
			WonBy:     req.UserID,
		})
		if err != nil {
			return nil, err
		}

		return &types.HandleAttemptResponse{
			Correct: true,
			Game:    game,
		}, nil
	}

	return &types.HandleAttemptResponse{
		Correct: false,
		Game:    game,
	}, nil
}

func (s *gameService) GetHint(ctx context.Context, req *types.GetHintRequest) (*types.GetHintResponse, error) {
	if err := validator.New().Struct(req); err != nil {
		return nil, ierr.NewErrorWithContext(ctx, ierr.ErrCodeValidation, err).WithMessage("invalid inputs")
	}

	game, err := s.getRunningInChannel(ctx, req.ChannelID)
	if err != nil {
		return nil, err
	}
	if game == nil {
		return nil, ierr.New(ierr.ErrCodeNotFound, fmt.Sprintf("There is no running game in <#%s>", req.ChannelID))
	}

	ctx = context.WithValue(ctx, lib.CtxGameID, game.ID)

	var hint string

	switch req.HintType {
	case types.HintTypeNumber:
		if game.Answer < req.CompareTo {
			hint = "lower"
		} else {
			hint = "higher"
		}
	case types.HintTypeFirstDigit:
		ans := strconv.FormatInt(game.Answer, 10)
		hint = string(ans[0])
	case types.HintTypeLastDigit:
		ans := strconv.FormatInt(game.Answer, 10)
		hint = string(ans[len(ans)-1])
	default:
		return nil, ierr.New(ierr.ErrCodeValidation, "invalid hint type")
	}

	s.logger.FromContext(ctx).Debugw("hint generated", "hint", hint)

	return &types.GetHintResponse{
		GameID: game.ID,
		Hint:   hint,
	}, nil
}

func (s *gameService) GetGameInfo(ctx context.Context, req *types.GetGameInfoRequest) (*types.GetGameInfoResponse, error) {
	if req.GameID == "" && req.ChannelID == "" {
		return nil, ierr.New(ierr.ErrCodeValidation, "gameID or channelID is required")
	}

	if req.GameID != "" {
		game, err := s.gameRepo.GetByID(ctx, req.GameID)
		if err != nil {
			return nil, err
		}
		if game == nil {
			return nil, ierr.New(ierr.ErrCodeNotFound, "There's no active game in given channel!")
		}
		return &types.GetGameInfoResponse{
			Game: game,
		}, nil
	}

	game, err := s.getRunningInChannel(ctx, req.ChannelID)
	if err != nil {
		return nil, err
	}
	if game == nil {
		return nil, ierr.New(ierr.ErrCodeNotFound, "There's no active game in given channel!")
	}

	return &types.GetGameInfoResponse{
		Game: game,
	}, nil
}

func (s *gameService) GetAllActiveGames(ctx context.Context) ([]*domain.Game, error) {
	games := make([]*domain.Game, len(s.runningGames))
	i := 0
	for _, game := range s.runningGames {
		games[i] = game
		i++
	}
	return games, nil
}

func (s *gameService) getRunningInChannel(ctx context.Context, channelID string) (*domain.Game, error) {
	game, ok := s.runningGames[channelID]
	if ok && !game.Finished {
		return game, nil
	}

	dbGame, err := s.gameRepo.GetRunningInChannel(ctx, channelID)
	if err != nil {
		return nil, err
	}

	if dbGame == nil || dbGame.Finished {
		return nil, nil
	}

	return dbGame, nil
}
