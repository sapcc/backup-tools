package main

import (
	_ "expvar"
	"log"
	"math"
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
	initialized      bool
	alreadyPrinted   int
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

	if DebugOutput {
		log.Printf("Worker%d File %d/%d: started process file %s\n", id, j.FileNumber, j.FileAllCount, j.File)
	}
	for _, to := range j.EnvTo {

		// skip if file already on the replication region present
		if stringInSlice(j.File, to.Files) {
			if DebugOutput {
				log.Printf("Worker%d File %d/%d: Skip %s to %s\n", id, j.FileNumber, j.FileAllCount, j.File, to.Cfg.OsRegionName)
			}
			continue
		}

		// Download only file if not already done
		if !lastDownload {
			if DebugOutput {
				log.Printf("Worker%d File %d/%d: Download File: %s from %s\n", id, j.FileNumber, j.FileAllCount, j.File, j.EnvFrom.Cfg.OsRegionName)
			}
			dlFilePath, dlErr = swiftcli.SwiftDownloadFile(from.SwiftCli, j.File, &backupDir, true)
			if dlErr != nil {
				PromGauge.SetError()
				log.Printf("Worker%d File %d/%d (%s): Download File Error: %s\n", id, j.FileNumber, j.FileAllCount, j.File, dlErr)
				// GoToEndWhileDownloadError used for exit the replication for error while downloading the file
				goto GoToEndWhileDownloadError
				break
			}
			lastDownload = true
		}

		// Upload file if no error before was triggered, and the dlFilePath is longer then 0
		if dlErr == nil && len(dlFilePath) > 0 {
			fakeName := strings.TrimPrefix(dlFilePath, backupDir)
			if DebugOutput {
				log.Printf("Worker%d File %d/%d: Upload File: %s to %s\n", id, j.FileNumber, j.FileAllCount, fakeName, to.Cfg.OsRegionName)
			}
			uploadDone, upErr = swiftcli.SwiftUploadFile(to.SwiftCli, dlFilePath, &expiration, &fakeName)
			if upErr != nil || !uploadDone {
				PromGauge.SetError()
				log.Printf("Worker%d File %d/%d (%s): Upload File Error: %s\n", id, j.FileNumber, j.FileAllCount, j.File, upErr)
			}
		}
	}

	// GoToEndWhileDownloadError used for exit the replication for error while downloading the file
GoToEndWhileDownloadError:

	os.Remove(dlFilePath)

	currentFilesDone += 1

	PromGauge.CurrentFile(currentFilesDone)

	if DebugOutput {
		log.Printf("Worker%d File %d/%d: completed %s\n", id, j.FileNumber, j.FileAllCount, j.File)
	}

	num := int(Round((float64(currentFilesDone) / float64(j.FileAllCount)) * 100.0))

	if num > alreadyPrinted || num == alreadyPrinted {
		alreadyPrinted += 5
		log.Printf("%d%% of replication done\n", num)
	}

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
		if strings.HasSuffix(file, "base") {
			EnvFrom.Files[id] = ""
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

	log.Printf("End replication task in %g seconds\n", time.Since(startTime).Seconds())
}

func LoadAndStartJobs() {
	var err error

	if !initialized {
		// cfg used for the parsed YAML Configuration
		cfg := configuration.YAMLReplication("/backup/env/config.yml")

		// Set all to false for a new loop as default
		alreadyPrinted = 0

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
		EnvFrom.SwiftCli, err = swiftcli.SwiftConnection(
			EnvFrom.Cfg.OsAuthVersion,
			EnvFrom.Cfg.OsAuthURL,
			EnvFrom.Cfg.OsUsername,
			EnvFrom.Cfg.OsPassword,
			EnvFrom.Cfg.OsUserDomainName,
			EnvFrom.Cfg.OsProjectName,
			EnvFrom.Cfg.OsProjectDomainName,
			EnvFrom.Cfg.OsRegionName,
			EnvFrom.Cfg.ContainerPrefix)
		if err != nil {
			log.Println("Error can't connect swift for", EnvFrom.Cfg.OsRegionName, err)
			return
		}

		// Create for each replication region an own Env
		for id, toConfig := range cfg.To {
			EnvTo[id] = &Env{Cfg: toConfig}
			EnvTo[id].Cfg.ContainerPrefix = EnvFrom.Cfg.OsRegionName
			EnvTo[id].SwiftCli, err = swiftcli.SwiftConnection(
				EnvTo[id].Cfg.OsAuthVersion,
				EnvTo[id].Cfg.OsAuthURL,
				EnvTo[id].Cfg.OsUsername,
				EnvTo[id].Cfg.OsPassword,
				EnvTo[id].Cfg.OsUserDomainName,
				EnvTo[id].Cfg.OsProjectName,
				EnvTo[id].Cfg.OsProjectDomainName,
				EnvTo[id].Cfg.OsRegionName,
				EnvTo[id].Cfg.ContainerPrefix)
			if err != nil {
				log.Println("Error can't connect swift for", EnvTo[id].Cfg.OsRegionName, err)
				return
			}
		}

		initialized = true
	}

	EnvFrom.Files, err = swiftcli.SwiftListPrefixFiles(EnvFrom.SwiftCli, EnvFrom.Cfg.ContainerPrefix)

	for id := range EnvTo {
		EnvTo[id].Files, err = swiftcli.SwiftListPrefixFiles(EnvTo[id].SwiftCli, EnvTo[id].Cfg.ContainerPrefix)
		if err != nil {
			log.Println("Error get files for", EnvTo[id].Cfg.OsRegionName, err)
			return
		}
	}

	if err != nil {
		log.Println("Error fet files for", EnvFrom.Cfg.OsRegionName, err)
		return
	}

	// Start Job Worker
	StartJobWorkers()
	return
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

func Round(f float64) float64 {
	return math.Floor(f + .5)
}
