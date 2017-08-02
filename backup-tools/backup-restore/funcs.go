package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/xtgo/set"
)

var (
	list       string
	list2      []string
	backupType string
	backupPath = "/newbackup"
)

func exeCmd(cmd string) string {
	//fmt.Println("command is ", cmd)
	// splitting head => g++ parts => rest of the command
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:]
	//fmt.Printf("in all caps: %q - %q\n", head, parts)

	cmdExec := exec.Command(head, parts...)
	var out bytes.Buffer
	cmdExec.Stdout = &out

	err := cmdExec.Run()
	if err != nil {
		log.Fatal(err, out.String())
	}
	//fmt.Printf("in all caps: %q\n", out.String())

	return out.String()
}

func exeCmdBashC(cmd string) string {
	//fmt.Println("command is ", cmd)
	// splitting head => g++ parts => rest of the command
	parts := "-c"
	head := "bash"
	//fmt.Printf("in all caps: %q - %q\n", head, parts)

	cmdExec := exec.Command(head, parts, cmd)
	var out bytes.Buffer
	cmdExec.Stdout = &out

	err := cmdExec.Run()
	if err != nil {
		log.Fatal(err, out.String())
	}
	//fmt.Printf("in all caps: %q\n", out.String())

	return out.String()
}

func deleteFile(path string) error {
	// delete file
	var err = os.Remove(path)
	return err
}

func deleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if strings.HasSuffix(str, "mysql.gz") {
			continue
		}
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func deleteNoGzSuffix(s []string) []string {
	var r []string
	for _, str := range s {
		if strings.HasSuffix(str, "mysql.gz") {
			continue
		}
		if strings.HasSuffix(str, ".gz") {
			r = append(r, str)
		}
	}
	return r
}

func makePrefixPathOnly(s []string) []string {
	var r []string
	for _, str := range s {
		r = append(r, filepath.Dir(str))
	}
	data := sort.StringSlice(r)
	sort.Sort(data)
	n := set.Uniq(data) // Uniq returns the size of the set
	data = data[:n]
	return data
}

func times(str string, n int) (out string) {
	for i := 0; i < n; i++ {
		out += str
	}
	return
}

// Left left-pads the string with pad up to len runes
// len may be exceeded if
func leftPad(str string, len int, pad string) string {
	return times(pad, len-utf8.RuneCountInString(str)) + str
}

func tarit(source, target string) error {
	filename := filepath.Base(source)
	target = filepath.Join(target, fmt.Sprintf("%s.tar", filename))
	tarfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	tarball := tar.NewWriter(tarfile)
	defer tarball.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	return filepath.Walk(source,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			var header *tar.Header
			header, err = tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			if baseDir != "" {
				header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
			}

			if err = tarball.WriteHeader(header); err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tarball, file)
			return err
		})
}

func untar(tarball, target string) error {
	return untar2Wrapped(tarball, target, false)
}

func untarSplit(tarball, target string) error {
	return untar2Wrapped(tarball, target, true)
}

func untar2Wrapped(tarball, target string, strip bool) error {
	reader, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		fname := header.Name
		if strip == true {
			fname = strings.TrimPrefix(fname, "/backup/"+backupType+"/base/")
			fname = strings.TrimPrefix(fname, "backup/"+backupType+"/base/")
		}
		path := filepath.Join(target, fname)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}
	return nil
}

func gzipit(source, target string) error {
	reader, err := os.Open(source)
	if err != nil {
		return err
	}

	filename := filepath.Base(source)
	target = filepath.Join(target, fmt.Sprintf("%s.gz", filename))
	writer, err := os.Create(target)
	if err != nil {
		return err
	}
	defer writer.Close()

	archiver := gzip.NewWriter(writer)
	archiver.Name = filename
	defer archiver.Close()

	_, err = io.Copy(archiver, reader)
	return err
}

func ungzip(source, target string) error {
	reader, err := os.Open(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	archive, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer archive.Close()

	var tfile string

	if archive.Name == "" {
		tfile = strings.TrimSuffix(source, ".gz")
	} else {
		tfile = archive.Name
	}

	target = filepath.Join(target, tfile)
	writer, err := os.Create(target)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, archive)
	return err
}
