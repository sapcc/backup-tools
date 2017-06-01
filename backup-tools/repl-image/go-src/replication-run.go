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
	appName = "Database Backup Replication"
)

var (
	lastSuccess = prometheus.NewGauge(prometheus.GaugeOpts {
		Name: "last_success",
		Help: "Unix Timestamp of last successful database backup replication run",
	})

	lastError = prometheus.NewGauge(prometheus.GaugeOpts {
		Name: "last_error",
		Help: "Unix Timestamp of last failed database backup replication run",
	})

	registry = prometheus.NewRegistry()
)

func init() {
	registry.MustRegister(lastSuccess)
	registry.MustRegister(lastError)
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
	app.Usage = "Replicating Database Backups around the world"
	app.Action = runServer
	app.Run(os.Args)
}

func runServer(c *cli.Context) {
	lastSuccess.Set(0)
	lastError.Set(0)
	go func() {
		cmd := "/usr/local/sbin/backup-replication.sh"
		for {
			command := exec.Command(cmd)
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr
			if err := command.Run(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				lastError.Set(float64(time.Now().Unix()))
			} else {
				lastSuccess.Set(float64(time.Now().Unix()))
			}
			time.Sleep(14400 * time.Second)
		}
	}()

	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func versionString() string {
	return fmt.Sprintf("0.1.2")
}
