package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/diabolusgx/guess-the-number-go/pkg/logger"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/diabolusgx/guess-the-number-go/internal/domain"
	ierr "github.com/diabolusgx/guess-the-number-go/pkg/errors"
	pkgMongo "github.com/diabolusgx/guess-the-number-go/pkg/mongo"
)

type GameRepository struct {
	log        *logger.Logger
	collection *mongo.Collection
}

func NewGameRepository(log *logger.Logger, mongoClient pkgMongo.BaseClient) *GameRepository {
	return &GameRepository{
		log:        log,
		collection: mongoClient.GetCollection(gameCollection),
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

	r.log.Debugw("game created", "game", game)
	SetSpanSuccess(span)
	return nil
}

func (r *GameRepository) Finish(ctx context.Context, gameID, wonBy string, points, guesses int) error {
	span := StartRepositorySpan(ctx, gameCollection, "finish", map[string]any{
		"gameID":  gameID,
		"wonBy":   wonBy,
		"points":  points,
		"guesses": guesses,
	})
	defer FinishSpan(span)

	filter := bson.M{"id": gameID}
	update := bson.M{
		"$set": bson.M{
			"wonBy":      wonBy,
			"points":     points,
			"guesses":    guesses,
			"finished":   true,
			"finishedAt": time.Now(),
			"updatedAt":  time.Now(),
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to finish game: %w", err))
		SetSpanError(span, err)
		return err
	}

	r.log.Debugw("game finished", "gameID", gameID, "wonBy", wonBy, "points", points, "guesses", guesses)
	SetSpanSuccess(span)
	return nil
}

func (r *GameRepository) GetGameByChannelID(ctx context.Context, channelID string) (*domain.Game, error) {
	span := StartRepositorySpan(ctx, gameCollection, "get_game_by_channel_id", map[string]any{
		"channelID": channelID,
	})
	defer FinishSpan(span)

	filter := bson.M{"channelID": channelID}

	var game *domain.Game
	err := r.collection.FindOne(ctx, filter).Decode(&game)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			SetSpanSuccess(span)
			r.log.Debugw("game not found", "channelID", channelID)
			return nil, nil
		}
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to get game by channel ID: %w", err))
		SetSpanError(span, err)
		return nil, err
	}

	SetSpanSuccess(span)
	return game, nil
}

func (r *GameRepository) IncrementGuesses(ctx context.Context, channelID string) (int64, error) {
	span := StartRepositorySpan(ctx, gameCollection, "increment_guesses", map[string]any{
		"channelID": channelID,
	})
	defer FinishSpan(span)

	filter := bson.M{"channelID": channelID}
	update := bson.M{
		"$inc": bson.M{
			"guesses": 1,
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			SetSpanSuccess(span)
			return 0, ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("game not found: %w", err))
		}
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to increment guesses: %w", err))
		SetSpanError(span, err)
		return 0, err
	}

	SetSpanSuccess(span)
	return 0, nil
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

func (r *GameRepository) DisableGame(ctx context.Context, gameID string, guesses int) error {
	span := StartRepositorySpan(ctx, gameCollection, "disable_game", map[string]any{
		"gameID":  gameID,
		"guesses": guesses,
	})
	defer FinishSpan(span)

	filter := bson.M{"id": gameID}
	update := bson.M{
		"$set": bson.M{
			"guesses":  guesses,
			"disabled": true,
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to disable game: %w", err))
		SetSpanError(span, err)
		return err
	}

	SetSpanSuccess(span)
	return nil
}

func (r *GameRepository) GetBotLeaderboard(ctx context.Context, startDate, endDate time.Time) ([]domain.Game, error) {
	span := StartRepositorySpan(ctx, gameCollection, "get_bot_leaderboard", map[string]any{
		"startDate": startDate,
		"endDate":   endDate,
	})
	defer FinishSpan(span)

	filter := bson.M{
		"finished": true,
		"finishedAt": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}
	opts := options.Find().SetSort(bson.M{"points": -1})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to get bot leaderboard: %w", err))
		SetSpanError(span, err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var games []domain.Game
	err = cursor.All(ctx, &games)
	if err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to get bot leaderboard: %w", err))
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
