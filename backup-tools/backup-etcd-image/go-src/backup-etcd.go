package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ncw/swift"
	"github.com/sapcc/containers/backup-tools/go-src/configuration"
	"github.com/sapcc/containers/backup-tools/go-src/prometheus"
	"github.com/sapcc/containers/backup-tools/go-src/swiftcli"
	"github.com/sapcc/containers/backup-tools/go-src/utils"
	"github.com/urfave/cli"
)

const (
	appName               = "ETCD Backup"
	layoutTimestamp       = "2006-01-02 15:04:05"
	layoutTimestampBackup = "2006-01-02_1504"
	tmpTimestampFile      = "/tmp/last_backup_timestamp"
	etcdDir               = "/var/lib/etcd2"
	tmpDir                = "/tmp"
	cmd                   = "/etcdctl"
)

var (
	backupDir      = utils.BackupPath
	etcdBackupDir  string
	etcdBackupDir2 string
	cmdArgsTemp    = []string{"backup"}
	cfg            *configuration.EnvironmentStruct
	t              []byte
	swiftCliConn   *swift.Connection
)

func init() {
	var err error
	var stats os.FileInfo

	log.SetOutput(os.Stderr)
	log.SetPrefix("[backup-etcd]")

	cfg = configuration.DefaultConfiguration

	if stats, err = os.Stat(tmpDir); err != nil {
		if os.IsNotExist(err) {
			// file does not exist
			os.MkdirAll(tmpDir, 0777)
		}
	}
	if err == nil && stats.IsDir() && stats.Mode() != 0777 {
		os.Chmod(tmpDir, 0777)
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

	swiftCliConn = swiftcli.SwiftConnection(
		cfg.OsAuthVersion,
		cfg.OsAuthURL,
		cfg.OsUsername,
		cfg.OsPassword,
		cfg.OsUserDomainName,
		cfg.OsProjectName,
		cfg.OsProjectDomainName,
		cfg.OsRegionName,
		cfg.ContainerPrefix)

	if _, err = os.Stat(tmpTimestampFile); os.IsNotExist(err) {
		swiftcli.SwiftDownloadFile(swiftCliConn, tmpTimestampFile, &backupDir)
	}

	files, _ := ioutil.ReadDir(etcdDir)
	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		if !strings.HasPrefix(f.Name(), "master") {
			continue
		}
		etcdBackupDir = strings.Join([]string{etcdDir, f.Name()}, string(os.PathSeparator))
		break
	}

	cmdArgsTemp = append(cmdArgsTemp, "--data-dir="+strconv.Quote(etcdBackupDir))
	etcdBackupDir2 = strings.Join([]string{backupDir, etcdBackupDir}, string(os.PathSeparator))
	os.MkdirAll(etcdBackupDir2, 0777)

	go func() {
		for {
			tsBackup := time.Now().UTC()
			cmdArgs := append(cmdArgsTemp, "--backup-dir="+strconv.Quote(strings.Join([]string{backupDir, etcdBackupDir, tsBackup.Format(layoutTimestampBackup)}, string(os.PathSeparator))))

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
				if err := ioutil.WriteFile(tmpTimestampFile, []byte(tsBackup.Format(layoutTimestamp)), 0777); err != nil {
					log.Println(err)
				} else {
					bp.SetSuccess(&timestamp)
				}
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
