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
)

func InitMetrics() {
	prometheus.MustRegister(backupLastSuccessGauge)
	backupLastSuccessGauge.Set(0)
}

// SetSuccess set the current time for lastSuccess and incement the countSuccess
func SetSuccess(t time.Time) {
	backupLastSuccessGauge.Set(float64(t.Unix()))
}
