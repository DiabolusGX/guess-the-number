package repository

import (
	"github.com/diabolusgx/guess-the-number-go/internal/config"
	"github.com/diabolusgx/guess-the-number-go/internal/domain"
	"github.com/diabolusgx/guess-the-number-go/internal/repository/mongo"
	"github.com/diabolusgx/guess-the-number-go/pkg/logger"
	mongoPkg "github.com/diabolusgx/guess-the-number-go/pkg/mongo"
	"go.uber.org/fx"
)

// RepositoryParams holds common dependencies for repositories
type RepositoryParams struct {
	fx.In

	Logger      *logger.Logger
	MongoClient mongoPkg.BaseClient
	Config      *config.Configuration
}

func NewGameRepository(p RepositoryParams) domain.GameRepository {
	return mongo.NewGameRepository(p.Logger, p.MongoClient)
}

func NewGuildConfigRepository(p RepositoryParams) domain.GuildConfigRepository {
	return mongo.NewGuildConfigRepository(p.Logger, p.MongoClient)
}

func NewGuildDataRepository(p RepositoryParams) domain.GuildDataRepository {
	return mongo.NewGuildDataRepository(p.Logger, p.MongoClient)
}
