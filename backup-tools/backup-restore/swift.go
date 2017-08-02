package main

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

	"github.com/sapcc/containers/backup-tools/backup-restore/configuration"

	"github.com/ncw/swift"
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
	region,
	contPrefix string,
) *swift.Connection {

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
		panic(err)
	}

	return &client
}

// SwiftListFiles swift list all files in container
func SwiftListFiles(clientSwift *swift.Connection) ([]string, error) {
	objects := make([]string, 0)
	err := clientSwift.ObjectsWalk(configuration.ContainerName, nil, func(opts *swift.ObjectsOpts) (interface{}, error) {
		newObjects, err := clientSwift.ObjectNames(configuration.ContainerName, opts)
		if err == nil {
			objects = append(objects, newObjects...)
		}
		return newObjects, err
	})
	//fmt.Println("Found all the objects", objects, err)
	return objects, err
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
func SwiftDownloadFile(clientSwift *swift.Connection, file string) (string, error) {
	var w io.Writer
	var bw *bufio.Writer

	fmt.Println("Download File: " + file)

	mypath := filepath.Join(backupPath, path.Base(file))
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

//SwiftDownloadPrefix swift download all files from container that start with this prefix
func SwiftDownloadPrefix(clientSwift *swift.Connection, prefix string) ([]string, error) {
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
		file, err := SwiftDownloadFile(clientSwift, str)
		if err != nil {
			return nil, err
		}
		objects = append(objects, file)
	}
	return objects, nil
}

//UnpackFiles Unpack files like .tar.gz and .gz
func UnpackFiles(files []string) error {
	for _, file := range files {
		if strings.HasSuffix(file, ".tar.gz") {

			err := ungzip(file, backupPath)
			if err != nil {
				log.Println("ungzip", file, backupPath)
				log.Fatal(err)
			}
			err = untarSplit(strings.TrimSuffix(file, ".gz"), backupPath)
			if err != nil {
				log.Println("untarSplit", strings.TrimSuffix(file, ".gz"), backupPath)
				log.Fatal(err)
			}
			defer os.Remove(file)
			defer os.Remove(strings.TrimSuffix(file, ".gz"))

		} else if strings.HasSuffix(file, ".gz") {

			err := ungzip(file, backupPath)
			if err != nil {
				log.Println(file, backupPath)
				log.Fatal(err)
			}
			defer os.Remove(file)

		}

	}
	return nil
}
