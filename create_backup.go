/*******************************************************************************
*
* Copyright 2023 SAP SE
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You should have received a copy of the License along with this
* program. If not, you may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*
*******************************************************************************/

package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/kballard/go-shellquote"
	"github.com/majewsky/schwift"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sapcc/containers/internal/core"
	"github.com/sapcc/containers/internal/prometheus"
	"github.com/sapcc/go-bits/httpext"
	"github.com/sapcc/go-bits/logg"
	"github.com/sapcc/go-bits/must"
)

func commandCreateBackup() {
	ctx := httpext.ContextWithSIGINT(context.Background(), 1*time.Second)
	cfg := must.Return(core.NewConfiguration())

	//listen to SIGUSR1 (this signal causes a backup to be created immediately, regardless of schedule)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGUSR1)

	//fork off the main loop
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := createBackupIfNecessary(cfg)
				if err != nil {
					logg.Error(err.Error())
				}
			case <-signalChan:
				err := createBackup(cfg, "because of user request (SIGUSR1)")
				if err != nil {
					logg.Error(err.Error())
				}
			}
		}
	}()

	//serve Prometheus metrics on the main thread
	prometheus.InitMetrics()
	http.Handle("/metrics", promhttp.Handler())
	must.Succeed(httpext.ListenAndServeContext(ctx, cfg.ListenAddress, nil))

	//on SIGINT/SIGTERM, give the backup main loop a chance to complete a backup that's currently in flight
	wg.Wait()
}

const (
	// A time format corresponding to YYYYMMDDHHMM without any separator chars.
	backupTimeFormat = "200601021504"
	// Segment size for Swift uploads
	uploadSegmentSize = 2 << 30 // 2 GiB
	// Retention time for backups.
	retentionTime time.Duration = 10 * 24 * time.Hour // 10 days
)

func createBackupIfNecessary(cfg *core.Configuration) error {
	//check last_backup_timestamp to see if a backup is needed
	lastTimeObj := cfg.Container.Object(cfg.ObjectNamePrefix + "last_backup_timestamp")
	lastTime, err := readLastBackupTimestamp(lastTimeObj)
	if err != nil {
		return err
	}
	if lastTime.Add(cfg.Interval).After(time.Now()) {
		prometheus.SetSuccess(lastTime)
		return nil
	}

	return createBackup(cfg, "because of schedule")
}

func readLastBackupTimestamp(obj *schwift.Object) (time.Time, error) {
	str, err := obj.Download(nil).AsString()
	if err != nil {
		if schwift.Is(err, http.StatusNotFound) {
			//this branch is esp. relevant for the first ever backup -> we just report a very old last backup to force a backup immediately
			return time.Unix(0, 0).UTC(), nil
		}
		return time.Time{}, err
	}
	t, err := time.Parse(backupTimeFormat, str)
	if err != nil {
		//recover from malformed timestamp files by forcing a new backup immediately, same as above
		return time.Unix(0, 0).UTC(), nil
	}
	return t, nil
}

func createBackup(cfg *core.Configuration, reason string) (returnedError error) {
	//track metrics for this backup
	nowTime := time.Now()
	nowTimeStr := nowTime.Format(backupTimeFormat)
	prometheus.Begin()
	defer func() {
		if returnedError == nil {
			prometheus.SetSuccess(nowTime)
		} else {
			prometheus.SetError()
		}
		prometheus.Finish()
	}()
	logg.Info("creating backup %s%s %s...", cfg.ObjectNamePrefix, nowTimeStr, reason)

	//enumerate databases that need to be backed up
	cmd := exec.Command("psql",
		"-qAt", "-h", cfg.PgHostname, "-U", cfg.PgUsername, "-c", //NOTE: PGPASSWORD comes via inherited env variable
		`SELECT datname FROM pg_database WHERE datname !~ '^template|^postgres$'`)
	logg.Info(">> " + shellquote.Join(cmd.Args...))
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("could not enumerate databases with psql: %w", err)
	}
	databaseNames := strings.Fields(string(output))

	//stream backups from pg_dump into Swift
	for _, databaseName := range databaseNames {
		//NOTE: We need to do this in two separate goroutines because we need to
		//Close() the writer side in order for LargeObject.Append() to return on
		//the reader side.

		//run pg_dump
		pipeReader, pipeWriter := io.Pipe()
		errChan := make(chan error, 1) //must be buffered to ensure that `pipewriter.Close()` runs immediately
		go func() {
			defer pipeWriter.Close()
			cmd := exec.Command("pg_dump",
				"-h", cfg.PgHostname, "-U", cfg.PgUsername, //NOTE: PGPASSWORD comes via inherited env variable
				"-c", "--if-exist", "-C", "-Z", "5", databaseName)
			logg.Info(">> " + shellquote.Join(cmd.Args...))
			cmd.Stdout = pipeWriter
			cmd.Stderr = os.Stderr
			errChan <- cmd.Run()
		}()

		//upload the outputs of pg_dump into Swift
		obj := cfg.Container.Object(cfg.ObjectNamePrefix + fmt.Sprintf("%s/backup/pgsql/base/%s.sql.gz", nowTimeStr, databaseName))
		lo, err := obj.AsNewLargeObject(schwift.SegmentingOptions{
			Strategy:         schwift.StaticLargeObject,
			SegmentContainer: cfg.SegmentContainer,
		}, nil)
		if err != nil {
			return fmt.Errorf("could not start upload into Swift: %w", err)
		}

		hdr := schwift.NewObjectHeaders()
		hdr.ExpiresAt().Set(nowTime.Add(retentionTime))
		err = lo.Append(pipeReader, uploadSegmentSize, hdr.ToOpts())
		if err != nil {
			return fmt.Errorf("could not write into Swift: %w", err)
		}

		err = lo.WriteManifest(hdr.ToOpts())
		if err != nil {
			return fmt.Errorf("could not finalize upload into Swift: %w", err)
		}

		//wait for pg_dump to finish
		err = <-errChan
		if err != nil {
			return fmt.Errorf("could not run pg_dump: %w", err)
		}
	}

	//write last_backup_timestamp to indicate that this backup is completed successfully
	lastTimeObj := cfg.Container.Object(cfg.ObjectNamePrefix + "last_backup_timestamp")
	err = lastTimeObj.Upload(strings.NewReader(nowTimeStr), nil, nil)
	if err != nil {
		return fmt.Errorf("could not write last_backup_timestamp into Swift: %w", err)
	}
	logg.Info(">> done")
	return nil
}
