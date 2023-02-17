package prometheus

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Gauge struct for the prometheus
type Gauge struct {
	processBegin  prometheus.Gauge
	processFinish prometheus.Gauge
	lastSuccess   prometheus.Gauge
	lastError     prometheus.Gauge
	countSuccess  prometheus.Counter
	countError    prometheus.Counter
}

var (
	registry *prometheus.Registry
)

func init() {
	registry = prometheus.NewRegistry()
}

func (g *Gauge) initBackup() {
	g.lastSuccess = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "backup_last_success",
		Help: "Unix Timestamp of last successful backup run",
	})

	g.lastError = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "backup_last_error",
		Help: "Unix Timestamp of last failed backup run",
	})

	g.countSuccess = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "backup_count_success",
		Help: "Counter for successful backup runs",
	})

	g.countError = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "backup_count_error",
		Help: "Counter for failed backup runs",
	})

	g.processBegin = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "backup_last_start",
		Help: "Unix Timestamp of last backup start",
	})

	g.processFinish = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "backup_last_finish",
		Help: "Unix Timestampf of last backup finish",
	})

	g.registry()
}

func (g *Gauge) registry() {
	registry.MustRegister(g.countError)
	registry.MustRegister(g.countSuccess)
	registry.MustRegister(g.lastError)
	registry.MustRegister(g.lastSuccess)
	registry.MustRegister(g.processBegin)
	registry.MustRegister(g.processFinish)
}

func (g *Gauge) initLast() {
	g.lastError.Set(0)
	g.lastSuccess.Set(0)
}

// Beginn set the start time of the process
func (g *Gauge) Beginn() {
	g.processBegin.Set(float64(time.Now().Unix()))
}

// Finish set the finish time of the process
func (g *Gauge) Finish() {
	g.processFinish.Set(float64(time.Now().Unix()))
}

// SetError set the current time for lastError and incement the countError
func (g *Gauge) SetError() {
	g.lastError.Set(float64(time.Now().Unix()))
	g.countError.Inc()
}

// SetSuccess set the current time for lastSuccess and incement the countSuccess
func (g *Gauge) SetSuccess(ts *time.Time) {
	if ts != nil {
		g.lastSuccess.Set(float64(ts.Unix()))
	} else {
		g.lastSuccess.Set(float64(time.Now().Unix()))
	}
	g.countSuccess.Inc()
}

// ServerStart starts the prometheus metrics server
func (g *Gauge) ServerStart() {
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	port := os.Getenv("BACKUP_METRICS_PORT")
	if port == "" {
		log.Fatal(http.ListenAndServe(":9188", nil))
	} else {
		log.Fatal(http.ListenAndServe(":"+port, nil))
	}
}

// NewBackup create a Gauge object for Backup
func NewBackup() *Gauge {
	gauge := new(Gauge)
	gauge.initBackup()
	gauge.initLast()
	return gauge
}
