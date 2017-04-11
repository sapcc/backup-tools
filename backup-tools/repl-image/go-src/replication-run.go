package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

func main() {
	cmd := "/usr/local/sbin/backup-replication.sh"
	for {
		if err := exec.Command(cmd).Run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		time.Sleep(14400 * time.Second)
	}
}
