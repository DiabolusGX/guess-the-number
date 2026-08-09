package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/diabolusgx/guess-the-number/internal/lib"
	"github.com/diabolusgx/guess-the-number/internal/validator"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Configuration struct {
	Deployment DeploymentConfig `mapstructure:"deployment" validate:"required"`
	Bot        BotConfig        `mapstructure:"bot" validate:"required"`
	Mongo      MongoConfig      `mapstructure:"mongo" validate:"required"`
	Redis      RedisConfig      `mapstructure:"redis" validate:"required"`
	Logging    LoggingConfig    `mapstructure:"logging" validate:"required"`
	Sentry     SentryConfig     `mapstructure:"sentry" validate:"required"`
	Metrics    MetricsConfig    `mapstructure:"metrics" validate:"required"`
	Sync       SyncConfig       `mapstructure:"sync" validate:"required"`
}

type DeploymentConfig struct {
	Mode lib.RunMode `mapstructure:"mode" validate:"required"`
}

type BotConfig struct {
	Token        string         `mapstructure:"token" validate:"required"`
	DevGuilds    []string       `mapstructure:"dev_guilds"`
	SyncCommands bool           `mapstructure:"sync_commands" default:"true"`
	Sharding     ShardingConfig `mapstructure:"sharding"`
}

type ShardingConfig struct {
	Enabled     bool  `mapstructure:"enabled" default:"false"`
	ShardCount  int   `mapstructure:"shard_count" default:"0"` // 0 means auto-detect
	ShardIDs    []int `mapstructure:"shard_ids"`               // empty means all shards
	AutoScaling bool  `mapstructure:"auto_scaling" default:"true"`
}

type MongoConfig struct {
	Host        string `mapstructure:"host" validate:"required"`
	Username    string `mapstructure:"username" validate:"required"`
	Password    string `mapstructure:"password" validate:"required"`
	Database    string `mapstructure:"database" validate:"required"`
	Params      string `mapstructure:"params"`
	MaxPoolSize uint64 `mapstructure:"max_pool_size" validate:"required" default:"10"`
	MinPoolSize uint64 `mapstructure:"min_pool_size" validate:"required" default:"1"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host" validate:"required"`
	Port     string `mapstructure:"port" validate:"required"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db" default:"0"`
	PoolSize int    `mapstructure:"pool_size" default:"10"`
}

type LoggingConfig struct {
	Level lib.LogLevel `mapstructure:"level" validate:"required"`
}

type SentryConfig struct {
	Enabled     bool    `mapstructure:"enabled"`
	DSN         string  `mapstructure:"dsn"`
	Environment string  `mapstructure:"environment"`
	EnabledLogs bool    `mapstructure:"enabled_logs"`
	SampleRate  float64 `mapstructure:"sample_rate" default:"1.0"`
}

type MetricsConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	ListenAddr string `mapstructure:"listen_addr" default:":8080"`
}

type SyncConfig struct {
	Enabled      bool          `mapstructure:"enabled"`
	Interval     time.Duration `mapstructure:"interval"`
	OnGameFinish bool          `mapstructure:"on_game_finish"`
	BatchSize    int           `mapstructure:"batch_size"`
}

// Legacy type alias for backward compatibility
type Config = Configuration

func NewConfig() (*Configuration, error) {
	v := viper.New()

	// Step 1: Load `.env` if it exists
	_ = godotenv.Load()

	// Step 2: Initialize Viper
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./internal/config")
	v.AddConfigPath("../../internal/config")
	v.AddConfigPath("./config")
	v.AddConfigPath("/app/internal/config") // Docker container path

	// Step 3: Set up environment variables support
	v.SetEnvPrefix("GTN")
	v.AutomaticEnv()

	// Step 4: Environment variable key mapping
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Step 5: Read the YAML file
	if err := v.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return nil, err
		}
	} else {
		fmt.Printf("Using config file: %s\n", v.ConfigFileUsed())
	}

	var cfg Configuration
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into config struct: %v", err)
	}

	return &cfg, nil
}

func (c Configuration) Validate() error {
	return validator.ValidateRequest(c)
}

// GetDefaultConfig returns a default configuration for local development
func GetDefaultConfig() *Configuration {
	return &Configuration{
		Deployment: DeploymentConfig{Mode: lib.ModeLocal},
		Logging:    LoggingConfig{Level: lib.LogLevelDebug},
	}
}

func (m MongoConfig) GetConnectionURI() string {
	if m.Host == "" {
		return ""
	}

	if m.Username != "" && m.Password != "" {
		if m.Params != "" {
			return fmt.Sprintf("mongodb://%s:%s@%s/%s?%s", m.Username, m.Password, m.Host, m.Database, m.Params)
		}
		return fmt.Sprintf("mongodb+srv://%s:%s@%s/%s", m.Username, m.Password, m.Host, m.Database)
	}

	return fmt.Sprintf("mongodb://%s/%s", m.Host, m.Database)
}
