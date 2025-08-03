package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	GamesCreated  prometheus.Counter
	GamesRunning  prometheus.Gauge
	Guesses       *prometheus.CounterVec
	Guilds        prometheus.Gauge
	GuildsCreated prometheus.Counter
	GuildsDeleted prometheus.Counter
	Errors        *prometheus.CounterVec
	Operations    *prometheus.CounterVec
}

const (
	MetricLabelValueTrue  = "true"
	MetricLabelValueFalse = "false"
	MetricLabelValueNA    = "N/A"
)

func NewMetrics() *Metrics {
	return &Metrics{
		GamesCreated: promauto.NewCounter(prometheus.CounterOpts{
			Name: "gtn_games_created_total",
			Help: "The total number of created games",
		}),
		GamesRunning: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "gtn_games_running",
			Help: "The number of running games",
		}),
		Guesses: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "gtn_guesses_total",
			Help: "The total number of guesses",
		}, []string{"correct", "error"}),
		Guilds: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "gtn_guilds",
			Help: "The number of guilds",
		}),
		GuildsCreated: promauto.NewCounter(prometheus.CounterOpts{
			Name: "gtn_guilds_created_total",
			Help: "The total number of created guilds",
		}),
		GuildsDeleted: promauto.NewCounter(prometheus.CounterOpts{
			Name: "gtn_guilds_deleted_total",
			Help: "The total number of deleted guilds",
		}),
		Errors: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "gtn_errors_total",
			Help: "The total number of errors",
		}, []string{"code"}),
		Operations: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "gtn_operations_total",
			Help: "The total number of operations (events and commands) processed",
		}, []string{"type", "name", "success"}),
	}
}
