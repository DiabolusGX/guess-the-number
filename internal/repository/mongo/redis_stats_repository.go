package mongo

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/diabolusgx/guess-the-number/internal/domain"
	ierr "github.com/diabolusgx/guess-the-number/pkg/errors"
	"github.com/diabolusgx/guess-the-number/pkg/logger"
	"github.com/redis/go-redis/v9"
)

const (
	topLimit = 20
	ttl      = 24 * time.Hour

	prefixGameStats = "gs"

	// Redis key suffixes
	suffixGuesses = ":guesses"
	suffixNumbers = ":numbers"
	suffixUsers   = ":users"
	suffixClosest = ":closest"
)

// RedisStatsRepository handles Redis operations for game statistics
type RedisStatsRepository struct {
	redis  *redis.Client
	logger *logger.Logger
}

// LightweightGuess represents a compact guess record in Redis
type LightweightGuess struct {
	UserID    string `json:"u"`
	Guess     int64  `json:"g"`
	Distance  int64  `json:"d"`
	Timestamp int64  `json:"t"`
}

func NewRedisStatsRepository(logger *logger.Logger, redisClient *redis.Client) *RedisStatsRepository {
	return &RedisStatsRepository{
		redis:  redisClient,
		logger: logger,
	}
}

// RecordGuessAsync records a guess asynchronously with no latency impact
func (r *RedisStatsRepository) RecordGuessAsync(ctx context.Context, game *domain.Game, userID string, guess, distance, timestamp int64) error {
	span := StartRepositorySpan(ctx, "redis_stats", "record_guess_async", map[string]any{
		"gameID":    game.ID,
		"channelID": game.ChannelID,
		"userID":    userID,
		"guess":     guess,
	})
	defer FinishSpan(span)

	// Store individual guess (lightweight JSON)
	guessRecord := LightweightGuess{
		UserID:    userID,
		Guess:     guess,
		Timestamp: timestamp,
		Distance:  distance,
	}
	guessJSON, err := json.Marshal(guessRecord)
	if err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to marshal guess record: %w", err))
		SetSpanError(span, err)
		return err
	}

	// Build Redis keys with contextual information
	gamePrefix := getGamePrefix(game)

	// Redis pipeline for atomic operations
	pipe := r.redis.Pipeline()

	pipe.LPush(ctx, gamePrefix+suffixGuesses, string(guessJSON))

	// Update frequency counters
	pipe.HIncrBy(ctx, gamePrefix+suffixNumbers, fmt.Sprintf("%d", guess), 1)
	pipe.HIncrBy(ctx, gamePrefix+suffixUsers, userID, 1)

	// Update closest guesses (compact string format)
	closestMember := fmt.Sprintf("%s:%d:%d", userID, guess, timestamp)
	pipe.ZAdd(ctx, gamePrefix+suffixClosest, redis.Z{
		Score:  float64(distance),
		Member: closestMember,
	})
	pipe.ZRemRangeByRank(ctx, gamePrefix+suffixClosest, topLimit, -1) // Keep limited top guesses

	// Set TTL for cleanup (24 hours)
	keys := []string{
		gamePrefix + suffixGuesses,
		gamePrefix + suffixNumbers,
		gamePrefix + suffixUsers,
		gamePrefix + suffixClosest,
	}
	for _, key := range keys {
		pipe.Expire(ctx, key, ttl)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to record guess in redis: %w", err))
		SetSpanError(span, err)
		return err
	}

	SetSpanSuccess(span)
	return nil
}

// GetGameGuesses retrieves all guesses for a game from Redis
func (r *RedisStatsRepository) GetGameGuesses(ctx context.Context, game *domain.Game) ([]*domain.GuessAttempt, error) {
	span := StartRepositorySpan(ctx, "redis_stats", "get_game_guesses", map[string]any{
		"gameID": game.ID,
	})
	defer FinishSpan(span)

	gamePrefix := getGamePrefix(game)

	// Get all guesses from Redis
	guessesJSON, err := r.redis.LRange(ctx, gamePrefix+suffixGuesses, 0, -1).Result()
	if err != nil {
		if err == redis.Nil {
			SetSpanSuccess(span)
			return []*domain.GuessAttempt{}, nil
		}
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to get guesses from redis: %w", err))
		SetSpanError(span, err)
		return nil, fmt.Errorf("failed to get guesses from redis: %w", err)
	}

	// Parse lightweight JSON to full GuessAttempt structs
	var attempts []*domain.GuessAttempt
	for _, guessJSON := range guessesJSON {
		var lightGuess LightweightGuess
		if err := json.Unmarshal([]byte(guessJSON), &lightGuess); err == nil {
			attempt := &domain.GuessAttempt{
				GameID:    game.ID,
				GuildID:   game.GuildID,
				ChannelID: game.ChannelID,
				UserID:    lightGuess.UserID,
				Guess:     lightGuess.Guess,
				Timestamp: time.Unix(lightGuess.Timestamp, 0),
				Distance:  lightGuess.Distance,
				IsCorrect: lightGuess.Guess == game.Answer,
			}
			attempts = append(attempts, attempt)
		}
	}

	SetSpanSuccess(span)
	return attempts, nil
}

// GetGameTopNumbers retrieves top guessed numbers for a game from Redis
func (r *RedisStatsRepository) GetGameTopNumbers(ctx context.Context, game *domain.Game, limit int) ([]*domain.NumberFrequencyStats, error) {
	span := StartRepositorySpan(ctx, "redis_stats", "get_game_top_numbers", map[string]any{
		"gameID": game.ID,
		"limit":  limit,
	})
	defer FinishSpan(span)

	gamePrefix := getGamePrefix(game)

	// Get all number frequencies
	numberCounts, err := r.redis.HGetAll(ctx, gamePrefix+suffixNumbers).Result()
	if err != nil {
		if err == redis.Nil {
			// Key doesn't exist, return empty slice
			SetSpanSuccess(span)
			return []*domain.NumberFrequencyStats{}, nil
		}
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to get number frequencies from redis: %w", err))
		SetSpanError(span, err)
		return nil, err
	}

	// Convert to stats and sort by count
	stats := make([]*domain.NumberFrequencyStats, 0, len(numberCounts))
	for numberStr, countStr := range numberCounts {
		number, err := strconv.ParseInt(numberStr, 10, 64)
		if err != nil {
			r.logger.FromContext(ctx).Errorw("failed to parse number", "error", err)
			continue
		}
		count, err := strconv.ParseInt(countStr, 10, 64)
		if err != nil {
			r.logger.FromContext(ctx).Errorw("failed to parse count", "error", err)
			continue
		}
		stats = append(stats, &domain.NumberFrequencyStats{
			Number: number,
			Count:  count,
		})
	}

	// Sort by count descending (simple bubble sort for small datasets)
	for i := 0; i < len(stats)-1; i++ {
		for j := 0; j < len(stats)-1-i; j++ {
			if stats[j].Count < stats[j+1].Count {
				stats[j], stats[j+1] = stats[j+1], stats[j]
			}
		}
	}

	// Limit results
	if limit > 0 && limit < len(stats) {
		stats = stats[:limit]
	}

	SetSpanSuccess(span)
	return stats, nil
}

// GetGameTopGuessers retrieves top guessers for a game from Redis
func (r *RedisStatsRepository) GetGameTopGuessers(ctx context.Context, game *domain.Game, limit int) ([]*domain.UserGuessStats, error) {
	span := StartRepositorySpan(ctx, "redis_stats", "get_game_top_guessers", map[string]any{
		"gameID": game.ID,
		"limit":  limit,
	})
	defer FinishSpan(span)

	gamePrefix := getGamePrefix(game)

	// Get user guess counts
	userCounts, err := r.redis.HGetAll(ctx, gamePrefix+suffixUsers).Result()
	if err != nil {
		if err == redis.Nil {
			SetSpanSuccess(span)
			return []*domain.UserGuessStats{}, nil
		}
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to get user counts from redis: %w", err))
		SetSpanError(span, err)
		return nil, err
	}

	// Convert to stats and sort by count
	stats := make([]*domain.UserGuessStats, 0, len(userCounts))
	for userID, countStr := range userCounts {
		count, _ := strconv.ParseInt(countStr, 10, 64)
		stats = append(stats, &domain.UserGuessStats{
			UserID:        userID,
			TotalGuesses:  count,
			UniqueGuesses: count, // For Redis, we don't track unique separately yet
		})
	}

	// Sort by total guesses descending
	for i := 0; i < len(stats)-1; i++ {
		for j := 0; j < len(stats)-1-i; j++ {
			if stats[j].TotalGuesses < stats[j+1].TotalGuesses {
				stats[j], stats[j+1] = stats[j+1], stats[j]
			}
		}
	}

	// Limit results
	if limit > 0 && limit < len(stats) {
		stats = stats[:limit]
	}

	SetSpanSuccess(span)
	return stats, nil
}

// GetGameClosestGuesses retrieves closest guesses for a game from Redis
func (r *RedisStatsRepository) GetGameClosestGuesses(ctx context.Context, game *domain.Game, limit int) ([]*domain.GuessAttempt, error) {
	span := StartRepositorySpan(ctx, "redis_stats", "get_game_closest_guesses", map[string]any{
		"gameID": game.ID,
		"limit":  limit,
	})
	defer FinishSpan(span)

	gamePrefix := getGamePrefix(game)

	// Get closest guesses from sorted set
	closestMembers, err := r.redis.ZRange(ctx, gamePrefix+suffixClosest, 0, int64(limit-1)).Result()
	if err != nil {
		if err == redis.Nil {
			SetSpanSuccess(span)
			return []*domain.GuessAttempt{}, nil
		}
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to get closest guesses from redis: %w", err))
		SetSpanError(span, err)
		return nil, err
	}

	var attempts []*domain.GuessAttempt
	for _, member := range closestMembers {
		// Parse format: "userID:guess:timestamp"
		parts := parseClosestMember(member)
		if len(parts) == 3 {
			userID := parts[0]
			guess, _ := strconv.ParseInt(parts[1], 10, 64)
			timestamp, _ := strconv.ParseInt(parts[2], 10, 64)
			distance := abs(guess - game.Answer)

			attempt := &domain.GuessAttempt{
				GameID:    game.ID,
				GuildID:   game.GuildID,
				ChannelID: game.ChannelID,
				UserID:    userID,
				Guess:     guess,
				Timestamp: time.Unix(timestamp, 0),
				Distance:  distance,
				IsCorrect: guess == game.Answer,
			}
			attempts = append(attempts, attempt)
		}
	}

	SetSpanSuccess(span)
	return attempts, nil
}

func (r *RedisStatsRepository) CleanupGameData(ctx context.Context, game *domain.Game) error {
	span := StartRepositorySpan(ctx, "redis_stats", "cleanup_game_data", map[string]any{
		"gameID": game.ID,
	})
	defer FinishSpan(span)

	gamePrefix := getGamePrefix(game)

	// Delete all keys associated with the game
	keys := []string{
		gamePrefix + suffixGuesses,
		gamePrefix + suffixNumbers,
		gamePrefix + suffixUsers,
		gamePrefix + suffixClosest,
	}

	_, err := r.redis.Del(ctx, keys...).Result()
	if err != nil && err != redis.Nil {
		ierr.NewErrorWithContext(ctx, ierr.ErrCodeDatabase, fmt.Errorf("failed to cleanup game data from redis: %w", err))
		SetSpanError(span, err)
		return err
	}

	SetSpanSuccess(span)
	return nil
}

// Helper functions for Redis keys
func getGamePrefix(game *domain.Game) string {
	return fmt.Sprintf("%s:%s:%s", prefixGameStats, game.GuildID, game.ChannelID)
}

// Helper functions
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func parseClosestMember(member string) []string {
	result := make([]string, 0, 3)
	current := ""
	for _, char := range member {
		if char == ':' {
			result = append(result, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
