package main

import (
	"fmt"
	"os"
	"time"

	"github.com/sapcc/containers/backup-tools/go-src/prometheus"
	"github.com/urfave/cli"
)

const (
	appName = "Database Backup Replication"
)

var (
	// PromGauge is the prometheus pointer to use in the other files on same directory path
	PromGauge *prometheus.Gauge
)

func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Version = versionString()
	app.Authors = []cli.Author{
		{
			Name:  "Norbert Tretkowski",
			Email: "norbert.tretkowski@sap.com",
		},
		{
			Name:  "Josef Fröhle",
			Email: "josef.froehle@sap.com",
		},
	}
	app.Usage = "Replicating Database Backups around the world"
	app.Action = runServer
	app.Run(os.Args)
}

func runServer(c *cli.Context) {
	PromGauge = prometheus.NewReplication()
	go func() {
		for {
			LoadAndStartJobs()
			PromGauge.Finish()
			time.Sleep(14400 * time.Second)
		}
	}()

	PromGauge.ServerStart()
}

func versionString() string {
	return fmt.Sprintf("0.1.2")
}
