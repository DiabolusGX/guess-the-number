package main

import (
	"runtime"

	"go.uber.org/fx"

	"github.com/diabolusgx/guess-the-number/internal/bot"
	"github.com/diabolusgx/guess-the-number/internal/config"
	"github.com/diabolusgx/guess-the-number/internal/repository"
	"github.com/diabolusgx/guess-the-number/internal/service"
	"github.com/diabolusgx/guess-the-number/internal/validator"
	"github.com/diabolusgx/guess-the-number/pkg/logger"
	"github.com/diabolusgx/guess-the-number/pkg/metrics"
	"github.com/diabolusgx/guess-the-number/pkg/mongo"
	"github.com/diabolusgx/guess-the-number/pkg/redis"
	"github.com/diabolusgx/guess-the-number/pkg/sentry"
)

func main() {
	// set GOMAXPROCS to equal to number of cores
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Initialize Fx application
	var opts []fx.Option

	// Core dependencies
	opts = append(
		opts,
		fx.Provide(
			// Validator
			validator.NewValidator,

			// Config
			config.NewConfig,

			// Sentry
			sentry.NewSentry,

			// Logger
			logger.NewLogger,

			// Metrics
			metrics.NewMetrics,

			// MongoDB
			mongo.NewClient,

			// Redis
			redis.NewClient,

			// Repositories
			repository.NewTransactionManager,
			repository.NewGameRepository,
			repository.NewGuildConfigRepository,
			repository.NewGuildDataRepository,
			repository.NewGameStatsRepository,
			repository.NewRedisStatsRepository,
		),
	)

	// Service layer
	opts = append(
		opts,
		fx.Provide(
			service.NewGameService,
			service.NewGuildManagementService,
			service.NewStatsService,
			service.NewSyncService,
		),
	)

	// Bot options
	opts = append(opts, bot.Module)

	// Lifecycle hooks
	opts = append(opts,
		fx.Invoke(
			bot.Start,
			metrics.NewMetricsServer,
		),
	)

	app := fx.New(opts...)
	app.Run()
}
