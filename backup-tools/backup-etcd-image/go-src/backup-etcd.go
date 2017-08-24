package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"path"
	"strconv"

	"github.com/mholt/archiver"
	"github.com/ncw/swift"
	"github.com/sapcc/containers/backup-tools/go-src/configuration"
	"github.com/sapcc/containers/backup-tools/go-src/prometheus"
	"github.com/sapcc/containers/backup-tools/go-src/swiftcli"
	"github.com/sapcc/containers/backup-tools/go-src/utils"
	"github.com/urfave/cli"
)

const (
	appName               = "ETCD Backup"
	layoutTimestamp       = "200601021504"
	layoutTimestamp2      = "2006-01-02 15:04:05"
	layoutTimestampBackup = "2006-01-02_1504"
	tmpTimestampFile      = "/tmp/last_backup_timestamp"
	etcdDir               = "/var/lib/etcd2"
	cmd                   = "/etcdctl"
	interval              = time.Duration(60) * time.Second
)

var (
	backupExpire      int64
	backupExpireTmp   int
	backupInterval    time.Duration
	backupIntervalTmp int
	tmpDir            = "/tmp"
	backupDir         = utils.BackupPath
	etcdBackupDir     string
	etcdBackupDir2    string
	cmdArgsTemp       = []string{"backup"}
	cfg               *configuration.EnvironmentStruct
	t                 []byte
	swiftCliConn      *swift.Connection
)

func init() {
	var err error
	var stats os.FileInfo

	exe3 := os.Getenv("BACKUP_EXPIRATION_AFTER")

	backupExpireTmp, err = strconv.Atoi(exe3)
	if err != nil {
		backupExpire = 864000
	} else {
		backupExpire = int64(backupExpireTmp)
	}

	exe3 = os.Getenv("BACKUP_INTERVAL")
	backupIntervalTmp, err = strconv.Atoi(exe3)
	if err != nil {
		backupIntervalTmp = 864000
	}
	backupInterval = time.Duration(backupIntervalTmp) * time.Second

	log.SetOutput(os.Stdout)
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
		_, err = swiftcli.SwiftDownloadFile(swiftCliConn, cfg.ContainerPrefix+tmpTimestampFile, &tmpDir)
		if err != nil {
			log.Println("Download TimeStampFile:", err)
		}
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

	cmdArgsTemp = append(cmdArgsTemp, "--data-dir="+etcdBackupDir)

	go func() {
		for {
			tsBackup := time.Now().UTC()
			backupDataDir := strings.Join([]string{backupDir, tsBackup.Format(layoutTimestampBackup)}, string(os.PathSeparator))
			cmdArgs := append(cmdArgsTemp, "--backup-dir="+backupDataDir)

			command := exec.Command(cmd, cmdArgs...)
			command.Stdout = os.Stdout
			command.Stderr = os.Stdout
			bp.Beginn()

			file, err := os.Open(tmpTimestampFile)
			if err != nil {
				log.Println("TimeStampFile Error:", err)
			}

			t, err := ioutil.ReadAll(file)
			file.Close()
			if err != nil {
				t = []byte("200001010101")
				log.Println("Read TimestampFile error:", err)
			}

			rx := regexp.MustCompile(`^([0-9]{4})([0-9]{2})([0-9]{2})([0-9]{2})([0-9]{2})$`)
			ts := rx.ReplaceAllString(strings.Trim(string(t), "\n"), "$1-$2-$3 $4:$5:00")

			timestamp, _ := time.Parse(layoutTimestamp2, ts)

			bp.SetSuccess(&timestamp)

			if timestamp.Add(backupInterval).After(time.Now()) {
				goto GOSLEEP
			}

			log.Println("backupDataDir:", backupDataDir)

			if err := command.Run(); err != nil {
				log.Println("Command error:", err)
				bp.SetError()
			} else {
				if err := ioutil.WriteFile(tmpTimestampFile, []byte(tsBackup.Format(layoutTimestamp)), 0777); err != nil {
					log.Println("TimestampFile error:", err)
				} else {
					fakeObjectName := path.Clean(strings.Join([]string{cfg.ContainerPrefix, tmpTimestampFile}, string(os.PathSeparator)))
					log.Println("fakeObjectName:", fakeObjectName)
					done1, err := swiftcli.SwiftUploadFile(swiftCliConn, tmpTimestampFile, nil, &fakeObjectName)
					if err != nil {
						log.Println("SwiftUploadFile tmp:", err)
					}
					if err := archiver.TarGz.Make(backupDataDir+".tgz", []string{backupDataDir}); err != nil {
						log.Println("archiver.TarGz:", err)
					}
					fakeObjectName = path.Clean(strings.Join([]string{cfg.ContainerPrefix, backupDataDir + ".tgz"}, string(os.PathSeparator)))
					log.Println("fakeObjectName:", fakeObjectName)
					done2, err := swiftcli.SwiftUploadFile(swiftCliConn, backupDataDir+".tgz", &backupExpire, &fakeObjectName)
					if err != nil {
						log.Println("SwiftUploadFile backup:", err)
					}
					if done1 && done2 {
						log.Println("Backup successful")
						bp.SetSuccess(nil)
					}
					os.Remove(backupDataDir + ".tgz")
					os.RemoveAll(backupDataDir)
				}
			}
		GOSLEEP:
			bp.Finish()
			time.Sleep(interval)
		}
	}()

	bp.ServerStart()
}

func versionString() string {
	return fmt.Sprintf("0.0.1")
}
