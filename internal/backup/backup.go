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
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kballard/go-shellquote"
	"github.com/majewsky/schwift/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sapcc/go-bits/logg"

	"github.com/sapcc/backup-tools/internal/core"
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

func getPgdumpForVersion(majorVersion string) string {
	return fmt.Sprintf("/usr/libexec/postgresql%s/pg_dump", majorVersion)
}

// Create creates a backup unconditionally. The provided `reason` is used
// in log messages to explain why the backup was created.
func Create(cfg *core.Configuration, reason string) (nowTime time.Time, returnedError error) {
	// track metrics for this backup
	nowTime = time.Now()
	nowTimeStr := nowTime.UTC().Format(TimeFormat)
	defer func() {
		if returnedError == nil {
			backupLastSuccessGauge.Set(float64(nowTime.Unix()))
		}
	}()
	logg.Info("creating backup %s%s %s...", cfg.ObjectNamePrefix, nowTimeStr, reason)

	// enumerate databases that need to be backed up
	query := `SELECT datname FROM pg_database WHERE datname !~ '^template|^postgres$'`
	cmd := exec.Command("psql", cfg.ArgsForPsql("-t", "-c", query)...) //nolint:gosec // input is user supplied and self executed
	logg.Info(">> " + shellquote.Join(cmd.Args...))
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return nowTime, fmt.Errorf("could not enumerate databases with psql: %w", err)
	}
	databaseNames := strings.Fields(string(output))

	// this context will be given to child processes to ensure that they get
	// cleaned up properly in case of unexpected errors (esp. network errors when
	// uploading to Swift)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// stream backups from pg_dump into Swift
	for _, databaseName := range databaseNames {
		//NOTE: We need to do this in two separate goroutines because we need to
		// Close() the writer side in order for LargeObject.Append() to return on
		// the reader side.

		// determine postgresql server version
		cmd := exec.CommandContext(ctx, "psql", //nolint:gosec // input is user supplied and self executed
			cfg.ArgsForPsql("--csv", "--tuples-only", "-c", "SHOW SERVER_VERSION")...) // output not decoration or padding
		cmd.Cancel = func() error { return cmd.Process.Signal(os.Interrupt) }
		cmd.WaitDelay = 3 * time.Second
		output, err := cmd.Output()
		if err != nil {
			return nowTime, fmt.Errorf("could not determine postgresql server version: %w", err)
		}

		majorVersion := strings.Split(string(output), ".")[0]
		pgdump := getPgdumpForVersion(majorVersion)

		// if the pgdump version was not found fallback to pgdump version 12
		if _, err := os.Stat(pgdump); errors.Is(err, os.ErrNotExist) {
			pgdump = getPgdumpForVersion("12")
		}

		// run pg_dump
		pipeReader, pipeWriter := io.Pipe()
		errChan := make(chan error, 1) // must be buffered to ensure that `pipewriter.Close()` runs immediately
		go func() {
			defer pipeWriter.Close()
			cmd := exec.CommandContext(ctx, pgdump,
				"--host", cfg.PgHostname,
				"--username", cfg.PgUsername, //NOTE: PGPASSWORD comes via inherited env variable
				"--compress", "5",
				"--clean", "--create", "--if-exist", "--no-privileges", databaseName)
			logg.Info(">> " + shellquote.Join(cmd.Args...))
			cmd.Cancel = func() error { return cmd.Process.Signal(os.Interrupt) }
			cmd.Stdout = pipeWriter
			cmd.Stderr = os.Stderr
			cmd.WaitDelay = 3 * time.Second
			errChan <- cmd.Run()
		}()

		// upload the outputs of pg_dump into Swift
		obj := cfg.Container.Object(cfg.ObjectNamePrefix + fmt.Sprintf("%s/backup/pgsql/base/%s.sql.gz", nowTimeStr, databaseName))
		lo, err := obj.AsNewLargeObject(ctx, schwift.SegmentingOptions{
			Strategy:         schwift.StaticLargeObject,
			SegmentContainer: cfg.SegmentContainer,
		}, nil)
		if err != nil {
			return nowTime, fmt.Errorf("could not start upload into Swift: %w", err)
		}

		hdr := schwift.NewObjectHeaders()
		hdr.ExpiresAt().Set(nowTime.Add(retentionTime))
		err = lo.Append(ctx, pipeReader, uploadSegmentSize, hdr.ToOpts())
		if err != nil {
			return nowTime, fmt.Errorf("could not write into Swift: %w", err)
		}

		// wait for pg_dump to finish before finalizing the object in Swift
		// (otherwise we might end up storing incomplete or broken backups)
		err = <-errChan
		if err != nil {
			return nowTime, fmt.Errorf("could not run pg_dump: %w", err)
		}

		err = lo.WriteManifest(ctx, hdr.ToOpts())
		if err != nil {
			return nowTime, fmt.Errorf("could not finalize upload into Swift: %w", err)
		}
	}

	// write last_backup_timestamp to indicate that this backup is completed successfully
	err = WriteLastBackupTimestamp(ctx, cfg, nowTime)
	if err != nil {
		return nowTime, err
	}

	logg.Info(">> done")
	return
}

// CreateIfNecessary creates a backup if the schedule demands it.
func CreateIfNecessary(ctx context.Context, cfg *core.Configuration) error {
	// check last_backup_timestamp to see if a backup is needed
	lastTime, err := ReadLastBackupTimestamp(ctx, cfg)
	if err != nil {
		return err
	}
	if lastTime.Add(cfg.Interval).After(time.Now()) {
		// even if there is no work to do, we update the backup_last_success metric
		// to an accurate value (this is required after application startup to have
		// a useful metric value before the first scheduled backup)
		backupLastSuccessGauge.Set(float64(lastTime.Unix()))
		return nil
	}

	_, err = Create(cfg, "because of schedule")
	return err
}
