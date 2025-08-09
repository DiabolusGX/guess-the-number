package metrics

import (
	"context"
	"log"
	"net/http"

	"github.com/diabolusgx/guess-the-number/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/fx"
)

type MetricsServer struct {
	server *http.Server
}

func NewMetricsServer(lc fx.Lifecycle, cfg *config.Configuration, metrics *Metrics) *MetricsServer {
	if !cfg.Metrics.Enabled {
		return nil
	}

	reg := prometheus.NewRegistry()

	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	reg.MustRegister(metrics.GamesCreated)
	reg.MustRegister(metrics.GamesRunning)
	reg.MustRegister(metrics.Guesses)
	reg.MustRegister(metrics.Guilds)
	reg.MustRegister(metrics.GuildsCreated)
	reg.MustRegister(metrics.GuildsDeleted)
	reg.MustRegister(metrics.Errors)
	reg.MustRegister(metrics.Operations)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	server := &http.Server{
		Addr:    cfg.Metrics.ListenAddr,
		Handler: mux,
	}

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
