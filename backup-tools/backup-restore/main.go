package main

import (
    "fmt"
    "os"

    "gopkg.in/urfave/cli.v1"
)

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
