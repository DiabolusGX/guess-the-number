package service

import (
	"context"

	"github.com/diabolusgx/guess-the-number-go/internal/domain"
	"github.com/diabolusgx/guess-the-number-go/internal/repository"
	"github.com/diabolusgx/guess-the-number-go/pkg/metrics"
)

type GuildManagementService interface {
	GetGuildConfig(ctx context.Context, guildID string) (*domain.GuildConfig, error)
	GetGuildData(ctx context.Context, guildID string) (*domain.GuildData, error)
	ReplaceGuildConfig(ctx context.Context, cfg *domain.GuildConfig) error
	CreateGuild(ctx context.Context, guildID string) error
}

type guildManagementService struct {
	guildConfigRepo    domain.GuildConfigRepository
	guildDataRepo      domain.GuildDataRepository
	transactionManager repository.TransactionManager
	metrics            *metrics.Metrics
	guildConfigs       map[string]*domain.GuildConfig
}

func NewGuildManagementService(params ServiceParams) GuildManagementService {
	return &guildManagementService{
		guildConfigRepo:    params.GuildConfigRepo,
		guildDataRepo:      params.GuildDataRepo,
		transactionManager: params.TransactionManager,
		metrics:            params.Metrics,
		guildConfigs:       make(map[string]*domain.GuildConfig),
	}
}

func (s *guildManagementService) GetGuildConfig(ctx context.Context, guildID string) (*domain.GuildConfig, error) {
	// Check cache first
	if cfg, ok := s.guildConfigs[guildID]; ok {
		return cfg, nil
	}

	// Try to get from repository
	cfg, err := s.guildConfigRepo.Get(ctx, guildID)
	if err != nil {
		return nil, err
	}

	// If config doesn't exist (repository returns nil), create a new one with defaults
	if cfg == nil {
		cfg = &domain.GuildConfig{
			ID:     guildID,
			Prefix: "gg",
			DM:     true,
		}

		// Create the config in the repository
		if err := s.guildConfigRepo.Create(ctx, cfg); err != nil {
			return nil, err
		}
	}

	// Cache and return the config
	s.guildConfigs[guildID] = cfg
	return cfg, nil
}

func (s *guildManagementService) ReplaceGuildConfig(ctx context.Context, cfg *domain.GuildConfig) error {
	if err := s.guildConfigRepo.Replace(ctx, cfg); err != nil {
		return err
	}
	s.guildConfigs[cfg.ID] = cfg
	return nil
}

func (s *guildManagementService) CreateGuild(ctx context.Context, guildID string) error {
	_, txnErr := s.transactionManager.WithTransaction(ctx, func(ctx context.Context) (any, error) {
		cfg := &domain.GuildConfig{
			ID:     guildID,
			Prefix: "gg",
			DM:     true,
		}
		if err := s.guildConfigRepo.Create(ctx, cfg); err != nil {
			return nil, err
		}
		data := &domain.GuildData{
			ID:           guildID,
			RunningGames: make(map[string]*domain.GameState),
			Users:        make(map[string]*domain.UserStats),
		}

		err := s.guildDataRepo.Create(ctx, data)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})

	if txnErr != nil {
		return txnErr
	}

	s.metrics.GuildsCreated.Inc()
	s.metrics.Guilds.Inc()
	return nil
}

func (s *guildManagementService) GetGuildData(ctx context.Context, guildID string) (*domain.GuildData, error) {
	// Try to get from repository
	data, err := s.guildDataRepo.Get(ctx, guildID)
	if err != nil {
		return nil, err
	}

	// If data doesn't exist (repository returns nil), create a new one with defaults
	if data == nil {
		data = &domain.GuildData{
			ID:           guildID,
			RunningGames: make(map[string]*domain.GameState),
			Users:        make(map[string]*domain.UserStats),
		}

		// Create the data in the repository
		if err := s.guildDataRepo.Create(ctx, data); err != nil {
			return nil, err
		}
	}

	return data, nil
}
