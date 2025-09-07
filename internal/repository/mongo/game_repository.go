package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/diabolusgx/guess-the-number/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/diabolusgx/guess-the-number/internal/domain"
	ierr "github.com/diabolusgx/guess-the-number/pkg/errors"
	pkgMongo "github.com/diabolusgx/guess-the-number/pkg/mongo"
)

type GameRepository struct {
	log         *logger.Logger
	collection  *mongo.Collection
	redisClient *redis.Client
}

func NewGameRepository(log *logger.Logger, mongoClient pkgMongo.BaseClient, redisClient *redis.Client) *GameRepository {
	return &GameRepository{
		log:         log,
		collection:  mongoClient.GetCollection(gameCollection),
		redisClient: redisClient,
	}
}

func (r *GameRepository) Create(ctx context.Context, game *domain.Game) error {
	span := StartRepositorySpan(ctx, gameCollection, "create", map[string]any{
		"guildID":   game.GuildID,
		"channelID": game.ChannelID,
		"createdBy": game.CreatedBy,
		"answer":    game.Answer,
		"points":    game.Points,
		"guesses":   game.Guesses,
		"finished":  game.Finished,
	})
	defer FinishSpan(span)

	game.CreatedAt = time.Now()
	game.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, game)
	if err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to create game: %w", err))
		SetSpanError(span, err)
		return err
	}

	r.log.FromContext(ctx).Debugw("game started")
	SetSpanSuccess(span)
	return nil
}

func (r *GameRepository) GetByID(ctx context.Context, gameID string) (*domain.Game, error) {
	span := StartRepositorySpan(ctx, gameCollection, "get_by_id", map[string]any{
		"gameID": gameID,
	})
	defer FinishSpan(span)

	filter := bson.M{"_id": gameID}
	var game domain.Game
	err := r.collection.FindOne(ctx, filter).Decode(&game)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			SetSpanSuccess(span)
			return nil, nil
		}
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to get game by ID: %w", err))
		SetSpanError(span, err)
		return nil, err
	}

	SetSpanSuccess(span)
	return &game, nil
}

func (r *GameRepository) Finish(ctx context.Context, gameID, channelID, messageID, wonBy string, guesses int64) error {
	span := StartRepositorySpan(ctx, gameCollection, "finish", map[string]any{
		"gameID":    gameID,
		"channelID": channelID,
		"wonBy":     wonBy,
		"guesses":   guesses,
	})
	defer FinishSpan(span)

	// TODO: use gameID instead of channelID in future, had to use channelID till legacy games are running
	// filter := bson.M{"_id": gameID}
	filter := bson.M{"channelID": channelID, "finished": bson.M{"$ne": true}}
	update := bson.M{
		"$set": bson.M{
			"wonBy":        wonBy,
			"winMessageID": messageID,
			"guesses":      guesses,
			"finished":     true,
			"finishedAt":   time.Now(),
			"updatedAt":    time.Now(),
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to finish game: %w", err))
		SetSpanError(span, err)
		return err
	}

	// clear guess count from redis
	out := r.redisClient.Unlink(ctx, getGameGuessesKey(gameID))
	if out.Err() != nil && out.Err() != redis.Nil {
		r.log.FromContext(ctx).Error("failed to clear guess count from redis", "error", out.Err().Error())
	}

	r.log.FromContext(ctx).Debugw("game ended", "gameID", gameID, "wonBy", wonBy, "guesses", guesses)
	SetSpanSuccess(span)
	return nil
}

func (r *GameRepository) GetRunningInChannel(ctx context.Context, channelID string) (*domain.Game, error) {
	span := StartRepositorySpan(ctx, gameCollection, "get_running_in_channel", map[string]any{
		"channelID": channelID,
	})
	defer FinishSpan(span)

	filter := bson.M{"channelID": channelID, "finished": bson.M{"$ne": true}}

	var game *domain.Game
	err := r.collection.FindOne(ctx, filter).Decode(&game)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			SetSpanSuccess(span)
			r.log.FromContext(ctx).Debugw("game not found", "channelID", channelID)
			return nil, nil
		}
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to get game by channel ID: %w", err))
		SetSpanError(span, err)
		return nil, err
	}

	if !game.Finished || game.Guesses == 0 {
		guesses, err := r.redisClient.Get(ctx, getGameGuessesKey(game.ID)).Int64()
		if err != nil && err != redis.Nil {
			r.log.FromContext(ctx).Error("failed to get guesses from redis", "error", err)
		} else if guesses != 0 {
			game.Guesses = guesses
		}
	}

	SetSpanSuccess(span)
	return game, nil
}

func (r *GameRepository) IncrementGuesses(ctx context.Context, gameID, channelID string) (int64, error) {
	span := StartRepositorySpan(ctx, "redis_game_guesses", "increment_guesses", map[string]any{
		"gameID":    gameID,
		"channelID": channelID,
	})
	defer FinishSpan(span)

	out := r.redisClient.Incr(ctx, getGameGuessesKey(gameID))
	if out.Err() != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to increment guesses: %w", out.Err()))
		SetSpanError(span, out.Err())
		return 0, out.Err()
	}

	SetSpanSuccess(span)
	return out.Val(), nil
}

func (r *GameRepository) GetDailyLeaderboard(ctx context.Context, guildID string) ([]domain.Game, error) {
	span := StartRepositorySpan(ctx, gameCollection, "get_daily_leaderboard", map[string]any{
		"guildID": guildID,
	})
	defer FinishSpan(span)

	filter := bson.M{
		"guildID":    guildID,
		"finished":   true,
		"finishedAt": bson.M{"$gte": time.Now().Add(-24 * time.Hour)},
	}
	opts := options.Find().SetSort(bson.M{"points": -1})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to get daily leaderboard: %w", err))
		SetSpanError(span, err)
		return nil, err
	}

	defer cursor.Close(ctx)
	var games []domain.Game
	err = cursor.All(ctx, &games)
	if err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to get daily leaderboard: %w", err))
		SetSpanError(span, err)
		return nil, err
	}

	SetSpanSuccess(span)
	return games, nil
}

func (r *GameRepository) GetGamesBetweenDates(ctx context.Context, guildID string, startDate, endDate time.Time) ([]domain.Game, error) {
	span := StartRepositorySpan(ctx, gameCollection, "get_games_between_dates", map[string]any{
		"guildID":   guildID,
		"startDate": startDate,
		"endDate":   endDate,
	})
	defer FinishSpan(span)

	filter := bson.M{
		"guildID":  guildID,
		"finished": true,
		"finishedAt": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}
	opts := options.Find().SetSort(bson.M{"points": -1})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to get games between dates: %w", err))
		SetSpanError(span, err)
		return nil, err
	}

	defer cursor.Close(ctx)
	var games []domain.Game
	err = cursor.All(ctx, &games)
	if err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to get games between dates: %w", err))
		SetSpanError(span, err)
		return nil, err
	}

	SetSpanSuccess(span)
	return games, nil
}

func (r *GameRepository) GetAllTimeLeaderboard(ctx context.Context, guildID string) ([]domain.Game, error) {
	span := StartRepositorySpan(ctx, gameCollection, "get_all_time_leaderboard", map[string]any{
		"guildID": guildID,
	})
	defer FinishSpan(span)

	filter := bson.M{"guildID": guildID, "finished": true}
	opts := options.Find().SetSort(bson.M{"points": -1})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to get all time leaderboard: %w", err))
		SetSpanError(span, err)
		return nil, err
	}

	defer cursor.Close(ctx)
	var games []domain.Game
	err = cursor.All(ctx, &games)
	if err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to get all time leaderboard: %w", err))
		SetSpanError(span, err)
		return nil, err
	}

	SetSpanSuccess(span)
	return games, nil
}

func getGameGuessesKey(gameID string) string {
	return fmt.Sprintf("guesses:%s", gameID)
}
