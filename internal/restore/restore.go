// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package restore

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"slices"

	"github.com/kballard/go-shellquote"
	"github.com/sapcc/go-bits/logg"

	"github.com/sapcc/backup-tools/internal/core"
)

// SuperUserCredentials can be used to override the default credentials during Restore().
type SuperUserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Restore downloads and restores this backup into the Postgres.
func (bkp RestorableBackup) Restore(ctx context.Context, cfg *core.Configuration, suCreds *SuperUserCredentials) error {
	// download dumps
	dirPath := "/tmp/restore-" + bkp.ID
	err := os.MkdirAll(dirPath, 0777)
	if err != nil {
		return err
	}
	filePaths, err := bkp.DownloadTo(ctx, dirPath, cfg)
	if err != nil {
		return err
	}

	// override config if necessary
	if suCreds != nil {
		cloned := *cfg
		cloned.PgUsername = suCreds.Username
		cfg = &cloned
	}

	// playback dumps
	for _, filePath := range filePaths {
		cmd := exec.Command("psql", cfg.ArgsForPsql("-a", "-f", filePath)...) //nolint:gosec // input is user supplied and self executed
		logg.Info(">> " + shellquote.Join(cmd.Args...))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if suCreds != nil {
			cmd.Env = append(slices.Clone(os.Environ()), "PGPASSWORD="+suCreds.Password)
		}
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("could not import %s with psql: %w", filePath, err)
		}
	}

	return nil
}
