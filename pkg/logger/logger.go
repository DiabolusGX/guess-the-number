package logger

import (
	"context"

	"github.com/TheZeroSlave/zapsentry"
	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/diabolusgx/guess-the-number/internal/config"
	"github.com/diabolusgx/guess-the-number/internal/lib"
	sentryPkg "github.com/diabolusgx/guess-the-number/pkg/sentry"
)

// Logger wraps zap.SugaredLogger to provide logging functionality
type Logger struct {
	*zap.SugaredLogger
}

// Global logger for convenience
var L *Logger

// NewLogger creates and returns a new Logger instance
func NewLogger(cfg *config.Configuration, sentryClient *sentryPkg.Client) (*Logger, error) {
	config := zap.NewProductionConfig()

	if cfg.Deployment.Mode == lib.ModeLocal {
		config = zap.NewDevelopmentConfig()
	}

	// config.OutputPaths = []string{"stdout", "~/logs/gtn-bot/info.log"}
	// config.ErrorOutputPaths = []string{"stderr", "~/logs/gtn-bot/error.log"}

	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.MessageKey = "message"

	zapLogger, err := config.Build()
	if err != nil {
		return nil, err
	}

	// Add Sentry core only if Sentry is enabled and properly initialized
	if sentryClient != nil && sentryClient.GetHub() != nil && sentryClient.GetHub().Client() != nil {
		cfgSentry := zapsentry.Configuration{
			Level:             zapcore.ErrorLevel, // Only send Error+ logs to Sentry
			EnableBreadcrumbs: false,              // Disable automatic breadcrumbs, we'll handle them manually
			BreadcrumbLevel:   zapcore.InfoLevel,  // Info+ becomes breadcrumbs
		}

		sentryCore, err := zapsentry.NewCore(cfgSentry, zapsentry.NewSentryClientFromClient(sentryClient.GetHub().Client()))
		if err != nil {
			return nil, err
		}

		zapLogger = zapsentry.AttachCoreToLogger(sentryCore, zapLogger)
	}

	logger := &Logger{
		SugaredLogger: zapLogger.Sugar(),
	}

	// Set global logger
	L = logger

	return logger, nil
}

func GetLogger() *Logger {
	if L == nil {
		L, _ = NewLogger(config.GetDefaultConfig(), nil)
	}
	return L
}

func GetLoggerFromContext(ctx context.Context) *Logger {
	return GetLogger().FromContext(ctx)
}

// Helper methods to make logging more convenient
func (l *Logger) Debugf(template string, args ...any) {
	l.SugaredLogger.Debugf(template, args...)
}

func (l *Logger) Infof(template string, args ...any) {
	l.SugaredLogger.Infof(template, args...)
}

func (l *Logger) Warnf(template string, args ...any) {
	l.SugaredLogger.Warnf(template, args...)
}

func (l *Logger) Errorf(template string, args ...any) {
	l.SugaredLogger.Errorf(template, args...)
}

func (l *Logger) Fatalf(template string, args ...any) {
	l.SugaredLogger.Fatalf(template, args...)
}

func (l *Logger) FromContext(ctx context.Context) *Logger {
	loggerArgs := []any{}

	if requestID := lib.GetRequestID(ctx); requestID != "" {
		loggerArgs = append(loggerArgs, "request_id", requestID)
	}

	if userID := lib.GetUserID(ctx); userID != "" {
		loggerArgs = append(loggerArgs, "user_id", userID)
	}

	if guildID := lib.GetGuildID(ctx); guildID != "" {
		loggerArgs = append(loggerArgs, "guild_id", guildID)
	}

	if channelID := lib.GetChannelID(ctx); channelID != "" {
		loggerArgs = append(loggerArgs, "channel_id", channelID)
	}

	if shardID := lib.GetShardID(ctx); shardID != 0 {
		loggerArgs = append(loggerArgs, "shard_id", shardID)
	}

	if priority := lib.GetPriority(ctx); priority != lib.PriorityDefault {
		loggerArgs = append(loggerArgs, "priority", priority.String())
	}

	if gameID := lib.GetGameID(ctx); gameID != "" {
		loggerArgs = append(loggerArgs, "game_id", gameID)
	}

	if flowID := lib.GetFlowID(ctx); flowID != "" {
		loggerArgs = append(loggerArgs, "flow_id", flowID)
	}

	// Add breadcrumbs to Sentry using context-specific hub
	l.addSentryBreadcrumb(ctx, loggerArgs)

	return &Logger{
		SugaredLogger: l.SugaredLogger.With(loggerArgs...),
	}
}

// addSentryBreadcrumb adds breadcrumbs to the correct Sentry hub from context
func (l *Logger) addSentryBreadcrumb(ctx context.Context, loggerArgs []any) {
	// Get hub from context, fallback to current hub if not available
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	// Only add breadcrumbs if we have a valid hub
	if hub != nil {
		// Convert logger args to breadcrumb data
		data := make(map[string]interface{})
		for i := 0; i < len(loggerArgs)-1; i += 2 {
			if key, ok := loggerArgs[i].(string); ok {
				data[key] = loggerArgs[i+1]
			}
		}

		hub.AddBreadcrumb(&sentry.Breadcrumb{
			Message: "Logger context created",
			Level:   sentry.LevelInfo,
			Data:    data,
		}, nil)
	}
}

// Convenience methods for structured logging
func (l *Logger) Debugw(msg string, keysAndValues ...any) {
	l.SugaredLogger.Debugw(msg, keysAndValues...)
}

func (l *Logger) Infow(msg string, keysAndValues ...any) {
	l.SugaredLogger.Infow(msg, keysAndValues...)
}

func (l *Logger) Warnw(msg string, keysAndValues ...any) {
	l.SugaredLogger.Warnw(msg, keysAndValues...)
}

func (l *Logger) Errorw(msg string, keysAndValues ...any) {
	l.SugaredLogger.Errorw(msg, keysAndValues...)
}
