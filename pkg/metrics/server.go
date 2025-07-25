package metrics

import (
	"context"
	"log"
	"net/http"

	"github.com/diabolusgx/guess-the-number-go/internal/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/fx"
)

type MetricsServer struct {
	server *http.Server
}

func NewMetricsServer(lc fx.Lifecycle, cfg *config.Configuration) *MetricsServer {
	if !cfg.Metrics.Enabled {
		return nil
	}

	http.Handle("/metrics", promhttp.Handler())
	server := &http.Server{Addr: cfg.Metrics.ListenAddr}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Printf("metrics server error: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})

	return &MetricsServer{server: server}
}
