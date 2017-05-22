package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/urfave/cli"
	//"github.com/prometheus/client_golang/prometheus"
	//"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	appName = "Database Backup Replication"
)

var (
	landingPage = []byte(fmt.Sprintf(`<html>
<head><title>Database Backup Replication Exporter</title></head>
<body>
<h1>Database Backup Replication Exporter</h1>
<p>%s</p>
<p><a href='/metrics'>Metrics</a></p>
</body>
</html>
`, versionString()))

	Version   = "develop"
	GITCOMMIT = "HEAD"
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
	}
	app.Usage = "Database Backup Replication"
	app.Action = runServer
	app.Run(os.Args)
}

func runServer( c *cli.Context) {
	go func() {
		cmd := "/usr/local/sbin/backup-replication.sh"
		for {
			command := exec.Command(cmd)
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr
			if err := command.Run(); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			time.Sleep(14400 * time.Second)
		}
	}()
}

func versionString() string {
	return fmt.Sprintf("0.1.2")
}
