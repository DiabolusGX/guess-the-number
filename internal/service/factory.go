package service

import (
	"github.com/diabolusgx/guess-the-number-go/internal/config"
	"github.com/diabolusgx/guess-the-number-go/internal/domain"
	"github.com/diabolusgx/guess-the-number-go/internal/repository"
	"github.com/diabolusgx/guess-the-number-go/pkg/logger"
	"github.com/diabolusgx/guess-the-number-go/pkg/metrics"
	"github.com/diabolusgx/guess-the-number-go/pkg/mongo"
	"github.com/diabolusgx/guess-the-number-go/pkg/sentry"
	"go.uber.org/fx"
)

type ServiceParams struct {
	fx.In

	Logger             *logger.Logger
	Config             *config.Configuration
	MongoClient        mongo.BaseClient
	SentryClient       *sentry.Client
	TransactionManager repository.TransactionManager

	// Repositories
	GameRepo        domain.GameRepository
	GuildConfigRepo domain.GuildConfigRepository
	GuildDataRepo   domain.GuildDataRepository
	Metrics         *metrics.Metrics
}
