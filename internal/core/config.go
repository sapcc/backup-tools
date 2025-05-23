// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/majewsky/schwift/v2"
	"github.com/majewsky/schwift/v2/gopherschwift"
	"github.com/sapcc/go-bits/gophercloudext"
	"github.com/sapcc/go-bits/osext"
)

// Configuration contains all the configuration parameters that we read from
// the process environment on startup.
type Configuration struct {
	// configuration for upload to/download from Swift
	Container        *schwift.Container
	SegmentContainer *schwift.Container
	ObjectNamePrefix string
	// backup schedule
	Interval time.Duration
	// configuration for connection to Postgres
	PgHostname string
	PgUsername string
	PgPassword string
}

// NewConfiguration reads all configuration parameters from the process
// environment.
func NewConfiguration(ctx context.Context) (*Configuration, error) {
	// initialize connection to Swift
	provider, eo, err := gophercloudext.NewProviderClient(ctx, nil)
	if err != nil {
		return nil, err
	}
	client, err := openstack.NewObjectStorageV1(provider, eo)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to Swift: %w", err)
	}
	account, err := gopherschwift.Wrap(client, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to Swift: %w", err)
	}

	cfg := Configuration{
		Container:        account.Container("db_backup"),
		SegmentContainer: account.Container("db_backup_segments"),
		ObjectNamePrefix: fmt.Sprintf("%s/%s/%s/",
			osext.MustGetenv("OS_REGION_NAME"),
			osext.MustGetenv("MY_POD_NAMESPACE"),
			osext.MustGetenv("MY_POD_NAME"),
		),
		PgHostname: osext.GetenvOrDefault("PGSQL_HOST", "localhost"),
		PgUsername: osext.GetenvOrDefault("PGSQL_USER", "postgres"),
		PgPassword: osext.MustGetenv("PGPASSWORD"),
	}

	// read additional environment variables
	cfg.Interval, err = time.ParseDuration(osext.MustGetenv("BACKUP_PGSQL_FULL"))
	if err != nil {
		return nil, fmt.Errorf("malformed value for BACKUP_PGSQL_FULL: %q", os.Getenv("BACKUP_PGSQL_FULL"))
	}

	return &cfg, nil
}

// ArgsForPsql prepends common options for psql to the given list of arguments.
// The arguments given to this method are specific to a particular psql
// invocation, and this function adds those that are always required.
func (cfg Configuration) ArgsForPsql(args ...string) []string {
	common := []string{
		"--variable", "ON_ERROR_STOP=1",
		"--quiet", "--no-align",
		"--host", cfg.PgHostname,
		"--username", cfg.PgUsername, //NOTE: PGPASSWORD comes via inherited env variable
		"--dbname", "postgres", // ensure that -d does not default to the app username
	}
	return append(common, args...)
}
