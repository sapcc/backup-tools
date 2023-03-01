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

package backup

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kballard/go-shellquote"
	"github.com/majewsky/schwift"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sapcc/containers/internal/core"
	"github.com/sapcc/go-bits/logg"
)

var backupLastSuccessGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "backup_last_success",
	Help: "Unix Timestamp of last successful backup run",
})

func init() {
	prometheus.MustRegister(backupLastSuccessGauge)
}

const (
	// A time format corresponding to YYYYMMDDHHMM without any separator chars.
	TimeFormat = "200601021504"
	// Segment size for Swift uploads
	uploadSegmentSize = 2 << 30 // 2 GiB
	// Retention time for backups.
	retentionTime time.Duration = 10 * 24 * time.Hour // 10 days
)

// Create creates a backup unconditionally. The provided `reason` is used
// in log messages to explain why the backup was created.
func Create(cfg *core.Configuration, reason string) (returnedError error) {
	//track metrics for this backup
	nowTime := time.Now()
	nowTimeStr := nowTime.Format(TimeFormat)
	defer func() {
		if returnedError == nil {
			backupLastSuccessGauge.Set(float64(nowTime.Unix()))
		}
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
	err = WriteLastBackupTimestamp(cfg, nowTime)
	if err != nil {
		return err
	}

	logg.Info(">> done")
	return nil
}

// CreateIfNecessary creates a backup if the schedule demands it.
func CreateIfNecessary(cfg *core.Configuration) error {
	//check last_backup_timestamp to see if a backup is needed
	lastTime, err := ReadLastBackupTimestamp(cfg)
	if err != nil {
		return err
	}
	if lastTime.Add(cfg.Interval).After(time.Now()) {
		//even if there is no work to do, we update the backup_last_success metric
		//to an accurate value (this is required after application startup to have
		//a useful metric value before the first scheduled backup)
		backupLastSuccessGauge.Set(float64(lastTime.Unix()))
		return nil
	}

	return Create(cfg, "because of schedule")
}
