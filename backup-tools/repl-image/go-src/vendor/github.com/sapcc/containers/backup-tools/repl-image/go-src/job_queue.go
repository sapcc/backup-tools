package main

import (
	_ "expvar"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ncw/swift"
	"github.com/sapcc/containers/backup-tools/go-src/configuration"
	"github.com/sapcc/containers/backup-tools/go-src/swiftcli"
	"github.com/sapcc/containers/backup-tools/go-src/utils"
)

const maxWorkers = 5 // normal 5; debug 2

var (
	currentFilesDone = 0
	expiration       int64
	backupDir        = "/backup/tmp"
	EnvFrom          *Env
	EnvTo            = make([]*Env, 2)
)

type Env struct {
	Cfg      configuration.EnvironmentStruct
	Files    []string
	SwiftCli *swift.Connection
}

type Job struct {
	EnvFrom      *Env
	EnvTo        []*Env
	File         string
	FileAllCount int
	FileNumber   int
}

func doWork(id int, j Job) {
	var dlErr, upErr error
	var uploadDone bool
	var dlFilePath string
	lastDownload := false
	from := j.EnvFrom

	log.Printf("Worker%d File %d/%d: started process file %s\n", id, j.FileNumber, j.FileAllCount, j.File)

	for _, to := range j.EnvTo {
		uploadDone = false

		// skip if file already on the replication region present
		if stringInSlice(j.File, to.Files) {
			log.Printf("Worker%d File %d/%d: Skip %s to %s\n", id, j.FileNumber, j.FileAllCount, j.File, to.Cfg.OsRegionName)
			continue
		}

		// Download only file if not already done
		if !lastDownload {
			log.Printf("Worker%d File %d/%d: Download File: %s from %s\n", id, j.FileNumber, j.FileAllCount, j.File, from.Cfg.OsRegionName)
			dlFilePath, dlErr = swiftcli.SwiftDownloadFile(from.SwiftCli, j.File, &backupDir, true)
			if dlErr != nil {
				PromGauge.SetError()
				log.Printf("Worker%d File %d/%d: Download File Error: %s\n", id, j.FileNumber, j.FileAllCount, dlErr)
				// GoToEndWhileDownloadError used for exit the replication for error while downloading the file
				goto GoToEndWhileDownloadError
				break
			}
			lastDownload = true
		}

		// Upload file if no error before was triggered, and the dlFilePath is longer then 0
		if dlErr == nil && len(dlFilePath) > 0 {
			fakeName := strings.TrimPrefix(dlFilePath, backupDir)
			log.Printf("Worker%d File %d/%d: Upload File: %s to %s\n", id, j.FileNumber, j.FileAllCount, fakeName, to.Cfg.OsRegionName)
			uploadDone, upErr = swiftcli.SwiftUploadFile(to.SwiftCli, dlFilePath, &expiration, &fakeName)
			if upErr != nil || !uploadDone {
				PromGauge.SetError()
				log.Printf("Worker%d File %d/%d: Upload File Error: %s\n", id, j.FileNumber, j.FileAllCount, upErr)
			}
		}
	}

	// GoToEndWhileDownloadError used for exit the replication for error while downloading the file
GoToEndWhileDownloadError:

	if _, err := os.Stat(dlFilePath); !os.IsNotExist(err) {
		os.Remove(dlFilePath)
	}

	currentFilesDone += 1

	PromGauge.CurrentFile(currentFilesDone)

	log.Printf("Worker%d File %d/%d: completed %s\n", id, j.FileNumber, j.FileAllCount, j.File)
}

func StartJobWorkers() {

	var fileCounter = 0
	var countAllFiles = 0
	var startTime = time.Now()
	log.Println("Start replication task!")
	// channel for jobs
	jobs := make(chan Job)

	// start workers
	wg := &sync.WaitGroup{}
	wg.Add(maxWorkers)
	for i := 1; i <= maxWorkers; i++ {
		go func(i int) {
			defer wg.Done()

			for j := range jobs {
				doWork(i, j)
			}
		}(i)
	}

	for id, file := range EnvFrom.Files {

		if stringInAllSlice(file, EnvTo) {
			EnvFrom.Files[id] = ""
			//log.Printf("Skip File: %s its already on all replication regions\n", file)
		}

	}

	EnvFrom.Files = utils.DeleteEmpty(EnvFrom.Files)

	countAllFiles = len(EnvFrom.Files)
	PromGauge.AllFiles(countAllFiles)

	// add jobs
	for _, file := range EnvFrom.Files {

		fileCounter += 1

		jobs <- Job{
			EnvFrom:      EnvFrom,
			EnvTo:        EnvTo,
			File:         file,
			FileAllCount: countAllFiles,
			FileNumber:   fileCounter,
		}
	}
	close(jobs)

	// wait for workers to complete
	wg.Wait()

	os.RemoveAll(backupDir)

	var replicationDuration = time.Since(startTime)
	log.Printf("End replication task in %v\n", replicationDuration-(replicationDuration%time.Second))
}

func LoadAndStartJobs() {
	// cfg used for the parsed YAML Configuration
	cfg := configuration.YAMLReplication("/backup/env/config.yml")

	var err error
	var tmpExpireInt int
	tmpExpire := os.Getenv("BACKUP_EXPIRE_AFTER")
	if tmpExpire == "" {
		expiration = 864000
	} else {
		tmpExpireInt, err = strconv.Atoi(tmpExpire)
		if err == nil {
			expiration = int64(tmpExpireInt)
		} else {
			expiration = 864000
		}
	}

	os.MkdirAll(backupDir, 0777)

	EnvFrom = &Env{Cfg: cfg.From}
	EnvFrom.Cfg.ContainerPrefix = EnvFrom.Cfg.OsRegionName
	EnvFrom.SwiftCli = swiftcli.SwiftConnection(
		EnvFrom.Cfg.OsAuthVersion,
		EnvFrom.Cfg.OsAuthURL,
		EnvFrom.Cfg.OsUsername,
		EnvFrom.Cfg.OsPassword,
		EnvFrom.Cfg.OsUserDomainName,
		EnvFrom.Cfg.OsProjectName,
		EnvFrom.Cfg.OsProjectDomainName,
		EnvFrom.Cfg.OsRegionName,
		EnvFrom.Cfg.ContainerPrefix)
	EnvFrom.Files, _ = swiftcli.SwiftListPrefixFiles(EnvFrom.SwiftCli, EnvFrom.Cfg.ContainerPrefix)

	// Create for each replication region an own Env
	for id, toConfig := range cfg.To {
		EnvTo[id] = &Env{Cfg: toConfig}
		EnvTo[id].Cfg.ContainerPrefix = EnvFrom.Cfg.OsRegionName
		EnvTo[id].SwiftCli = swiftcli.SwiftConnection(
			EnvTo[id].Cfg.OsAuthVersion,
			EnvTo[id].Cfg.OsAuthURL,
			EnvTo[id].Cfg.OsUsername,
			EnvTo[id].Cfg.OsPassword,
			EnvTo[id].Cfg.OsUserDomainName,
			EnvTo[id].Cfg.OsProjectName,
			EnvTo[id].Cfg.OsProjectDomainName,
			EnvTo[id].Cfg.OsRegionName,
			EnvTo[id].Cfg.ContainerPrefix)
		EnvTo[id].Files, _ = swiftcli.SwiftListPrefixFiles(EnvTo[id].SwiftCli, EnvTo[id].Cfg.ContainerPrefix)
	}

	// Start Job Worker
	StartJobWorkers()
}

// helper function to look if path is already there
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func stringInAllSlice(a string, lists []*Env) bool {
	count := len(lists)
	found := 0
	for _, to := range lists {
		for _, b := range to.Files {
			if b == a {
				found += 1
				break
			}
		}
	}
	if count == found {
		return true
	}
	return false
}
