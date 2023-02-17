package prometheus

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	backupLastSuccessGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "backup_last_success",
		Help: "Unix Timestamp of last successful backup run",
	})
	backupLastErrorGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "backup_last_error",
		Help: "Unix Timestamp of last failed backup run",
	})
	backupSuccessCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "backup_count_success",
		Help: "Counter for successful backup runs",
	})
	backupErrorCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "backup_count_error",
		Help: "Counter for failed backup runs",
	})
	backupLastStartGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "backup_last_start",
		Help: "Unix Timestamp of last backup start",
	})
	backupLastFinishGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "backup_last_finish",
		Help: "Unix Timestampf of last backup finish",
	})
)

func InitMetrics() {
	prometheus.MustRegister(backupErrorCounter)
	prometheus.MustRegister(backupSuccessCounter)
	prometheus.MustRegister(backupLastErrorGauge)
	prometheus.MustRegister(backupLastSuccessGauge)
	prometheus.MustRegister(backupLastStartGauge)
	prometheus.MustRegister(backupLastFinishGauge)

	backupLastErrorGauge.Set(0)
	backupLastSuccessGauge.Set(0)
	backupLastStartGauge.Set(0)
	backupLastFinishGauge.Set(0)
}

// Begin set the start time of the process
func Begin() {
	backupLastStartGauge.Set(float64(time.Now().Unix()))
}

// Finish set the finish time of the process
func Finish() {
	backupLastFinishGauge.Set(float64(time.Now().Unix()))
}

// SetError set the current time for lastError and incement the countError
func SetError() {
	backupLastErrorGauge.Set(float64(time.Now().Unix()))
	backupErrorCounter.Inc()
}

// SetSuccess set the current time for lastSuccess and incement the countSuccess
func SetSuccess(t time.Time) {
	backupLastSuccessGauge.Set(float64(t.Unix()))
	backupSuccessCounter.Inc()
}
