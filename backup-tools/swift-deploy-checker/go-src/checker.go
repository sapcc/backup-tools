package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/ncw/swift"
	"github.com/sapcc/containers/backup-tools/go-src/configuration"
)

const backupContainer = "db_backup_test"

var (
	EnvFrom       *Env
	EnvTo         = make([]*Env, 2)
	NotAllInSlice = false
)

type Env struct {
	Cfg      configuration.EnvironmentStruct
	Files    []swift.Object
	SwiftCli *swift.Connection
}

func main() {
	// cfg used for the parsed YAML Configuration
	cfg := configuration.YAMLReplication("/backup/env/config.yml")

	EnvFrom = &Env{Cfg: cfg.From}
	EnvFrom.Cfg.ContainerPrefix = EnvFrom.Cfg.OsRegionName
	var err error
	EnvFrom.SwiftCli, err = SwiftConnection(
		EnvFrom.Cfg.OsAuthVersion,
		EnvFrom.Cfg.OsAuthURL,
		EnvFrom.Cfg.OsUsername,
		EnvFrom.Cfg.OsPassword,
		EnvFrom.Cfg.OsUserDomainName,
		EnvFrom.Cfg.OsProjectName,
		EnvFrom.Cfg.OsProjectDomainName,
		EnvFrom.Cfg.OsRegionName)
	if err != nil {
		panic(fmt.Sprint("Error can't connect swift for", EnvFrom.Cfg.OsRegionName, err))
		os.Exit(1)
	}

	EnvFrom.Files, err = EnvFrom.SwiftCli.ObjectsAll(backupContainer, &swift.ObjectsOpts{})
	if err != nil {
		panic(fmt.Sprint("Error can't get files from swift for", EnvFrom.Cfg.OsRegionName, err))
		os.Exit(1)
	}

	// Create for each replication region an own Env
	for id, toConfig := range cfg.To {
		EnvTo[id] = &Env{Cfg: toConfig}
		EnvTo[id].Cfg.ContainerPrefix = EnvFrom.Cfg.OsRegionName
		EnvTo[id].SwiftCli, err = SwiftConnection(
			EnvTo[id].Cfg.OsAuthVersion,
			EnvTo[id].Cfg.OsAuthURL,
			EnvTo[id].Cfg.OsUsername,
			EnvTo[id].Cfg.OsPassword,
			EnvTo[id].Cfg.OsUserDomainName,
			EnvTo[id].Cfg.OsProjectName,
			EnvTo[id].Cfg.OsProjectDomainName,
			EnvTo[id].Cfg.OsRegionName)
		if err != nil {
			panic(fmt.Sprint("Error can't connect swift for", EnvTo[id].Cfg.OsRegionName, err))
			os.Exit(1)
		}

		EnvTo[id].Files, err = EnvTo[id].SwiftCli.ObjectsAll(backupContainer, &swift.ObjectsOpts{})

		if err != nil {
			panic(fmt.Sprint("Error can't get files from swift for", EnvTo[id].Cfg.OsRegionName, err))
			os.Exit(1)
		}
	}

	for _, obj := range EnvFrom.Files {

		if !stringInAllSlice(obj, EnvTo) {
			NotAllInSlice = true
			break
		}
	}

	if NotAllInSlice == false {
		os.Exit(0)
		return
	}

	os.Exit(1)
	return
}

func stringInAllSlice(a swift.Object, lists []*Env) bool {
	count := len(lists)
	found := 0
	for _, to := range lists {
		for _, b := range to.Files {
			if b.Name == a.Name && b.Hash == a.Hash && b.Bytes == a.Bytes {
				found += 1
				break
			}
		}
	}
	if count == found {
		return true
	}
	return false
}

func SwiftConnection(
	version,
	endpoint,
	username,
	password,
	userDomainName,
	projectName,
	projectDomainName,
	region string,
) (*swift.Connection, error) {

	vInt, _ := strconv.Atoi(version)
	// Create a connection
	//initialize Swift connection
	client := swift.Connection{
		AuthVersion:  vInt,
		AuthUrl:      endpoint,
		UserName:     username,
		Domain:       userDomainName,
		Tenant:       projectName,
		TenantDomain: projectDomainName,
		ApiKey:       password,
		Region:       region,
	}
	// Authenticate
	err := client.Authenticate()
	if err != nil {
		return nil, err
	}

	return &client, nil
}
