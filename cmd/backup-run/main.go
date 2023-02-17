package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/sapcc/containers/internal/prometheus"
)

func main() {
	bp := prometheus.NewBackup()
	go func() {
		for {
			command := exec.Command("/usr/local/sbin/db-backup.sh")
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr
			bp.Beginn()

			t, err := ioutil.ReadFile("/tmp/last_backup_timestamp")
			if err != nil {
				fmt.Println(err)
			}
			rx := regexp.MustCompile(`^([0-9]{4})([0-9]{2})([0-9]{2})([0-9]{2})([0-9]{2})$`)
			ts := rx.ReplaceAllString(strings.Trim(string(t), "\n"), "$1-$2-$3 $4:$5:00")

			layout := "2006-01-02 15:04:05"
			timestamp, _ := time.Parse(layout, ts)

			if err := command.Run(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				bp.SetError()
			} else {
				bp.SetSuccess(&timestamp)
			}
			bp.Finish()
			time.Sleep(300 * time.Second)
		}
	}()

	bp.ServerStart()
}
