package service

import (
	"context"
	"fmt"
	"time"

	"github.com/diabolusgx/guess-the-number/internal/domain"
	"github.com/diabolusgx/guess-the-number/internal/lib"
	"github.com/diabolusgx/guess-the-number/pkg/logger"
)

// SyncService handles background synchronization from Redis to MongoDB
type SyncService interface {
	StartPeriodicSync(ctx context.Context) error
	SyncGameGuesses(ctx context.Context, gameID string) error
	SyncAllActiveGames(ctx context.Context) error
	SyncGameOnFinish(ctx context.Context, gameID string) error
}

type syncService struct {
	logger *logger.Logger

	redisStatsRepo domain.RedisStatsRepository
	gameStatsRepo  domain.GameStatsRepository
	gameRepo       domain.GameRepository

	gameService GameService

	syncInterval     time.Duration
	enabled          bool
	onGameFinishSync bool
	batchSize        int
}

func NewSyncService(params ServiceParams, gameService GameService) SyncService {
	return &syncService{
		logger: params.Logger,

		redisStatsRepo: params.RedisStatsRepo,
		gameStatsRepo:  params.GameStatsRepo,
		gameRepo:       params.GameRepo,

		gameService: gameService,

		syncInterval:     params.Config.Sync.Interval,
		enabled:          params.Config.Sync.Enabled,
		onGameFinishSync: params.Config.Sync.OnGameFinish,
		batchSize:        params.Config.Sync.BatchSize,
	}
}

// StartPeriodicSync starts the background synchronization process
func (s *syncService) StartPeriodicSync(ctx context.Context) error {
	if !s.enabled {
		s.logger.FromContext(ctx).Info("sync service disabled, skipping periodic sync")
		return nil
	}

	s.logger.FromContext(ctx).Info("starting periodic sync service", "interval", s.syncInterval)

	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.FromContext(ctx).Info("sync service stopping due to context cancellation")
			return ctx.Err()
		case <-ticker.C:
			if err := s.SyncAllActiveGames(ctx); err != nil {
				s.logger.FromContext(ctx).Error("periodic sync failed", "error", err)
			}
		}
	}
}

// SyncAllActiveGames syncs all active games across all guilds
func (s *syncService) SyncAllActiveGames(ctx context.Context) error {
	ctx = context.WithValue(ctx, lib.CtxRequestID, lib.NewRequestID())

	s.logger.FromContext(ctx).Debug("starting sync for all active games")

	games, err := s.gameService.GetAllActiveGames(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all active games: %w", err)
	}

	for _, game := range games {
		if err := s.SyncGameGuesses(ctx, game.ID); err != nil {
			s.logger.FromContext(ctx).Error("failed to sync game guesses", "error", err)
		}
	}

	s.logger.FromContext(ctx).Debug("completed sync for all active games")
	return nil
}

// SyncGameGuesses syncs guesses for a specific game from Redis to MongoDB
func (s *syncService) SyncGameGuesses(ctx context.Context, gameID string) error {
	s.logger.FromContext(ctx).Debug("syncing game guesses")

	// Get game information from MongoDB
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game from database: %w", err)
	}
	if game == nil {
		s.logger.FromContext(ctx).Debug("game not found in database, skipping sync")
		return nil
	}

	// Get all guesses from Redis
	attempts, err := s.redisStatsRepo.GetGameGuesses(ctx, game)
	if err != nil {
		return fmt.Errorf("failed to get guesses from redis: %w", err)
	}

	if len(attempts) == 0 {
		s.logger.FromContext(ctx).Debug("no guesses to sync")
		return nil
	}

	// Generate IDs for the attempts if they don't have them
	for i := range attempts {
		if attempts[i].ID == "" {
			attempts[i].ID = fmt.Sprintf("%s_%d_%d", gameID, attempts[i].Timestamp.Unix(), i)
		}
	}

	// Bulk insert to MongoDB
	if err := s.gameStatsRepo.BulkInsert(ctx, attempts); err != nil {
		return fmt.Errorf("failed to bulk insert guesses: %w", err)
	}

	// TODO: fix this issue: redis stats repo is continously updated
	// so during the time of mongo insertion, the redis stats repo is also updated
	// cleanup will delete those new entries from redis

	// Clean up synced data from Redis
	if err := s.redisStatsRepo.CleanupGameData(ctx, game); err != nil {
		s.logger.FromContext(ctx).Error("failed to cleanup game data from redis after sync", "error", err)
		// Don't fail the sync for cleanup errors
	}

	s.logger.FromContext(ctx).Info("synced game guesses to mongodb", "count", len(attempts))
	return nil
}

// SyncGameOnFinish immediately syncs a game when it finishes
func (s *syncService) SyncGameOnFinish(ctx context.Context, gameID string) error {
	if !s.onGameFinishSync {
		return nil
	}

	s.logger.FromContext(ctx).Debug("syncing game on finish")
	return s.SyncGameGuesses(ctx, gameID)
}
