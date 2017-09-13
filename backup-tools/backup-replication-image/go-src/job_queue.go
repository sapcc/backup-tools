package main

import (
	_ "expvar"
	"fmt"
	"io"
	"log"
	"math"
	"os"
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
	backupContainer    = "db_backup"
	initialized        bool
	alreadyPrinted     int
	currentFilesDone   = 0
	expiration         string
	EnvFrom            *Env
	EnvTo              = make([]*Env, 2)
	lockAlreadyPrinted = sync.RWMutex{}
)

//FileState is used by GetFile() to describe the state of a file.
type FileState struct {
	Etag         string
	LastModified string
	//the following fields are only used in `sourceState`, not `targetState`
	SkipTransfer bool
	ContentType  string
	DeleteAt     string
	Mtime        string
}

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

	if DebugOutput {
		log.Printf("Worker%d File %d/%d: started process file %s\n", id, j.FileNumber, j.FileAllCount, j.File)
	}

	for _, to := range j.EnvTo {
		// skip if file already on the replication region present
		if stringInSlice(j.File, to.Files) {
			continue
		}

		err := retry(5, 2*time.Second, func() (err error) {

			//query the file metadata at the target
			_, headers, err := to.SwiftCli.Object(
				backupContainer,
				j.File,
			)
			if err != nil {
				if err == swift.ObjectNotFound {
					headers = swift.Headers{}
				} else {
					//log all other errors and skip the file (we don't want to waste
					//bandwidth downloading stuff if there is reasonable doubt that we will
					//not be able to upload it to Swift)
					log.Printf("Worker%d File %d/%d: skipping target %s %s/%s: HEAD failed: %s",
						id, j.FileNumber, j.FileAllCount, j.EnvFrom.Cfg.OsRegionName, backupContainer, j.File,
						err.Error(),
					)
					return nil
				}
			}

			//retrieve object from source, taking advantage of Etag and Last-Modified where possible
			metadata := headers.ObjectMetadata()
			targetState := FileState{
				Etag:         metadata["source-etag"],
				LastModified: metadata["source-last-modified"],
			}
			body, sourceState, err := GetFile(&j, j.File, targetState)
			if err != nil {
				log.Println(err.Error())
				return err
			}
			if body != nil {
				defer body.Close()
			}
			if sourceState.SkipTransfer { // 304 Not Modified
				return nil
			}

			//store some headers from the source to later identify whether this
			//resource has changed
			metadata = make(swift.Metadata)
			if sourceState.Etag != "" {
				metadata["source-etag"] = sourceState.Etag
			}
			if sourceState.LastModified != "" {
				metadata["source-last-modified"] = sourceState.LastModified
			}
			if sourceState.Mtime != "" {
				metadata["source-mtime"] = sourceState.Mtime
			}

			newHeaders := metadata.ObjectHeaders()
			if sourceState.DeleteAt != "" {
				newHeaders["X-Delete-After"] = "864000"
			}
			//upload file to target
			_, err = to.SwiftCli.ObjectPut(
				backupContainer,
				j.File,
				body,
				false, "",
				sourceState.ContentType,
				newHeaders,
			)
			if err != nil {
				log.Printf("Worker%d File %d/%d: PUT %s %s/%s failed: %s", id, j.FileNumber, j.FileAllCount, to.Cfg.OsRegionName, backupContainer, j.File, err.Error())

				//delete potentially incomplete upload
				err := to.SwiftCli.ObjectDelete(
					backupContainer,
					j.File,
				)
				if err != nil {
					log.Printf("Worker%d File %d/%d: DELETE %s %s/%s failed: %s", id, j.FileNumber, j.FileAllCount, to.Cfg.OsRegionName, backupContainer, j.File, err.Error())
				}
				return err
			}
			return nil
		})

		// error for retry
		if err != nil {
			log.Printf("Worker%d File %d/%d: PUT %s %s/%s with retry failed: %s", id, j.FileNumber, j.FileAllCount, to.Cfg.OsRegionName, backupContainer, j.File, err.Error())
			PromGauge.SetError()
			return
		}
	}

	if DebugOutput {
		log.Printf("Worker%d File %d/%d: completed %s\n", id, j.FileNumber, j.FileAllCount, j.File)
	}

	currentFilesDone += 1
	PromGauge.CurrentFile(currentFilesDone)
	num := int(Round((float64(currentFilesDone) / float64(j.FileAllCount)) * 100.0))
	lockAlreadyPrinted.Lock()
	if 100 <= alreadyPrinted {
		lockAlreadyPrinted.Unlock()
		lockAlreadyPrinted.RLock()
		alreadyPrinted = 0
		lockAlreadyPrinted.RUnlock()
	}
	lockAlreadyPrinted.Lock()
	if 100 <= alreadyPrinted {
		lockAlreadyPrinted.Unlock()
		lockAlreadyPrinted.RLock()
		alreadyPrinted += 5
		lockAlreadyPrinted.RUnlock()
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

	log.Printf("End replication task in %g seconds\n", time.Since(startTime).Seconds())
}

func LoadAndStartJobs() {
	var err error

	if !initialized {
		// cfg used for the parsed YAML Configuration
		cfg := configuration.YAMLReplication("/backup/env/config.yml")

		expiration = os.Getenv("BACKUP_EXPIRE_AFTER")
		if expiration == "" {
			expiration = "864000"
		}

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
			EnvFrom.Cfg.OsRegionName)
		if err != nil {
			log.Println("Error can't connect swift for", EnvFrom.Cfg.OsRegionName, err)
			PromGauge.SetError()
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
				EnvTo[id].Cfg.OsRegionName)
			if err != nil {
				log.Println("Error can't connect swift for", EnvTo[id].Cfg.OsRegionName, err)
				PromGauge.SetError()
				return
			}
		}

		initialized = true
	}

	EnvFrom.Files, err = swiftcli.SwiftListPrefixFiles(EnvFrom.SwiftCli, EnvFrom.Cfg.ContainerPrefix)
	if err != nil {
		log.Println("Error fet files for", EnvFrom.Cfg.OsRegionName, err)
		PromGauge.SetError()
		return
	}

	for id := range EnvTo {
		EnvTo[id].Files, err = swiftcli.SwiftListPrefixFiles(EnvTo[id].SwiftCli, EnvTo[id].Cfg.ContainerPrefix)
		if err != nil {
			log.Println("Error get files for", EnvTo[id].Cfg.OsRegionName, err)
			PromGauge.SetError()
			return
		}
	}

	// Set all to false for a new loop as default
	lockAlreadyPrinted.RLock()
	alreadyPrinted = 0
	lockAlreadyPrinted.RUnlock()

	// Start Job Worker
	StartJobWorkers()

	// Set Success Prometheus
	PromGauge.SetSuccess(nil)
	return
}

// GetFile is the function with that we get the content information to skip this file or an error
func GetFile(job *Job, filePath string, targetState FileState) (io.ReadCloser, FileState, error) {
	reqHeaders := make(swift.Headers)
	if targetState.Etag != "" {
		reqHeaders["If-None-Match"] = targetState.Etag
	}
	if targetState.LastModified != "" {
		reqHeaders["If-Modified-Since"] = targetState.LastModified
	}

	body, respHeaders, err := job.EnvFrom.SwiftCli.ObjectOpen(backupContainer, filePath, false, reqHeaders)
	switch err {
	case nil:
		return body, FileState{
			Etag:         respHeaders["Etag"],
			LastModified: respHeaders["Last-Modified"],
			ContentType:  respHeaders["Content-Type"],
			DeleteAt:     respHeaders["X-Delete-At"],
			Mtime:        respHeaders["X-Object-Meta-Mtime"],
		}, nil
	case swift.NotModified:
		return nil, FileState{SkipTransfer: true}, nil
	default:
		return nil, FileState{}, err
	}
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

func retry(attempts int, sleep time.Duration, callback func() error) (err error) {
	for i := 0; ; i++ {
		err = callback()
		if err == nil {
			return
		}

		if i >= (attempts - 1) {
			break
		}

		time.Sleep(sleep)

		log.Println("retrying after error:", err)
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}
