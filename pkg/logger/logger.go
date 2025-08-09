package logger

import (
	"context"

	"github.com/TheZeroSlave/zapsentry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/diabolusgx/guess-the-number/internal/config"
	"github.com/diabolusgx/guess-the-number/internal/lib"
	"github.com/diabolusgx/guess-the-number/pkg/sentry"
)

// Logger wraps zap.SugaredLogger to provide logging functionality
type Logger struct {
	*zap.SugaredLogger
}

// Global logger for convenience
var L *Logger

// NewLogger creates and returns a new Logger instance
func NewLogger(cfg *config.Configuration, sentry *sentry.Client) (*Logger, error) {
	config := zap.NewProductionConfig()

	if cfg.Deployment.Mode == lib.ModeLocal {
		config = zap.NewDevelopmentConfig()
	}

	config.OutputPaths = []string{"stdout", "/var/log/gtn-botinfo.log"}
	config.ErrorOutputPaths = []string{"stderr", "/var/log/gtn-boterror.log"}

	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.MessageKey = "message"

	zapLogger, err := config.Build()
	if err != nil {
		return nil, err
	}

	// Add Sentry core
	cfgSentry := zapsentry.Configuration{
		Hub:               sentry.GetHub(),
		Level:             zapcore.ErrorLevel, // Only send Error+ logs to Sentry
		EnableBreadcrumbs: true,
		BreadcrumbLevel:   zapcore.InfoLevel, // Info+ becomes breadcrumbs
	}

	sentryCore, err := zapsentry.NewCore(cfgSentry, zapsentry.NewSentryClientFromClient(sentry.GetHub().Client()))
	if err != nil {
		return nil, err
	}

	zapLogger = zapsentry.AttachCoreToLogger(sentryCore, zapLogger)

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
	loggerArgs := []any{
		"request_id", lib.GetRequestID(ctx),
		"user_id", lib.GetUserID(ctx),
		"guild_id", lib.GetGuildID(ctx),
		"channel_id", lib.GetChannelID(ctx),
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

	return &Logger{
		SugaredLogger: l.SugaredLogger.With(loggerArgs...),
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
