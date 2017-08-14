package configuration // import "github.com/sapcc/containers/backup-tools/go-src/configuration"

import (
	"os"
	"strings"

	"github.com/sapcc/containers/backup-tools/go-src/underscore"
)

const (
	// LongDateForm Date Format
	LongDateForm = "200601021504"
	// ContainerName is used as Default Container
	ContainerName = "db_backup"
)

var (
	// DefaultConfiguration variable automatic filled with ENV data
	DefaultConfiguration *EnvironmentStruct
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

func init() {
	DefaultConfiguration = EnvironmentStruct{
		ContainerPrefix:      strings.Join([]string{os.Getenv("BACKUP_REGION_NAME"), os.Getenv("MY_POD_NAMESPACE"), os.Getenv("MY_POD_NAME")}, "/"),
		OsAuthURL:            os.Getenv(strings.ToUpper(underscore.Underscore("OsAuthURL"))),
		OsAuthVersion:        os.Getenv(strings.ToUpper(underscore.Underscore("OsAuthVersion"))),
		OsIdentityAPIVersion: os.Getenv(strings.ToUpper(underscore.Underscore("OsIdentityAPIVersion"))),
		OsUsername:           os.Getenv(strings.ToUpper(underscore.Underscore("OsUsername"))),
		OsUserDomainName:     os.Getenv(strings.ToUpper(underscore.Underscore("OsUserDomainName"))),
		OsProjectName:        os.Getenv(strings.ToUpper(underscore.Underscore("OsProjectName"))),
		OsProjectDomainName:  os.Getenv(strings.ToUpper(underscore.Underscore("OsProjectDomainName"))),
		OsRegionName:         os.Getenv(strings.ToUpper(underscore.Underscore("OsRegionName"))),
		OsPassword:           os.Getenv(strings.ToUpper(underscore.Underscore("OsPassword"))),
	}
}
