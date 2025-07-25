package repository

import (
	"context"

	"github.com/diabolusgx/guess-the-number-go/pkg/logger"
	"github.com/diabolusgx/guess-the-number-go/pkg/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
)

type TransactionManager interface {
	WithTransaction(ctx context.Context, fn TransactionFunc) (any, error)
}

type TransactionFunc func(ctx context.Context) (any, error)

type mongoTransactionManager struct {
	Logger      *logger.Logger
	MongoClient mongo.BaseClient
}

func NewTransactionManager(p RepositoryParams) TransactionManager {
	// Since we're only using MongoDB right now, we can directly return the mongoTransactionManager.
	// In the future, if we need to support other databases, we can create a factory method
	return &mongoTransactionManager{
		Logger:      p.Logger,
		MongoClient: p.MongoClient,
	}
}

func (m *mongoTransactionManager) WithTransaction(ctx context.Context, fn TransactionFunc) (any, error) {
	session, err := m.MongoClient.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	val, err := session.WithTransaction(ctx, func(sessCtx context.Context) (any, error) {
		return fn(sessCtx)
	}, options.Transaction().SetWriteConcern(writeconcern.Majority()))

	return val, err
}
