/*******************************************************************************
*
* Copyright 2023 SAP SE
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You should have received a copy of the License along with this
* program. If not, you may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*
*******************************************************************************/

package main

import (
	"fmt"
	"os"

	"github.com/sapcc/go-bits/logg"
)

func usage() {
	fmt.Fprintf(os.Stderr, "USAGE: %s restore              - Restore from a backup stored in the regional Swift.\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "       %s restore-crossregion  - Restore from a backup stored in a different Swift.\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "       %s credentials          - Show credentials for a cross-regional restore.\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "       %s manual               - Restore from a backup file on the local disk.\n", os.Args[0])
}

func main() {
	switch len(os.Args) {
	case 1:
		usage()
		return
	case 2:
		switch os.Args[1] {
		case "create": //NOTE: This subcommand is not shown in usage() because it's not meant for interactive use.
			commandCreateBackup()
		case "restore":
			logg.Fatal("TODO")
		case "restore-crossregion":
			logg.Fatal("TODO")
		case "credentials":
			logg.Fatal("TODO")
		case "manual":
			logg.Fatal("TODO")
		default: //invalid subcommand
			usage()
			os.Exit(1)
		}
	default: //invalid number of args
		usage()
		os.Exit(1)
	}
}

func commandCreateBackup() {
	logg.Fatal("TODO")
}
