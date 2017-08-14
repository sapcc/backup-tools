package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/sapcc/containers/backup-tools/go-src/configuration"
	"github.com/sapcc/containers/backup-tools/go-src/prometheus"
	"github.com/urfave/cli"
)

const (
	appName               = "ETCD Backup"
	layoutTimestamp       = "2006-01-02 15:04:05"
	layoutTimestampBackup = "2006-01-02_1504"
)

var (
	cmd         = "/etcdctl"
	cmdArgsTemp = []string{"backup", "--data-dir=\"/var/lib/etcd2/master0\""}
	cfg         *configuration.EnvironmentStruct
)

func init() {
	cfg = new(configuration.EnvironmentStruct)
	os.MkdirAll("/tmp", 0755)
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
		{
			Name:  "Josef Fröhle",
			Email: "josef.froehle@sap.com",
		},
	}
	app.Usage = "Create ETCD Backups"
	app.Action = runServer
	app.Run(os.Args)
}

func runServer(c *cli.Context) {
	bp := prometheus.NewBackup()
	go func() {
		for {
			tsBackup := time.Now().UTC().Format(layoutTimestampBackup)
			cmdArgs := append(cmdArgsTemp, "--backup-dir=\"/backup/master0/"+tsBackup+"\"")
			command := exec.Command(cmd, cmdArgs...)
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr
			bp.Beginn()

			t, err := ioutil.ReadFile("/tmp/last_backup_timestamp")
			if err != nil {
				fmt.Print(err)
			}
			rx := regexp.MustCompile(`^([0-9]{4})([0-9]{2})([0-9]{2})([0-9]{2})([0-9]{2})$`)
			ts := rx.ReplaceAllString(strings.Trim(string(t), "\n"), "$1-$2-$3 $4:$5:00")

			timestamp, _ := time.Parse(layoutTimestamp, ts)

			if err := command.Run(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				bp.SetError()
			} else {
				bp.SetSuccess(&timestamp)
			}
			bp.Finish()
			time.Sleep(300 * time.Second)
		}
	}()

	bp.ServerStart()
}

func versionString() string {
	return fmt.Sprintf("0.0.1")
}
