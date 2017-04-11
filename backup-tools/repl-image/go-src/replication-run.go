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
		command := exec.Command(cmd)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		if err := command.Run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		time.Sleep(14400 * time.Second)
	}
}
