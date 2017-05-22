package main

import (
    "bufio"
    "errors"
    "fmt"
    "io"
    "os"
    "path"
    "path/filepath"
    "strconv"
    "strings"

    "github.com/ncw/swift"
)

var (
    client *swift.Connection
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
) {

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

    return
}

// SwiftListFiles swift list all files in container
func SwiftListFiles() ([]string, error) {
    objects := make([]string, 0)
    err := client.ObjectsWalk(containerName, nil, func(opts *swift.ObjectsOpts) (interface{}, error) {
        newObjects, err := client.ObjectNames(containerName, opts)
        if err == nil {
            objects = append(objects, newObjects...)
        }
        return newObjects, err
    })
    fmt.Println("Found all the objects", objects, err)
    return objects, err
}

// SwiftListPrefixFiles swift list all files in container with this prefix
func SwiftListPrefixFiles(prefix string) ([]string, error) {
    opts := swift.ObjectsOpts{
        Prefix: prefix,
    }
    objects := make([]string, 0)
    err := client.ObjectsWalk(containerName, &opts, func(opts *swift.ObjectsOpts) (interface{}, error) {
        newObjects, err := client.ObjectNames(containerName, opts)
        if err == nil {
            objects = append(objects, newObjects...)
        }
        return newObjects, err
    })
    fmt.Println("Found all the prefix objects", objects, err)
    return objects, err
}

// SwiftDownloadFile swift download this file from container
func SwiftDownloadFile(file string) error {
    var w io.Writer
    var bw *bufio.Writer

    mypath := filepath.Join(backupPath, path.Base(file))
    outFile, err := os.Create(mypath)
    if err != nil {
        return err
    }
    bw = bufio.NewWriter(outFile)
    w = bw
    defer outFile.Close()
    defer bw.Flush()

    _, err = client.ObjectGet(containerName, file, w, false, nil)
    return err
}

//SwiftDownloadPrefix swift download all files from container that start with this prefix
func SwiftDownloadPrefix(prefix string) error {
    list, err := SwiftListPrefixFiles(prefix)
    if err != nil {
        return err
    }
    if len(list) == 0 {
        return errors.New("Prefix List empty")
    }

    // list of localfiles
    //objects := make([]string, 0)

    // TODO: download files via SwiftDownloadFile and add file to objects
    for _, str := range list {
        err := SwiftDownloadFile(str)
        if err != nil {
            return err
        }
    }
    return nil
}

//UnpackFiles Unpack files like .tar.gz and .gz
func UnpackFiles(files []string) error {
    for _, file := range files {

        if strings.HasSuffix(file, ".tar.gz") {

            ungzip(file, strings.TrimSuffix(file, ".gz"))
            untar(strings.TrimSuffix(file, ".gz"), strings.TrimSuffix(file, ".tar"))

        } else if strings.HasSuffix(file, ".gz") {

            ungzip(file, strings.TrimSuffix(file, ".gz"))

        } else {
            return errors.New("Unknown archive - must be .tar.gz or .gz file")
        }
    }
    return nil
}
