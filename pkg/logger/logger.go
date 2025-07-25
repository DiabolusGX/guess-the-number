package logger

import (
	"context"

	"github.com/diabolusgx/guess-the-number-go/internal/config"
	"github.com/diabolusgx/guess-the-number-go/internal/lib"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.SugaredLogger to provide logging functionality
type Logger struct {
	*zap.SugaredLogger
}

// Global logger for convenience
var L *Logger

// NewLogger creates and returns a new Logger instance
func NewLogger(cfg *config.Configuration) (*Logger, error) {
	config := zap.NewProductionConfig()

	if cfg.Logging.Level == lib.LogLevelDebug {
		config = zap.NewDevelopmentConfig()
	}

	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	zapLogger, err := config.Build()
	if err != nil {
		return nil, err
	}

	logger := &Logger{
		SugaredLogger: zapLogger.Sugar(),
	}

	// Set global logger
	L = logger

	return logger, nil
}

// Initialize default logger and set it as global while also using Dependency Injection
func init() {
	L, _ = NewLogger(config.GetDefaultConfig())
}

func GetLogger() *Logger {
	if L == nil {
		L, _ = NewLogger(config.GetDefaultConfig())
	}
	return L
}

func GetLoggerWithContext(ctx context.Context) *Logger {
	return GetLogger().WithContext(ctx)
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

func (l *Logger) WithContext(ctx context.Context) *Logger {
	return &Logger{
		SugaredLogger: l.SugaredLogger.With(
			"request_id", lib.GetRequestID(ctx),
			"user_id", lib.GetUserID(ctx),
			"guild_id", lib.GetGuildID(ctx),
		),
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
