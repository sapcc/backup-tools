package configuration

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	// LongDateForm Date Format
	LongDateForm = "200601021504"
)

var (
	// ContainerName is used as Default Container
	ContainerName = "db_backup"
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

func init() {
	if dbtest := os.Getenv("BACKUP_TEST_CONTAINER"); "" != dbtest {
		ContainerName = dbtest
	}
}

// EnvironmentStruct structure for the export and import for usage with
// backup-restore crossregion
type EnvironmentStruct struct {
	ContainerPrefix      string `json:"cp,omitempty" yaml:"container_prefix,omitempty"`
	OsAuthURL            string `json:"oau,omitempty" yaml:"os_auth_url,omitempty"`
	OsAuthVersion        string `json:"oauv,omitempty" yaml:"os_auth_version,omitempty"`
	OsIdentityAPIVersion string `json:"oiav,omitempty" yaml:"identify_api_version,omitempty"`
	OsUsername           string `json:"ou,omitempty" yaml:"os_username,omitempty"`
	OsUserDomainName     string `json:"oud,omitempty" yaml:"os_user_domain,omitempty"`
	OsProjectName        string `json:"opn,omitempty" yaml:"os_project_name,omitempty"`
	OsProjectDomainName  string `json:"opdn,omitempty" yaml:"os_project_domain,omitempty"`
	OsRegionName         string `json:"orn,omitempty" yaml:"os_region_name,omitempty"`
	OsPassword           string `json:"op,omitempty" yaml:"os_password,omitempty"`
}

type EnvironmentYamlReplication struct {
	From EnvironmentStruct
	To   []EnvironmentStruct
}

func YAMLReplication(filename string) EnvironmentYamlReplication {
	c := EnvironmentYamlReplication{}
	// get the abs
	// which will try to find the 'filename' from current working dir too.
	yamlAbsPath, err := filepath.Abs(filename)
	if err != nil {
		panic(err)
	}

	// read the raw contents of the file
	data, err := ioutil.ReadFile(yamlAbsPath)
	if err != nil {
		panic(err)
	}

	// put the file's contents as yaml to the default configuration(c)
	if err := yaml.Unmarshal(data, &c); err != nil {
		panic(err)
	}

	return c
}

func init() {
	backupRegionName := os.Getenv("BACKUP_REGION_NAME")
	if len(backupRegionName) == 0 {
		backupRegionName = os.Getenv("OS_REGION_NAME")
	}
	DefaultConfiguration = &EnvironmentStruct{
		ContainerPrefix:      strings.Join([]string{backupRegionName, os.Getenv("MY_POD_NAMESPACE"), os.Getenv("MY_POD_NAME")}, "/"),
		OsAuthURL:            os.Getenv("OS_AUTH_URL"),
		OsAuthVersion:        os.Getenv("OS_AUTH_VERSION"),
		OsIdentityAPIVersion: os.Getenv("OS_IDENTITY_API_VERSION"),
		OsUsername:           os.Getenv("OS_USERNAME"),
		OsUserDomainName:     os.Getenv("OS_USER_DOMAIN_NAME"),
		OsProjectName:        os.Getenv("OS_PROJECT_NAME"),
		OsProjectDomainName:  os.Getenv("OS_PROJECT_DOMAIN_NAME"),
		OsRegionName:         os.Getenv("OS_REGION_NAME"),
		OsPassword:           os.Getenv("OS_PASSWORD"),
	}
}
