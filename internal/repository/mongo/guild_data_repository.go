package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/diabolusgx/guess-the-number-go/internal/domain"
	ierr "github.com/diabolusgx/guess-the-number-go/pkg/errors"
	"github.com/diabolusgx/guess-the-number-go/pkg/logger"
	mongoPkg "github.com/diabolusgx/guess-the-number-go/pkg/mongo"
)

type GuildDataRepository struct {
	log        *logger.Logger
	collection *mongo.Collection
}

func NewGuildDataRepository(log *logger.Logger, mongoClient mongoPkg.BaseClient) *GuildDataRepository {
	return &GuildDataRepository{
		log:        log,
		collection: mongoClient.GetCollection(guildDataCollection),
	}
}

func (r *GuildDataRepository) Get(ctx context.Context, id string) (*domain.GuildData, error) {
	span := StartRepositorySpan(ctx, guildDataCollection, "get", map[string]any{
		"guildID": id,
	})
	defer FinishSpan(span)

	filter := bson.M{"id": id}
	var data domain.GuildData
	err := r.collection.FindOne(ctx, filter).Decode(&data)
	if err != nil {
		if err == ErrNoDocuments {
			ierr.NewErrorWithContext(ctx, ierr.ErrCodeNotFound, fmt.Errorf("guild data not found"))
			SetSpanError(span, err)
			return nil, nil
		}
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to get guild data: %w", err))
		SetSpanError(span, err)
		return nil, err
	}

	SetSpanSuccess(span)
	return &data, nil
}

func (r *GuildDataRepository) Replace(ctx context.Context, data *domain.GuildData) (*domain.GuildData, error) {
	span := StartRepositorySpan(ctx, guildDataCollection, "replace", map[string]any{
		"guildID": data.ID,
	})
	defer FinishSpan(span)

	filter := bson.M{"id": data.ID}
	var updatedData *domain.GuildData

	err := r.collection.FindOneAndReplace(ctx, filter, data).Decode(&updatedData)
	if err != nil {
		if err == ErrNoDocuments {
			ierr.NewErrorWithContext(ctx, ierr.ErrCodeNotFound, fmt.Errorf("guild data not found"))
			SetSpanError(span, err)
			return nil, nil
		}
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to replace guild data: %w", err))
		SetSpanError(span, err)
		return nil, err
	}

	SetSpanSuccess(span)
	return updatedData, nil
}

func (r *GuildDataRepository) UpdateGameAndUserStats(ctx context.Context, guildID, channelID, userID string, points, wins int64) error {
	span := StartRepositorySpan(ctx, guildDataCollection, "update_game_and_user_stats", map[string]any{
		"guildID":   guildID,
		"channelID": channelID,
		"userID":    userID,
		"points":    points,
		"wins":      wins,
	})
	defer FinishSpan(span)

	incr := bson.M{
		"totalGames": 1,
	}

	if userID != "" {
		incr[fmt.Sprintf("users.%s.points", userID)] = points
		incr[fmt.Sprintf("users.%s.wins", userID)] = wins
	}

	filter := bson.M{"id": guildID}
	update := bson.M{
		"$inc": incr,
		"$unset": bson.M{
			fmt.Sprintf("runningGames.%s", channelID): "",
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to update game and user stats: %w", err))
		SetSpanError(span, err)
		return err
	}

	SetSpanSuccess(span)
	return nil
}

// NOTE: merged UpdateGameStats & UpdateUserStats into UpdateGameAndUserStats however, we can keep them here for now
func (r *GuildDataRepository) UpdateGameStats(ctx context.Context, guildID, channelID string) error {
	span := StartRepositorySpan(ctx, guildDataCollection, "update_game_stats", map[string]any{
		"guildID":   guildID,
		"channelID": channelID,
	})
	defer FinishSpan(span)

	filter := bson.M{"id": guildID}
	update := bson.M{
		"$inc": bson.M{
			"totalGames": 1,
		},
		"$unset": bson.M{
			fmt.Sprintf("runningGames.%s", channelID): "",
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to update game stats: %w", err))
		SetSpanError(span, err)
		return err
	}

	SetSpanSuccess(span)
	return nil
}

// NOTE: merged UpdateGameStats & UpdateUserStats into UpdateGameAndUserStats however, we can keep them here for now
func (r *GuildDataRepository) UpdateUserStats(ctx context.Context, guildID, userID string, points, wins int64) error {
	span := StartRepositorySpan(ctx, guildDataCollection, "update_user_stats", map[string]any{
		"guildID": guildID,
		"userID":  userID,
		"points":  points,
		"wins":    wins,
	})
	defer FinishSpan(span)

	filter := bson.M{"id": guildID}
	update := bson.M{
		"$inc": bson.M{
			fmt.Sprintf("users.%s.points", userID): points,
			fmt.Sprintf("users.%s.wins", userID):   wins,
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to update user stats: %w", err))
		SetSpanError(span, err)
		return err
	}

	SetSpanSuccess(span)
	return nil
}

func (r *GuildDataRepository) Create(ctx context.Context, data *domain.GuildData) error {
	span := StartRepositorySpan(ctx, guildDataCollection, "create", map[string]any{
		"guildID": data.ID,
	})
	defer FinishSpan(span)

	_, err := r.collection.InsertOne(ctx, data)
	if err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to create guild data: %w", err))
		SetSpanError(span, err)
		return err
	}

	SetSpanSuccess(span)
	return nil
}

func (r *GuildDataRepository) Delete(ctx context.Context, id string) error {
	span := StartRepositorySpan(ctx, guildDataCollection, "delete", map[string]any{
		"guildID": id,
	})
	defer FinishSpan(span)

	filter := bson.M{"id": id}
	_, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to delete guild data: %w", err))
		SetSpanError(span, err)
		return err
	}

	SetSpanSuccess(span)
	return nil
}

func (r *GuildDataRepository) DeleteUser(ctx context.Context, guildID, userID string) error {
	span := StartRepositorySpan(ctx, guildDataCollection, "delete_user", map[string]any{
		"guildID": guildID,
		"userID":  userID,
	})
	defer FinishSpan(span)

	filter := bson.M{"id": guildID}
	update := bson.M{"$unset": bson.M{fmt.Sprintf("users.%s", userID): ""}}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to delete user: %w", err))
		SetSpanError(span, err)
		return err
	}

	SetSpanSuccess(span)
	return nil
}
