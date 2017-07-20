package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
	"net/http"
	"log"

	"github.com/urfave/cli"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	appName = "Database Backup"
)

var (
	lastSuccess = prometheus.NewGauge(prometheus.GaugeOpts {
		Name: "backup_last_success",
		Help: "Unix Timestamp of last successful backup run",
	})

	lastError = prometheus.NewGauge(prometheus.GaugeOpts {
		Name: "backup_last_error",
		Help: "Unix Timestamp of last failed backup run",
	})

	countSuccess = prometheus.NewCounter(prometheus.CounterOpts {
		Name: "backup_count_success",
		Help: "Counter for successful backup runs",
	})

	countError = prometheus.NewCounter(prometheus.CounterOpts {
		Name: "backup_count_error",
		Help: "Counter for failed backup runs",
	})

	backupBegin = prometheus.NewGauge(prometheus.GaugeOpts {
		Name: "backup_last_start",
		Help: "Unix Timestamp of last backup start",
	})

	backupFinish = prometheus.NewGauge(prometheus.GaugeOpts {
		Name: "backup_last_finish",
		Help: "Unix Timestampf of last backup finish",
	})

	registry = prometheus.NewRegistry()
)

func init() {
	registry.MustRegister(lastSuccess)
	registry.MustRegister(lastError)
	registry.MustRegister(countSuccess)
	registry.MustRegister(countError)
	registry.MustRegister(backupBegin)
	registry.MustRegister(backupFinish)
}

func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Version = versionString()
	app.Authors = []cli.Author{
		{
			Name:  "Norbert Tretkowski",
			Email: "norbert.tretkowski@sap.com",
		},
	}
	app.Usage = "Create Database Backups"
	app.Action = runServer
	app.Run(os.Args)
}

func runServer(c *cli.Context) {
	lastSuccess.Set(0)
	lastError.Set(0)
	go func() {
		cmd := "/usr/local/sbin/db-backup.sh"
		for {
			command := exec.Command(cmd)
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr
			backupBegin.Set(float64(time.Now().Unix()))
			if err := command.Run(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				lastError.Set(float64(time.Now().Unix()))
				countError.Inc()
			} else {
				lastSuccess.Set(float64(time.Now().Unix()))
				countSuccess.Inc()
			}
			backupFinish.Set(float64(time.Now().Unix()))
			time.Sleep(600 * time.Second)
		}
	}()

	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func versionString() string {
	return fmt.Sprintf("0.1.2")
}
