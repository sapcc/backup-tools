package configuration // import "github.com/sapcc/containers/backup-tools/backup-restore/configuration"

import ()

const (
	// LongDateForm Date Format
	LongDateForm = "200601021504"
	// ContainerName is used as Default Container
	ContainerName = "db_backup"
)

var (
	// ContainerPrefix variable for internal usage
	ContainerPrefix string
	// AuthVersion variable for internal usage
	AuthVersion string
	// AuthEndpoint variable for internal usage
	AuthEndpoint string
	// AuthUsername variable for internal usage
	AuthUsername string
	// AuthPassword variable for internal usage
	AuthPassword string
	// AuthUserDomainName variable for internal usage
	AuthUserDomainName string
	// AuthProjectName variable for internal usage
	AuthProjectName string
	// AuthProjectDomainName variable for internal usage
	AuthProjectDomainName string
	// AuthRegion variable for internal usage
	AuthRegion string
	// MysqlRootPassword variable for internal usage
	MysqlRootPassword string
)

// EnvironmentStruct structure for the export and import for usage with
// backup-restore crossregion
type EnvironmentStruct struct {
	ContainerPrefix      string `json:"cp,omitempty"`
	OsAuthURL            string `json:"oau,omitempty"`
	OsAuthVersion        string `json:"oauv,omitempty"`
	OsIdentityAPIVersion string `json:"oiav,omitempty"`
	OsUsername           string `json:"ou,omitempty"`
	OsUserDomainName     string `json:"oud,omitempty"`
	OsProjectName        string `json:"opn,omitempty"`
	OsProjectDomainName  string `json:"opdn,omitempty"`
	OsRegionName         string `json:"orn,omitempty"`
	OsPassword           string `json:"op,omitempty"`
}
