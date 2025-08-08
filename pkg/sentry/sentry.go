package sentry

import (
	"context"
	"fmt"
	"time"

	"github.com/diabolusgx/guess-the-number/internal/config"
	"github.com/diabolusgx/guess-the-number/pkg/logger"
	sentrygo "github.com/getsentry/sentry-go"
	"go.uber.org/fx"
)

// Client wraps the Sentry client
type Client struct {
	hub     *sentrygo.Hub
	enabled bool
}

func NewSentry(lc fx.Lifecycle, cfg *config.Configuration, logger *logger.Logger) (*Client, error) {
	client := &Client{
		enabled: cfg.Sentry.Enabled,
		hub:     sentrygo.CurrentHub(),
	}

	if !cfg.Sentry.Enabled {
		logger.Info("Sentry is disabled")
		return client, nil
	}

	if cfg.Sentry.DSN == "" {
		logger.Warn("Sentry is enabled but DSN is empty, disabling Sentry")
		client.enabled = false
		return client, nil
	}

	err := sentrygo.Init(sentrygo.ClientOptions{
		Dsn:              cfg.Sentry.DSN,
		Environment:      cfg.Sentry.Environment,
		TracesSampleRate: cfg.Sentry.SampleRate,
		BeforeSend: func(event *sentrygo.Event, hint *sentrygo.EventHint) *sentrygo.Event {
			// Add additional context or filtering here
			return event
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Sentry: %w", err)
	}

	client.hub = sentrygo.CurrentHub()

	logger.Infow("Sentry initialized",
		"environment", cfg.Sentry.Environment,
		"sample_rate", cfg.Sentry.SampleRate,
	)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			sentrygo.Flush(2 * time.Second)
			return nil
		},
	})

	return client, nil
}

// CaptureException captures an exception to Sentry if enabled
func (c *Client) CaptureException(err error) {
	if !c.enabled {
		return
	}
	c.hub.CaptureException(err)
}

// CaptureMessage captures a message to Sentry if enabled
func (c *Client) CaptureMessage(message string) {
	if !c.enabled {
		return
	}
	c.hub.CaptureMessage(message)
}

// AddBreadcrumb adds a breadcrumb to Sentry if enabled
func (c *Client) AddBreadcrumb(breadcrumb *sentrygo.Breadcrumb) {
	if !c.enabled {
		return
	}
	c.hub.AddBreadcrumb(breadcrumb, nil)
}

// WithScope executes a function with a new Sentry scope if enabled
func (c *Client) WithScope(f func(scope *sentrygo.Scope)) {
	if !c.enabled {
		return
	}
	c.hub.WithScope(f)
}

// GetHub returns the underlying Sentry hub
func (c *Client) GetHub() *sentrygo.Hub {
	return c.hub
}
