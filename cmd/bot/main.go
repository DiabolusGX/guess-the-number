package main

import (
	"go.uber.org/fx"

	"github.com/diabolusgx/guess-the-number-go/internal/bot"
	"github.com/diabolusgx/guess-the-number-go/internal/config"
	"github.com/diabolusgx/guess-the-number-go/internal/repository"
	"github.com/diabolusgx/guess-the-number-go/internal/service"
	"github.com/diabolusgx/guess-the-number-go/internal/validator"
	"github.com/diabolusgx/guess-the-number-go/pkg/logger"
	"github.com/diabolusgx/guess-the-number-go/pkg/metrics"
	"github.com/diabolusgx/guess-the-number-go/pkg/mongo"
	"github.com/diabolusgx/guess-the-number-go/pkg/redis"
	"github.com/diabolusgx/guess-the-number-go/pkg/sentry"
)

func main() {
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

			// Logger
			logger.NewLogger,

			// Sentry
			sentry.NewSentry,

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
		),
	)

	// Service layer
	opts = append(
		opts,
		fx.Provide(
			service.NewGameService,
			service.NewGuildManagementService,
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
