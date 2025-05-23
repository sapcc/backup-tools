// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package restore

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/sapcc/backup-tools/internal/backup"
	"github.com/sapcc/backup-tools/internal/core"
)

// RestorableBackup contains information about a backup that we found in Swift.
type RestorableBackup struct {
	ID             string   `json:"id"`
	ReadableTime   string   `json:"readable_time"`
	DatabaseNames  []string `json:"database_names"`
	TotalSizeBytes uint64   `json:"total_size_bytes"`
}

// RestorableBackups adds convenience functions on a slice of RestorableBackup.
type RestorableBackups []*RestorableBackup

// FindByID returns the backup with that id, or nil if no such backup is in
// this slice.
func (backups RestorableBackups) FindByID(id string) *RestorableBackup {
	for _, b := range backups {
		if b.ID == id {
			return b
		}
	}
	return nil
}

// ListRestorableBackups searches for restorable backups in Swift.
func ListRestorableBackups(ctx context.Context, cfg *core.Configuration) (RestorableBackups, error) {
	iter := cfg.Container.Objects()
	iter.Prefix = cfg.ObjectNamePrefix
	objInfos, err := iter.CollectDetailed(ctx)
	if err != nil {
		return nil, err
	}

	//NOTE: ObjectNamePrefix has a trailing slash
	rx := regexp.MustCompile(fmt.Sprintf(`^%s(\d{12})/backup/pgsql/base/([^.]*)\.sql\.gz$`, regexp.QuoteMeta(cfg.ObjectNamePrefix)))
	var result RestorableBackups
	for _, objInfo := range objInfos {
		// skip files not matching the above pattern (this especially skips the "last_backup_timestamp")
		match := rx.FindStringSubmatch(objInfo.Object.Name())
		if match == nil {
			continue
		}
		backupTimeStr, databaseName := match[1], match[2]
		backupTime, err := time.ParseInLocation(backup.TimeFormat, backupTimeStr, time.UTC)
		if err != nil {
			continue // treat malformed timestamp as "no match"
		}

		// do we already have an entry for this backup? (this can happen when a
		// Postgres contains multiple databases, since each database is backed up
		// into a separate Swift object)
		bkp := result.FindByID(backupTimeStr)
		if bkp == nil {
			result = append(result, &RestorableBackup{
				ID:             backupTimeStr,
				ReadableTime:   backupTime.Format(time.RFC1123),
				DatabaseNames:  []string{databaseName},
				TotalSizeBytes: objInfo.SizeBytes,
			})
		} else {
			bkp.DatabaseNames = append(bkp.DatabaseNames, databaseName)
			bkp.TotalSizeBytes += objInfo.SizeBytes
		}
	}

	return result, nil
}

// DownloadTo downloads and unzips the dumps belonging to this backup into the
// given directory on the local filesystem. The return value is the list of
// files that were written.
func (bkp RestorableBackup) DownloadTo(ctx context.Context, dirPath string, cfg *core.Configuration) ([]string, error) {
	var paths []string
	for _, databaseName := range bkp.DatabaseNames {
		path, err := bkp.downloadOneFile(ctx, dirPath, databaseName, cfg)
		if err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func (bkp RestorableBackup) downloadOneFile(ctx context.Context, dirPath, databaseName string, cfg *core.Configuration) (string, error) {
	// download from Swift
	objPath := fmt.Sprintf("%s/backup/pgsql/base/%s.sql.gz", bkp.ID, databaseName)
	obj := cfg.Container.Object(cfg.ObjectNamePrefix + objPath)
	reader, err := obj.Download(ctx, nil).AsReadCloser()
	if err != nil {
		return "", fmt.Errorf("could not GET %s: %w", obj.Name(), err)
	}
	defer reader.Close()

	// unpack gzip compression
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return "", fmt.Errorf("could not ungzip %s: %w", obj.Name(), err)
	}
	defer gzipReader.Close()

	// write to disk
	filePath := filepath.Join(dirPath, databaseName+".sql")
	writer, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer writer.Close()
	_, err = io.Copy(writer, gzipReader) //nolint:gosec // the archive is created by us and we must extract it completely
	if err != nil {
		return "", fmt.Errorf("could not ungzip %s: %w", obj.Name(), err)
	}
	return filePath, nil
}
