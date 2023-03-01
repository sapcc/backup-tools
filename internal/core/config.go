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

package core

import (
	"fmt"
	"os"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/utils/openstack/clientconfig"
	"github.com/majewsky/schwift"
	"github.com/majewsky/schwift/gopherschwift"
	"github.com/sapcc/go-bits/osext"
)

// Configuration contains all the configuration parameters that we read from
// the process environment on startup.
type Configuration struct {
	//configuraton for upload to/download from Swift
	Container        *schwift.Container
	SegmentContainer *schwift.Container
	ObjectNamePrefix string
	//backup schedule
	Interval time.Duration
	//HTTP server configuration
	ListenAddress string
	//configuration for connection to Postgres
	PgHostname string
	PgUsername string
}

// NewConfiguration reads all configuration parameters from the process
// environment.
func NewConfiguration() (*Configuration, error) {
	//initialize connection to Swift
	ao, err := clientconfig.AuthOptions(nil)
	if err != nil {
		return nil, fmt.Errorf("cannot find OpenStack credentials: %w", err)
	}
	ao.AllowReauth = true
	provider, err := openstack.AuthenticatedClient(*ao)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to OpenStack: %w", err)
	}
	eo := gophercloud.EndpointOpts{
		//note that empty values are acceptable in these two fields (but OS_REGION_NAME is strictly required down below)
		Region:       os.Getenv("OS_REGION_NAME"),
		Availability: gophercloud.Availability(os.Getenv("OS_INTERFACE")),
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
		ListenAddress: osext.GetenvOrDefault("BACKUP_METRICS_LISTEN_ADDRESS", ":9188"),
		PgHostname:    osext.GetenvOrDefault("PGSQL_HOST", "localhost"),
		PgUsername:    osext.GetenvOrDefault("PGSQL_USER", "postgres"),
	}

	//read additional environment variables
	cfg.Interval, err = time.ParseDuration(osext.MustGetenv("BACKUP_PGSQL_FULL"))
	if err != nil {
		return nil, fmt.Errorf("malformed value for BACKUP_PGSQL_FULL: %q", os.Getenv("BACKUP_PGSQL_FULL"))
	}

	return &cfg, nil
}
