package main

import (
    "fmt"
    "os"

    "gopkg.in/urfave/cli.v1"
)

const (
    longForm      = "200601021504"
    containerName = "db_backup"
)

var list string
var list2 []string
var backupType string
var backupPath = "/newbackup"
var influxDBPath = "/var/lib/influx"

type environmentStruct struct {
    MyPodName            string `json:"mpn1,omitempty"`
    MyPodNamespace       string `json:"mpn2,omitempty"`
    OsAuthURL            string `json:"oau,omitempty"`
    OsAuthVersion        string `json:"oauv,omitempty"`
    OsIdentityAPIVersion string `json:"oiav,omitempty"`
    OsUsername           string `json:"ou,omitempty"`
    OsUserDomainName     string `json:"oud,omitempty"`
    OsProjectName        string `json:"opn,omitempty"`
    OsProjectDomainName  string `json:"opdn,omitempty"`
    OsRegionName         string `json:"orn,omitempty"`
    OsPassword           string `json:"op,omitempty"`
    InfluxdbRootPassword string `json:"irp,omitempty"`
}

func init() {

    cli.AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
{{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .VisibleFlags}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}{{if .Copyright }}
COPYRIGHT:
   {{.Copyright}}
   {{end}}{{if .Version}}
VERSION:
   {{.Version}} at {{.Compiled}}
   {{end}}
`
    cli.VersionPrinter = func(c *cli.Context) {
        fmt.Fprintf(c.App.Writer, "version=%s build=%s\n", c.App.Version, c.App.Compiled)
    }

}

func main() {

    app := cli.NewApp()
    app.EnableBashCompletion = false
    app.Name = "backup-restore"
    app.Version = "1.0.0"
    app.Usage = "make backup-restore processes easy"
    app.UsageText = "make backup-restore processes easy"
    app.Authors = []cli.Author{
        {
            Name:  "Josef Fröhle",
            Email: "josef.froehle@sap.com",
        },
    }
    app.Copyright = "(c) 2017 Josef Fröhle (B1-Systems GmbH) for SAP SE"

    app.Commands = []cli.Command{
        {
            Name:    "influxconfig",
            Aliases: []string{"ic"},
            Usage:   "create a influxdb config",
            Action: func(c *cli.Context) error {

                return startInfluxInit()
            },
        },
    }

    app.Action = func(ctx *cli.Context) error {
        return startRestoreInit()
    }

    app.Run(os.Args)
}
