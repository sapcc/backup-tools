package main

import (
    "bufio"
    "bytes"
    b64 "encoding/base64"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "os/exec"
    "sort"
    "strconv"
    "strings"
    "time"
    "unicode/utf8"

    "./internal"

    "gopkg.in/urfave/cli.v1"
)

const longForm = "200601021504"

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

                if os.Getenv("INFLUXVER") != "" {
                    backupType = "influxdb"
                }

                if backupType == "" {
                    fmt.Println("\n\nNo System for the backup restore found.")
                    fmt.Println("\n\nFor Postgres and MariaDB/MySQL you can run this backup-restore programm in the \"backup\" container")
                    fmt.Println("Please RUN for Postgresql and MariaDB/MySQL\n# backup-restore\nin the backup-container")
                    fmt.Println("\n\n******** * * EXIT NO SUPPORTED SYSTEM FOUND * * ********")
                    return cli.NewExitError("-- E: 192902 --", 1)
                }

                group := environmentStruct{
                    MyPodName:            os.Getenv(strings.ToUpper(internal.Underscore("MyPodName"))),
                    MyPodNamespace:       os.Getenv(strings.ToUpper(internal.Underscore("MyPodNamespace"))),
                    OsAuthURL:            os.Getenv(strings.ToUpper(internal.Underscore("OsAuthURL"))),
                    OsAuthVersion:        os.Getenv(strings.ToUpper(internal.Underscore("OsAuthVersion"))),
                    OsIdentityAPIVersion: os.Getenv(strings.ToUpper(internal.Underscore("OsIdentityAPIVersion"))),
                    OsUsername:           os.Getenv(strings.ToUpper(internal.Underscore("OsUsername"))),
                    OsUserDomainName:     os.Getenv(strings.ToUpper(internal.Underscore("OsUserDomainName"))),
                    OsProjectName:        os.Getenv(strings.ToUpper(internal.Underscore("OsProjectName"))),
                    OsProjectDomainName:  os.Getenv(strings.ToUpper(internal.Underscore("OsProjectDomainName"))),
                    OsRegionName:         os.Getenv(strings.ToUpper(internal.Underscore("OsRegionName"))),
                    OsPassword:           os.Getenv(strings.ToUpper(internal.Underscore("OsPassword"))),
                    InfluxdbRootPassword: os.Getenv(strings.ToUpper(internal.Underscore("InfluxdbRootPassword"))),
                }

                data, err := json.Marshal(group)
                if err != nil {
                    fmt.Println("error:", err)
                }

                str := b64.StdEncoding.WithPadding(-1).EncodeToString(data)
                fmt.Println(str)

                //fmt.Println(strings.ToUpper(internal.Underscore("OsAuthURL")))

                return nil
            },
        },
    }

    app.Action = func(ctx *cli.Context) error {

        fmt.Println("Welcome to the Backup-Restore process!")
        fmt.Println("Please follow the instructions to the end to restore your backup.")
        fmt.Println("With the \"QUIT\" command on user-input requests, you can stop the backup process.")
        fmt.Println("\n\nPress 'Enter' to continue...")
        bufio.NewReader(os.Stdin).ReadBytes('\n')

        if os.Getenv("BACKUP_PGSQL_FULL") != "" {
            backupType = "pgsql"
        } else if os.Getenv("BACKUP_MYSQL_FULL") != "" {
            backupType = "mysql"
        } else if os.Getenv("INFLUXVER") != "" {
            backupType = "influxdb"
        }

        if backupType == "" {
            fmt.Println("\n\nNo System for the backup restore found.")
            fmt.Println("\n\nFor InfluxDB please RUN: # backup-restore ic")
            fmt.Println("With the generated command, you can start the backup process in the InfluxDB Container!")
            fmt.Println("\nfor Postgres and MariaDB/MySQL you can run this Backup-Restore programm in the \"backup\" container")
            fmt.Println("\n\n******** * * EXIT NO SUPPORTED SYSTEM FOUND * * ********")
            return cli.NewExitError("-- E: 1920291 --", 12)
        } else if backupType == "influxdb" {
        jumpToCJOFj30g2:

            reader := bufio.NewReader(os.Stdin)
            fmt.Print("Please enter the string from the backup container command \"backup-restore ic\" as a single-line: ")
            text, _ := reader.ReadString('\n')
            text = strings.TrimRight(text, "\n")
            if text == "" {
                goto jumpToCJOFj30g2
            }

            data, err := b64.StdEncoding.WithPadding(-1).DecodeString(text)
            if err != nil {
                fmt.Println("error:", err)
                return cli.NewExitError("-- Fatal Config Decoding E: 39303 --", 1)
            }
            //fmt.Printf("%q\n", data)
            var jsonReturn environmentStruct
            err = json.Unmarshal(data, &jsonReturn)
            if err != nil {
                return cli.NewExitError("-- Fatal Config Decoding E: 39343 --", 1)
            }
            fmt.Printf("%+v", jsonReturn)

            if jsonReturn.MyPodName != "" {
                os.Setenv(strings.ToUpper(internal.Underscore("MyPodName")), jsonReturn.MyPodName)
            }
            if jsonReturn.MyPodNamespace != "" {
                os.Setenv(strings.ToUpper(internal.Underscore("MyPodNamespace")), jsonReturn.MyPodNamespace)
            }
            if jsonReturn.OsAuthURL != "" {
                os.Setenv(strings.ToUpper(internal.Underscore("OsAuthURL")), jsonReturn.OsAuthURL)
            }
            if jsonReturn.OsAuthVersion != "" {
                os.Setenv(strings.ToUpper(internal.Underscore("OsAuthVersion")), jsonReturn.OsAuthVersion)
            }
            if jsonReturn.OsIdentityAPIVersion != "" {
                os.Setenv(strings.ToUpper(internal.Underscore("OsIdentityAPIVersion")), jsonReturn.OsIdentityAPIVersion)
            }
            if jsonReturn.OsUsername != "" {
                os.Setenv(strings.ToUpper(internal.Underscore("OsUsername")), jsonReturn.OsUsername)
            }
            if jsonReturn.OsUserDomainName != "" {
                os.Setenv(strings.ToUpper(internal.Underscore("OsUserDomainName")), jsonReturn.OsUserDomainName)
            }
            if jsonReturn.OsProjectName != "" {
                os.Setenv(strings.ToUpper(internal.Underscore("OsProjectName")), jsonReturn.OsProjectName)
            }
            if jsonReturn.OsProjectDomainName != "" {
                os.Setenv(strings.ToUpper(internal.Underscore("OsProjectDomainName")), jsonReturn.OsProjectDomainName)
            }
            if jsonReturn.OsRegionName != "" {
                os.Setenv(strings.ToUpper(internal.Underscore("OsRegionName")), jsonReturn.OsRegionName)
            }
            if jsonReturn.OsPassword != "" {
                os.Setenv(strings.ToUpper(internal.Underscore("OsPassword")), jsonReturn.OsPassword)
            }
            if jsonReturn.InfluxdbRootPassword != "" {
                os.Setenv(strings.ToUpper(internal.Underscore("InfluxdbRootPassword")), jsonReturn.InfluxdbRootPassword)
            }

        }
        /*
           fmt.Println("Original : ", intSlice[:])
           sort.Ints(intSlice)
           fmt.Println("Sort : ", intSlice)
           sort.Sort(sort.Reverse(sort.IntSlice(intSlice)))
           fmt.Println("Reverse Sort : ", intSlice)

            fmt.Printf("%q\n", strings.Split("a,b,c", ","))
            fmt.Printf("%q\n", strings.Split("a man a plan a canal panama", "a "))
            fmt.Printf("%q\n", strings.Split(" xyz ", ""))
            fmt.Printf("%q\n", strings.Split("", "Bernardo O'Higgins"))
        */

        _ = appQuest1()

        return nil
    }

    app.Run(os.Args)
}

func appQuit() error {

    cmdStr := []string{"rm -rf", backupPath}
    out := exeCmd(strings.Join(cmdStr, " "))
    fmt.Println(out)

    fmt.Println("You have request the Exit - Good Bye!")

    return cli.NewExitError("All Okay", 0)
}

func appQuest1() error {

    s1 := "swift list db_backup -p"
    stringPrefix := []string{os.Getenv("OS_REGION_NAME"), os.Getenv("MY_POD_NAMESPACE"), os.Getenv("MY_POD_NAME")}
    out := fmt.Sprintf("%s %s", s1, strings.Join(stringPrefix, "/"))

    list = exeCmd(out)

    list2 = deleteNoGzSuffix(deleteEmpty(strings.Split(list, "\n")))
    sort.Strings(list2)

    for id, str := range list2 {
        myStr := strings.Split(str, "/")
        // fmt.Printf("%q\n", myStr)

        if myStr[3] != "" && myStr[7] != "" {
            t, _ := time.Parse(longForm, myStr[3])
            fmt.Println(leftPad(strconv.Itoa(id), 3, "0"), ") ", myStr[0], "/", myStr[1], "/", myStr[2], " - ", myStr[7], " at ", t)
        }
    }

    reader := bufio.NewReader(os.Stdin)
    fmt.Print("ID of backup to restore or \"QUIT\" to Exit: ")
    text, _ := reader.ReadString('\n')
    text = strings.TrimRight(text, "\n")
    //fmt.Println(text)

    if listInt, err := strconv.Atoi(text); err == nil {
        if len(list2) > listInt && listInt > 0 {
            //fmt.Printf("%v is in map\n", listInt)
            // ToDo: add next step - download backup data
            _ = appQuest2(listInt)
        } else {
            fmt.Printf("%v is no backup ID\n", listInt)
            fmt.Println("Restart - Backup List")
            time.Sleep(2 * time.Second)
            _ = appQuest1()
        }
        return nil
    } else if strings.ToLower(text) == "quit" {
        // do nothing
        return appQuit()
    } else {
        _ = appQuest1()
    }
    return nil
}

// download backup
func appQuest2(index int) error {
    fmt.Printf("%v is in map\n", index)

    fmt.Println(list2[index])
    slicedStr := strings.Split(list2[index], "/")

    fmt.Println(slicedStr)

    stringPrefix := []string{os.Getenv("OS_REGION_NAME"), os.Getenv("MY_POD_NAMESPACE"), os.Getenv("MY_POD_NAME"), slicedStr[3], "backup", backupType, "base"}
    out := fmt.Sprintf("%s %s %s %s", "swift download db_backup -r -p", strings.Join(stringPrefix, "/"), "-D", backupPath)

    list = exeCmd(out)
    fmt.Println(list)

    // change workingdir to  /newbackup
    if err := os.Chdir(backupPath); err != nil {
        log.Fatal(err)
    }

    files, _ := ioutil.ReadDir(backupPath)
    for _, f := range files {
        if !f.IsDir() {
            if strings.HasSuffix(f.Name(), ".tar.gz") {
                fmt.Println("untargz file: " + f.Name())

                _ = exeCmd("tar -xvzf " + f.Name() + " --strip--components=3")

            } else if strings.HasSuffix(f.Name(), ".gz") {
                fmt.Println("gunzip file: " + f.Name())

                _ = exeCmd("gunzip " + f.Name())

            }

        }
    }

    return appQuit()
}

func appProcessRestore() error {

    files, _ := ioutil.ReadDir(backupPath)
    for _, f := range files {
        if !f.IsDir() {
            if strings.HasSuffix(f.Name(), ".gz") {
                filePath := []string{backupPath}
                filePath = append(filePath, f.Name())

                _ = deleteFile(strings.Join(filePath, "/"))
                fmt.Println("delete not needed file: " + f.Name())

            } else if strings.HasSuffix(f.Name(), ".sql") {

                table := strings.TrimSuffix(f.Name(), ".sql")

                if backupType == "mysql" {
                    return appMysqlDB(table)
                } else if backupType == "pgsql" {
                    return appPgsqlDB(table)
                }
            }

        } else {
            if f.Name() == "." || f.Name() == ".." {
                continue
            }
            if backupType == "influxdb" {
                _ = appInfluxDB(f.Name())
            }
        }
    }

    if backupType == "influxdb" {
        _ = exeCmd("chown -R influxdb:influxdb " + influxDBPath)
    }
    return nil
}

func appInfluxDB(table string) error {

    _ = exeCmd("influxd restore -metadir " + influxDBPath + "/meta " + backupPath + "/" + table)
    _ = exeCmd("influxd restore -database " + table + " -datadir " + influxDBPath + "/data " + backupPath + "/" + table)
    fmt.Println(">> database restore done: " + table)
    return nil
}

func appMysqlDB(table string) error {

    _ = exeCmdBashC("mysql -u root " + table + " < " + backupPath + "/" + table + ".sql")
    fmt.Println(">> database restore done: " + table)
    return nil
}

func appPgsqlDB(table string) error {

    _ = exeCmd("pg_restore -U postgres -h localhost --clean --create -d " + table + " " + backupPath + "/" + table + ".sql")
    fmt.Println(">> database restore done: " + table)
    return nil
}

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
        if str != "" {
            r = append(r, str)
        }
    }
    return r
}

func deleteNoGzSuffix(s []string) []string {
    var r []string
    for _, str := range s {
        if strings.HasSuffix(str, ".gz") {
            r = append(r, str)
        }
    }
    return r
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
