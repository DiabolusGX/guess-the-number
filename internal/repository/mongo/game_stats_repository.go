package mongo

import (
	"context"
	"fmt"

	"github.com/diabolusgx/guess-the-number/internal/domain"
	ierr "github.com/diabolusgx/guess-the-number/pkg/errors"
	"github.com/diabolusgx/guess-the-number/pkg/logger"
	mongoPkg "github.com/diabolusgx/guess-the-number/pkg/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type gameStatsRepository struct {
	collection *mongo.Collection
	logger     *logger.Logger
}

func NewGameStatsRepository(logger *logger.Logger, client mongoPkg.BaseClient) domain.GameStatsRepository {
	return &gameStatsRepository{
		collection: client.GetCollection(gameStatsCollection),
		logger:     logger,
	}
}

func (r *gameStatsRepository) BulkInsert(ctx context.Context, attempts []*domain.GuessAttempt) error {
	span := StartRepositorySpan(ctx, gameStatsCollection, "bulk_insert", map[string]any{})
	defer FinishSpan(span)

	if len(attempts) == 0 {
		SetSpanSuccess(span)
		return nil
	}

	documents := make([]any, len(attempts))
	for i, attempt := range attempts {
		documents[i] = attempt
	}

	opts := options.InsertMany().SetOrdered(false) // Continue on duplicate key errors
	_, err := r.collection.InsertMany(ctx, documents, opts)
	if err != nil {
		// Log but don't fail on duplicate key errors
		if mongo.IsDuplicateKeyError(err) {
			ierr.NewWarningWithContext(ctx, ierr.ErrCodeDatabase, fmt.Sprintf("some duplicate guess attempts ignored during bulk insert: %s", err.Error()))
			SetSpanSuccess(span)
			return nil
		}
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to bulk insert guess attempts: %w", err))
		SetSpanError(span, err)
		return fmt.Errorf("failed to bulk insert guess attempts: %w", err)
	}

	SetSpanSuccess(span)
	return nil
}

// Deprecated: do not use, not optimized
func (r *gameStatsRepository) GetByGame(ctx context.Context, gameID string) ([]*domain.GuessAttempt, error) {
	span := StartRepositorySpan(ctx, gameStatsCollection, "get_by_game", map[string]any{
		"gameID": gameID,
	})
	defer FinishSpan(span)

	filter := bson.M{"gameId": gameID}
	opts := options.Find().SetSort(bson.M{"timestamp": 1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to find guess attempts for game %s: %w", gameID, err))
		SetSpanError(span, err)
		return nil, fmt.Errorf("failed to find guess attempts for game %s: %w", gameID, err)
	}
	defer cursor.Close(ctx)

	var attempts []*domain.GuessAttempt
	if err := cursor.All(ctx, &attempts); err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to decode guess attempts: %w", err))
		SetSpanError(span, err)
		return nil, fmt.Errorf("failed to decode guess attempts: %w", err)
	}

	SetSpanSuccess(span)
	return attempts, nil
}

// Deprecated: do not use, not optimized
func (r *gameStatsRepository) GetByGuildAndTimeRange(ctx context.Context, guildID string, timeRange domain.TimeRange) ([]*domain.GuessAttempt, error) {
	span := StartRepositorySpan(ctx, gameStatsCollection, "get_by_guild_and_time_range", map[string]any{
		"guildID":   guildID,
		"timeRange": timeRange,
	})
	defer FinishSpan(span)

	start, end := timeRange.GetTimeRange()

	filter := bson.M{
		"guildId": guildID,
		"timestamp": bson.M{
			"$gte": start,
			"$lte": end,
		},
	}

	opts := options.Find().SetSort(bson.M{"timestamp": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to find guess attempts for guild %s: %w", guildID, err))
		SetSpanError(span, err)
		return nil, fmt.Errorf("failed to find guess attempts for guild %s: %w", guildID, err)
	}
	defer cursor.Close(ctx)

	var attempts []*domain.GuessAttempt
	if err := cursor.All(ctx, &attempts); err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to decode guess attempts: %w", err))
		SetSpanError(span, err)
		return nil, fmt.Errorf("failed to decode guess attempts: %w", err)
	}

	SetSpanSuccess(span)
	return attempts, nil
}

func (r *gameStatsRepository) GetTopGuessedNumbers(ctx context.Context, guildID, gameID string, timeRange domain.TimeRange, limit int) ([]*domain.NumberFrequencyStats, error) {
	span := StartRepositorySpan(ctx, gameStatsCollection, "get_top_guessed_numbers", map[string]any{
		"guildID":   guildID,
		"gameID":    gameID,
		"timeRange": timeRange,
		"limit":     limit,
	})
	defer FinishSpan(span)

	filter := getFilter(gameID, guildID, timeRange)

	pipeline := mongo.Pipeline{
		bson.D{{"$match", filter}},
		bson.D{{"$group", bson.M{
			"_id":   "$guess",
			"count": bson.M{"$sum": 1},
		}}},
		bson.D{{"$sort", bson.M{"count": -1}}},
		bson.D{{"$limit", limit}},
		bson.D{{"$project", bson.M{
			"number": "$_id",
			"count":  "$count",
			"_id":    0,
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to aggregate top guessed numbers: %w", err))
		SetSpanError(span, err)
		return nil, fmt.Errorf("failed to aggregate top guessed numbers: %w", err)
	}
	defer cursor.Close(ctx)

	var results []*domain.NumberFrequencyStats
	if err := cursor.All(ctx, &results); err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to decode number frequency stats: %w", err))
		SetSpanError(span, err)
		return nil, fmt.Errorf("failed to decode number frequency stats: %w", err)
	}

	SetSpanSuccess(span)
	return results, nil
}

func (r *gameStatsRepository) GetTopGuessers(ctx context.Context, guildID, gameID string, timeRange domain.TimeRange, limit int) ([]*domain.UserGuessStats, error) {
	span := StartRepositorySpan(ctx, gameStatsCollection, "get_top_guessers", map[string]any{
		"guildID":   guildID,
		"gameID":    gameID,
		"timeRange": timeRange,
		"limit":     limit,
	})
	defer FinishSpan(span)

	filter := getFilter(gameID, guildID, timeRange)

	pipeline := mongo.Pipeline{
		bson.D{{"$match", filter}},
		bson.D{{"$group", bson.M{
			"_id":           "$userId",
			"uniqueGuesses": bson.M{"$addToSet": "$guess"},
			"totalGuesses":  bson.M{"$sum": 1},
		}}},
		bson.D{{"$project", bson.M{
			"userId":        "$_id",
			"uniqueGuesses": bson.M{"$size": "$uniqueGuesses"},
			"totalGuesses":  "$totalGuesses",
			"_id":           0,
		}}},
		bson.D{{"$sort", bson.M{"totalGuesses": -1}}},
		bson.D{{"$limit", limit}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to aggregate top guessers: %w", err))
		SetSpanError(span, err)
		return nil, fmt.Errorf("failed to aggregate top guessers: %w", err)
	}
	defer cursor.Close(ctx)

	var results []*domain.UserGuessStats
	if err := cursor.All(ctx, &results); err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to decode user guess stats: %w", err))
		SetSpanError(span, err)
		return nil, fmt.Errorf("failed to decode user guess stats: %w", err)
	}

	SetSpanSuccess(span)
	return results, nil
}

func (r *gameStatsRepository) GetClosestGuesses(ctx context.Context, guildID, gameID string, limit int) ([]*domain.GuessAttempt, error) {
	span := StartRepositorySpan(ctx, gameStatsCollection, "get_closest_guesses", map[string]any{
		"guildID": guildID,
		"gameID":  gameID,
		"limit":   limit,
	})
	defer FinishSpan(span)

	filter := getFilter(gameID, guildID, domain.TimeRange{
		Type: "all-time",
	})

	pipeline := mongo.Pipeline{
		bson.D{{"$match", filter}},
		bson.D{{"$sort", bson.M{
			"distance":  1,
			"timestamp": 1, // Earlier guesses first for same distance
		}}},
		bson.D{{"$limit", limit}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to aggregate closest guesses: %w", err))
		SetSpanError(span, err)
		return nil, fmt.Errorf("failed to aggregate closest guesses: %w", err)
	}
	defer cursor.Close(ctx)

	var results []*domain.GuessAttempt
	if err := cursor.All(ctx, &results); err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to decode closest guesses: %w", err))
		SetSpanError(span, err)
		return nil, fmt.Errorf("failed to decode closest guesses: %w", err)
	}

	SetSpanSuccess(span)
	return results, nil
}

func getFilter(gameID, guildID string, timeRange domain.TimeRange) bson.M {
	filter := bson.M{}

	start, end := timeRange.GetTimeRange()
	if !start.IsZero() && !end.IsZero() {
		filter["timestamp"] = bson.M{
			"$gte": start,
			"$lte": end,
		}
	}

	if gameID != "" {
		filter["gameId"] = gameID
	}

	if guildID != "" {
		filter["guildId"] = guildID
	}

	return filter
}
