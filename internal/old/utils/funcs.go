package utils

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// BackupPath global const
	BackupPath = "/backup"
	// NewBackupPath global const
	NewBackupPath = "/newbackup"
)

var (
	// List global var
	List []string

	// List2 global var
	List2 []string
)

func FilterGzFiles(in []string) (out []string) {
	for _, str := range in {
		if strings.HasSuffix(str, ".gz") {
			out = append(out, str)
		}
	}
	return
}

func SortedUniqDirnames(in []string) (out []string) {
	seen := make(map[string]bool)
	for _, str := range in {
		str = filepath.Dir(str)
		if !seen[str] {
			out = append(out, str)
		}
		seen[str] = true
	}
	sort.Strings(out)
	return
}

func Ungzip(source, target string) error {
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
