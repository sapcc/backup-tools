package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
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
	tmpTimestampFile      = "/tmp/last_backup_timestamp"
)

var (
	etcdBackupDir string
	cmd           = "/etcdctl"
	cmdArgsTemp   = []string{"backup"}
	cfg           *configuration.EnvironmentStruct
	t             []byte
)

func init() {
	var err error
	var stats os.FileInfo

	cfg = new(configuration.EnvironmentStruct)

	if stats, err = os.Stat("/tmp"); err != nil {
		if os.IsNotExist(err) {
			// file does not exist
			os.MkdirAll("/tmp", 0777)
		}
	}
	if err == nil && stats.IsDir() && stats.Mode() != 0777 {
		os.Chmod("/tmp", 0777)
	}
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
	var err error
	bp := prometheus.NewBackup()

	files, _ := ioutil.ReadDir("/var/lib/etcd2")
	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		if !strings.HasPrefix(f.Name(), "master") {
			continue
		}
		etcdBackupDir = strings.Join([]string{"/var/lib/etcd2", f.Name()}, string(os.PathSeparator))
	}

	cmdArgsTemp = append(cmdArgsTemp, "--data-dir="+strconv.Quote("/var/lib/etcd2/"+etcdBackupDir))

	os.MkdirAll("/backup/"+etcdBackupDir, 0777)

	go func() {
		for {
			tsBackup := time.Now().UTC()
			cmdArgs := append(cmdArgsTemp, "--backup-dir="+strconv.Quote(strings.Join([]string{"/backup", etcdBackupDir, tsBackup.Format(layoutTimestampBackup)}, string(os.PathSeparator))))

			command := exec.Command(cmd, cmdArgs...)
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr
			bp.Beginn()

			t, err = ioutil.ReadFile(tmpTimestampFile)
			if err != nil {
				fmt.Print(err)
			} else {
				t = []byte("200001010101")
			}
			rx := regexp.MustCompile(`^([0-9]{4})([0-9]{2})([0-9]{2})([0-9]{2})([0-9]{2})$`)
			ts := rx.ReplaceAllString(strings.Trim(string(t), "\n"), "$1-$2-$3 $4:$5:00")

			timestamp, _ := time.Parse(layoutTimestamp, ts)

			if err := command.Run(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				bp.SetError()
			} else {
				bp.SetSuccess(&timestamp)
				ioutil.WriteFile(tmpTimestampFile, []byte(tsBackup.Format(layoutTimestamp)), 0777)
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
