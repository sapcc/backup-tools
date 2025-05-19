// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package backup

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/majewsky/schwift/v2"
	"github.com/sapcc/go-bits/errext"

	"github.com/sapcc/backup-tools/internal/core"
)

func lastBackupTimestampObj(cfg *core.Configuration) *schwift.Object {
	return cfg.Container.Object(cfg.ObjectNamePrefix + "last_backup_timestamp")
}

// ReadLastBackupTimestamp reads the "last_backup_timestamp" object in Swift to
// find when the most recent backup was created.
func ReadLastBackupTimestamp(ctx context.Context, cfg *core.Configuration) (time.Time, error) {
	var str string
	err := retryUpToThreeTimes(func() error {
		var err error
		str, err = lastBackupTimestampObj(cfg).Download(ctx, nil).AsString()
		if schwift.Is(err, http.StatusNotFound) {
			// for the first ever backup, this will force a new backup immediately down below
			str = ""
			return nil
		} else {
			return err
		}
	})
	if err != nil {
		return time.Time{}, fmt.Errorf("could not read last_backup_timestamp from Swift: %w", err)
	}

	t, err := time.ParseInLocation(TimeFormat, str, time.UTC)
	if err != nil {
		// recover from malformed timestamp files by forcing a new backup immediately, same as above
		return time.Unix(0, 0).UTC(), nil //nolint:nilerr // intended behaviour
	}
	return t, nil
}

// WriteLastBackupTimestamp updates the "last_backup_timestamp" object in Swift
// to indicate that a backup was completed successfully.
func WriteLastBackupTimestamp(ctx context.Context, cfg *core.Configuration, t time.Time) error {
	payload := strings.NewReader(t.UTC().Format(TimeFormat))
	err := retryUpToThreeTimes(func() error {
		return lastBackupTimestampObj(cfg).Upload(ctx, payload, nil, nil)
	})
	if err != nil {
		return fmt.Errorf("could not write last_backup_timestamp into Swift: %w", err)
	}
	return nil
}

// Retries fallible operations like Swift uploads/downloads up to three times to be more robust.
func retryUpToThreeTimes(action func() error) error {
	var errs errext.ErrorSet
	for {
		err := action()
		if err == nil {
			return nil
		} else {
			errs.Add(err)
		}

		if len(errs) == 3 {
			return errors.New(errs.Join(", "))
		}
		time.Sleep(1 * time.Second)
	}
}
