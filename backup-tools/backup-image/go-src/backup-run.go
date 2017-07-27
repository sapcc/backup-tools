package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
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

			t, err := ioutil.ReadFile("/tmp/last_backup_timestamp")
			if err != nil {
			  fmt.Print(err)
			}
			rx := regexp.MustCompile(`^([0-9]{4})([0-9]{2})([0-9]{2})([0-9]{2})([0-9]{2})$`)
			ts := rx.ReplaceAllString(strings.Trim(string(t), "\n"), "$1-$2-$3 $4:$5:00 PM")

			layout := "2006-01-02 03:04:05 PM"
			timestamp, err := time.Parse(layout, ts)

			if err := command.Run(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				lastError.Set(float64(time.Now().Unix()))
				countError.Inc()
			} else {
				lastSuccess.Set(float64(timestamp.Unix()))
				countSuccess.Inc()
			}
			backupFinish.Set(float64(time.Now().Unix()))
			time.Sleep(600 * time.Second)
		}
	}()

	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	port := os.Getenv("BACKUP_METRICS_PORT")
	if port == "" {
	  log.Fatal(http.ListenAndServe(":9188", nil))
	} else {
	  log.Fatal(http.ListenAndServe(":" + port, nil))
	}
}

func versionString() string {
	return fmt.Sprintf("0.1.2")
}
