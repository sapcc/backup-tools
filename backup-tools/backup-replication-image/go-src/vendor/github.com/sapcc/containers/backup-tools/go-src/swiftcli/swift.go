package swiftcli // import "github.com/sapcc/containers/backup-tools/go-src/swiftcli"

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ncw/swift"
	"github.com/sapcc/containers/backup-tools/go-src/configuration"
	"github.com/sapcc/containers/backup-tools/go-src/utils"
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

//SwiftDownloadPrefix swift download all files from container that start with this prefix
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

// SwiftUploadFile swift upload a file to the container
func SwiftUploadFile(clientSwift *swift.Connection, file string, expireAfter *int64, fakeObjectName *string) (done bool, err error) {
	var realPathFile string

	realPathFile, err = filepath.Abs(file)
	if err != nil {
		return false, err
	}

	contentType, md5hash, fileSize, err := GetMimeTypeAndMd5AndSize(realPathFile)
	if err != nil {
		return false, err
	}

	headers := swift.Headers{}

	// Now set content size correctly
	headers["Content-Length"] = strconv.FormatInt(fileSize, 10)
	if expireAfter != nil {
		headers["X-Delete-After"] = strconv.FormatInt(*expireAfter, 10)
	}

	fileOS, err := os.Open(realPathFile)
	if err != nil {
		return false, err
	}
	defer fileOS.Close()
	contents := bufio.NewReader(fileOS)

	if fakeObjectName != nil {
		realPathFile = *fakeObjectName
	}

	if strings.HasPrefix(realPathFile, configuration.DefaultConfiguration.ContainerPrefix) {
		realPathFile = strings.TrimPrefix(realPathFile, configuration.DefaultConfiguration.ContainerPrefix)
	}

	pathSlice := []string{configuration.DefaultConfiguration.ContainerPrefix}
	pathSlice = append(pathSlice, realPathFile)
	objectName := path.Clean(strings.Join(pathSlice, string(os.PathSeparator)))

	objectName = strings.TrimLeft(objectName, "/")

	h, err := clientSwift.ObjectPut(configuration.ContainerName, objectName, contents, true, md5hash, contentType, headers)
	if err != nil {
		return false, err
	}

	if h["Etag"] != md5hash {
		err = fmt.Errorf("Bad Etag want %q got %q", md5hash, h["Etag"])
		return false, err
	}

	// Fetch object info and compare
	info, _, err := clientSwift.Object(configuration.ContainerName, objectName)
	if err != nil {
		return false, err
	}
	if info.ContentType != contentType {
		err = fmt.Errorf("Bad ContentType want %q got %q", contentType, info.ContentType)
		return false, err
	}
	if info.Bytes != fileSize {
		err = fmt.Errorf("Bad file size want %q got %q", fileSize, info.Bytes)
		return false, err
	}
	if info.Hash != md5hash {
		err = fmt.Errorf("Bad hash want %q got %q", md5hash, info.Hash)
		return false, err
	}
	return true, nil
}

//SwiftUploadPrefix swift upload all files from system that start with this prefix
/*
func SwiftUploadPrefix(clientSwift *swift.Connection, prefix string) ([]string, error) {
	list, err := SwiftListPrefixFiles(clientSwift, prefix)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, errors.New("Prefix List empty")
	}

	// list of localfiles
	objects := make([]string, 0)

	// TODO: upload files via SwiftUploadFile and add file to objects
	for _, str := range list {
		if strings.HasPrefix(str, "mysql.") {
			continue
		}
		file, err := SwiftDownloadFile(clientSwift, str, nil, true)
		if err != nil {
			return nil, err
		}
		objects = append(objects, file)
	}
	return objects, nil
}
*/

//UnpackFiles Unpack files like .tar.gz and .gz
func UnpackFiles(files []string) error {
	for _, file := range files {
		if strings.HasSuffix(file, ".tar.gz") {
			err := utils.Ungzip(file, utils.BackupPath)
			if err != nil {
				log.Println("ungzip", file, utils.BackupPath)
				log.Fatal(err)
			}
			err = utils.UntarSplit(strings.TrimSuffix(file, ".gz"), utils.BackupPath)
			if err != nil {
				log.Println("untarSplit", strings.TrimSuffix(file, ".gz"), utils.BackupPath)
				log.Fatal(err)
			}
			os.Remove(file)
			os.Remove(strings.TrimSuffix(file, ".gz"))
		} else if strings.HasSuffix(file, ".gz") {
			err := utils.Ungzip(file, utils.BackupPath)
			if err != nil {
				log.Println(file, utils.BackupPath)
				log.Fatal(err)
			}
			os.Remove(file)
		}
	}
	return nil
}

//GetMimeTypeAndMd5AndSize returns the content type, md5 hash, size and error if an error is available...
func GetMimeTypeAndMd5AndSize(filePath string) (contentType string, md5Hash string, size int64, err error) {
	contentType = "application/octet-stream"
	path, err := filepath.Abs(filePath)
	if err != nil {
		return contentType, "", -1, err
	}

	file, err := os.Open(path)
	if err != nil {
		return contentType, "", -1, err
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return contentType, "", -1, err
	}
	size = fileStat.Size()

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return contentType, "", size, err
	}

	// Reset the read pointer if necessary.
	file.Seek(0, 0)

	// Always returns a valid content-type and "application/octet-stream" if no others seemed to match.
	contentType = http.DetectContentType(buffer)

	//Open a new hash interface to write to
	hash := md5.New()

	//Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		return contentType, "", size, err
	}

	//Get the 16 bytes hash
	hashInBytes := hash.Sum(nil)[:16]

	//Convert the bytes to a string
	returnMD5String := hex.EncodeToString(hashInBytes)

	return contentType, returnMD5String, size, nil

}
