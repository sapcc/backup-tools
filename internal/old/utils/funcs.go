package utils

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/xtgo/set"
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

func DeleteEmpty(s []string) []string {
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

func DeleteNoGzSuffix(s []string) []string {
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

func MakePrefixPathOnly(s []string) []string {
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
func LeftPad(str string, len int, pad string) string {
	return times(pad, len-utf8.RuneCountInString(str)) + str
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
