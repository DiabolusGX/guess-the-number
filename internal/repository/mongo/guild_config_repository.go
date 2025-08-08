package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/diabolusgx/guess-the-number/internal/domain"
	ierr "github.com/diabolusgx/guess-the-number/pkg/errors"
	"github.com/diabolusgx/guess-the-number/pkg/logger"
	mongoPkg "github.com/diabolusgx/guess-the-number/pkg/mongo"
)

type GuildConfigRepository struct {
	log        *logger.Logger
	collection *mongo.Collection
}

func NewGuildConfigRepository(log *logger.Logger, mongoClient mongoPkg.BaseClient) *GuildConfigRepository {
	return &GuildConfigRepository{
		log:        log,
		collection: mongoClient.GetCollection(guildConfigCollection),
	}
}

func (r *GuildConfigRepository) Get(ctx context.Context, id string) (*domain.GuildConfig, error) {
	span := StartRepositorySpan(ctx, guildConfigCollection, "get", map[string]any{
		"guildID": id,
	})
	defer FinishSpan(span)

	filter := bson.M{"id": id}
	var cfg domain.GuildConfig
	err := r.collection.FindOne(ctx, filter).Decode(&cfg)
	if err != nil {
		if err == ErrNoDocuments {
			ierr.NewErrorWithContext(ctx, ierr.ErrCodeNotFound, fmt.Errorf("guild config not found"))
			SetSpanError(span, err)
			return nil, nil
		}
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to get guild config: %w", err))
		SetSpanError(span, err)
		return nil, err
	}

	SetSpanSuccess(span)
	return &cfg, nil
}

func (r *GuildConfigRepository) Replace(ctx context.Context, cfg *domain.GuildConfig) error {
	span := StartRepositorySpan(ctx, guildConfigCollection, "replace", map[string]any{
		"guildID": cfg.ID,
	})
	defer FinishSpan(span)

	filter := bson.M{"id": cfg.ID}

	var updatedCfg domain.GuildConfig
	err := r.collection.FindOneAndReplace(ctx, filter, cfg).Decode(&updatedCfg)
	if err != nil {
		if err == ErrNoDocuments {
			ierr.NewErrorWithContext(ctx, ierr.ErrCodeNotFound, fmt.Errorf("guild config not found"))
			SetSpanError(span, err)
			return err
		}
		SetSpanError(span, err)
		return ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("update guild config failed, err: %w", err))
	}

	SetSpanSuccess(span)
	return nil
}

func (r *GuildConfigRepository) Create(ctx context.Context, cfg *domain.GuildConfig) error {
	span := StartRepositorySpan(ctx, guildConfigCollection, "create", map[string]any{
		"guildID": cfg.ID,
	})
	defer FinishSpan(span)

	_, err := r.collection.InsertOne(ctx, cfg)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			SetSpanSuccess(span)
			return nil
		}

		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to create guild config: %w", err))
		SetSpanError(span, err)
		return err
	}

	SetSpanSuccess(span)
	return nil
}

func (r *GuildConfigRepository) Delete(ctx context.Context, id string) error {
	span := StartRepositorySpan(ctx, guildConfigCollection, "delete", map[string]any{
		"guildID": id,
	})
	defer FinishSpan(span)

	filter := bson.M{"id": id}
	_, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to delete guild config: %w", err))
		SetSpanError(span, err)
		return err
	}

	SetSpanSuccess(span)
	return nil
}
