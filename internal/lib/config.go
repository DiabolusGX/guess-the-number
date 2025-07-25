package lib

type RunMode string

const (
	// ModeLocal is the mode for running both the bot and other services locally
	ModeLocal RunMode = "local"
	// ModeProd is the mode for running in production
	ModeProd RunMode = "prod"
)

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)
