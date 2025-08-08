package repository

import (
	"github.com/diabolusgx/guess-the-number/internal/config"
	"github.com/diabolusgx/guess-the-number/internal/domain"
	"github.com/diabolusgx/guess-the-number/internal/repository/mongo"
	"github.com/diabolusgx/guess-the-number/pkg/logger"
	mongoPkg "github.com/diabolusgx/guess-the-number/pkg/mongo"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

// RepositoryParams holds common dependencies for repositories
type RepositoryParams struct {
	fx.In

	Logger      *logger.Logger
	MongoClient mongoPkg.BaseClient
	RedisClient *redis.Client
	Config      *config.Configuration
}

func NewGameRepository(p RepositoryParams) domain.GameRepository {
	return mongo.NewGameRepository(p.Logger, p.MongoClient, p.RedisClient)
}

func NewGuildConfigRepository(p RepositoryParams) domain.GuildConfigRepository {
	return mongo.NewGuildConfigRepository(p.Logger, p.MongoClient)
}

func NewGuildDataRepository(p RepositoryParams) domain.GuildDataRepository {
	return mongo.NewGuildDataRepository(p.Logger, p.MongoClient)
}

func NewGameStatsRepository(p RepositoryParams) domain.GameStatsRepository {
	return mongo.NewGameStatsRepository(p.Logger, p.MongoClient)
}

func NewRedisStatsRepository(p RepositoryParams) domain.RedisStatsRepository {
	return mongo.NewRedisStatsRepository(p.Logger, p.RedisClient)
}
