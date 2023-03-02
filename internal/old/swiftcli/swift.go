package swiftcli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ncw/swift"

	"github.com/sapcc/containers/internal/old/configuration"
	"github.com/sapcc/containers/internal/old/utils"
)

// SwiftConnection connect swift
// con = connection params
// cont = container name
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

// SwiftListPrefixFiles swift list all files in container with this prefix
func SwiftListPrefixFiles(clientSwift *swift.Connection, prefix string) ([]string, error) {
	opts := swift.ObjectsOpts{
		Prefix: prefix,
	}
	objects := make([]string, 0)
	err := clientSwift.ObjectsWalk(configuration.ContainerName, &opts, func(opts *swift.ObjectsOpts) (interface{}, error) {
		newObjects, err := clientSwift.ObjectNames(configuration.ContainerName, opts)
		if err == nil {
			objects = append(objects, newObjects...)
		}
		return newObjects, err
	})
	//fmt.Println("Found all the prefix objects", objects, err)
	return objects, err
}

// SwiftDownloadFile swift download this file from container
func SwiftDownloadFile(clientSwift *swift.Connection, file string, backupDir *string, useRealPath bool) (string, error) {
	var mypath string
	var w io.Writer
	var bw *bufio.Writer
	if backupDir == nil {
		*backupDir = utils.BackupPath
	}

	mypath = filepath.Join(*backupDir, path.Base(file))
	if useRealPath {
		mypath = path.Clean(filepath.Join(*backupDir, file))
	}

	os.MkdirAll(path.Dir(mypath), 0777)

	outFile, err := os.Create(mypath)
	if err != nil {
		return "", err
	}
	bw = bufio.NewWriter(outFile)
	w = bw
	defer outFile.Close()
	defer bw.Flush()

	_, err = clientSwift.ObjectGet(configuration.ContainerName, file, w, false, nil)
	return mypath, err
}

// SwiftDownloadPrefix swift download all files from container that start with this prefix
func SwiftDownloadPrefix(clientSwift *swift.Connection, prefix string, backupDir *string, useRealPath bool) ([]string, error) {
	list, err := SwiftListPrefixFiles(clientSwift, prefix)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, errors.New("Prefix List empty")
	}

	// list of localfiles
	objects := make([]string, 0)

	// TODO: download files via SwiftDownloadFile and add file to objects
	for _, str := range list {
		if strings.HasPrefix(str, "mysql.") {
			continue
		}
		file, err := SwiftDownloadFile(clientSwift, str, backupDir, useRealPath)
		if err != nil {
			return nil, err
		}
		objects = append(objects, file)
	}
	return objects, nil
}

// UnpackFiles Unpack files like .tar.gz and .gz
func UnpackFiles(files []string, targetDir string) error {
	if targetDir == "" {
		targetDir = utils.BackupPath
	}

	for _, file := range files {
		if strings.HasSuffix(file, ".tar.gz") {
			return fmt.Errorf("do not know how to unpack %s", file)
		} else if strings.HasSuffix(file, ".gz") {
			err := utils.Ungzip(file, targetDir)
			if err != nil {
				log.Println(file, targetDir)
				log.Fatal(err)
			}
			os.Remove(file)
		}
	}
	return nil
}
