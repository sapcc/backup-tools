package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/sapcc/containers/backup-tools/go-src/prometheus"
	"github.com/urfave/cli"
)

const (
	appName = "Database Backup Replication"
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
	bp := prometheus.NewReplication()
	go func() {
		cmd := "/usr/local/sbin/backup-replication.sh"
		for {
			command := exec.Command(cmd)
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr
			bp.Beginn()
			if err := command.Run(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				bp.SetError()
			} else {
				bp.SetSuccess(nil)
			}
			bp.Finish()
			time.Sleep(14400 * time.Second)
		}
	}()

	bp.ServerStart()
}

func versionString() string {
	return fmt.Sprintf("0.1.2")
}
