package service

import (
	"context"

	"github.com/diabolusgx/guess-the-number-go/internal/domain"
)

type GuildManagementService interface {
	GetGuildConfig(ctx context.Context, guildID string) (*domain.GuildConfig, error)
	ReplaceGuildConfig(ctx context.Context, cfg *domain.GuildConfig) (*domain.GuildConfig, error)
	UpdateUserStats(ctx context.Context, guildID, userID string, points, wins int) error
	CreateGuild(ctx context.Context, guildID string) error
	DeleteGuild(ctx context.Context, guildID string) error
	DeleteUser(ctx context.Context, guildID, userID string) error
	SetGuilds(count int)
}

type guildManagementService struct {
	ServiceParams
}

func NewGuildManagementService(params ServiceParams) GuildManagementService {
	return &guildManagementService{
		ServiceParams: params,
	}
}

func (s *guildManagementService) GetGuildConfig(ctx context.Context, guildID string) (*domain.GuildConfig, error) {
	return s.ServiceParams.GuildConfigRepo.Get(ctx, guildID)
}

func (s *guildManagementService) ReplaceGuildConfig(ctx context.Context, cfg *domain.GuildConfig) (*domain.GuildConfig, error) {
	return s.ServiceParams.GuildConfigRepo.Replace(ctx, cfg)
}

func (s *guildManagementService) UpdateUserStats(ctx context.Context, guildID, userID string, points, wins int) error {
	return s.ServiceParams.GuildDataRepo.UpdateUserStats(ctx, guildID, userID, points, wins)
}

func (s *guildManagementService) CreateGuild(ctx context.Context, guildID string) error {
	cfg := &domain.GuildConfig{
		ID:     guildID,
		Prefix: "gg",
		DM:     true,
	}
	if err := s.ServiceParams.GuildConfigRepo.Create(ctx, cfg); err != nil {
		return err
	}
	s.ServiceParams.Metrics.GuildsCreated.Inc()
	s.ServiceParams.Metrics.Guilds.Inc()

	data := &domain.GuildData{
		ID:           guildID,
		RunningGames: make(map[string]*domain.GameState),
		Users:        make(map[string]*domain.UserStats),
	}
	return s.ServiceParams.GuildDataRepo.Create(ctx, data)
}

func (s *guildManagementService) DeleteGuild(ctx context.Context, guildID string) error {
	if err := s.ServiceParams.GuildConfigRepo.Delete(ctx, guildID); err != nil {
		return err
	}
	s.ServiceParams.Metrics.Guilds.Dec()
	return s.ServiceParams.GuildDataRepo.Delete(ctx, guildID)
}

func (s *guildManagementService) DeleteUser(ctx context.Context, guildID, userID string) error {
	return s.ServiceParams.GuildDataRepo.DeleteUser(ctx, guildID, userID)
}

func (s *guildManagementService) SetGuilds(count int) {
	s.ServiceParams.Metrics.Guilds.Set(float64(count))
}
