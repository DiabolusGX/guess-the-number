package service

import (
	"github.com/diabolusgx/guess-the-number/internal/config"
	"github.com/diabolusgx/guess-the-number/internal/domain"
	"github.com/diabolusgx/guess-the-number/internal/repository"
	"github.com/diabolusgx/guess-the-number/pkg/logger"
	"github.com/diabolusgx/guess-the-number/pkg/metrics"
	"github.com/diabolusgx/guess-the-number/pkg/mongo"
	"github.com/diabolusgx/guess-the-number/pkg/sentry"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

type ServiceParams struct {
	fx.In

	Logger             *logger.Logger
	Config             *config.Configuration
	MongoClient        mongo.BaseClient
	RedisClient        *redis.Client
	SentryClient       *sentry.Client
	TransactionManager repository.TransactionManager

	// Repositories
	GameRepo        domain.GameRepository
	GuildConfigRepo domain.GuildConfigRepository
	GuildDataRepo   domain.GuildDataRepository
	GameStatsRepo   domain.GameStatsRepository
	RedisStatsRepo  domain.RedisStatsRepository
	Metrics         *metrics.Metrics
}
