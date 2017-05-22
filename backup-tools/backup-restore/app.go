package main

import (
    "bufio"
    b64 "encoding/base64"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "sort"
    "strconv"
    "strings"
    "time"

    "./internal"

    "gopkg.in/urfave/cli.v1"
)

var (
    containerPrefix       = "staging/limes/limes-postgresql"
    authVersion           = "3"
    authEndpoint          = "https://identity-3.staging.cloud.sap/v3"
    authUsername          = "db_backup"
    authPassword          = "Dw9QKthZRCUMQUf"
    authUserDomainName    = "Default"
    authProjectName       = "master"
    authProjectDomainName = "ccadmin"
    authRegion            = "staging"
)

func appQuit() error {

    fmt.Println("Clearing " + backupPath + " ...")

    _ = os.RemoveAll(backupPath)

    fmt.Println("Clearing " + backupPath + " done!")

    fmt.Println("You have request the Exit - Good Bye!")

    return cli.NewExitError("All Okay", 0)
}

func startInfluxInit() error {
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
}

func startRestoreInit() error {
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

    } else {

        containerPrefix = strings.Join([]string{os.Getenv("OS_REGION_NAME"), os.Getenv("MY_POD_NAMESPACE"), os.Getenv("MY_POD_NAME")}, "/")
        authVersion = os.Getenv(strings.ToUpper(internal.Underscore("OsAuthVersion")))
        authEndpoint = os.Getenv(strings.ToUpper(internal.Underscore("OsAuthURL")))
        authUsername = os.Getenv(strings.ToUpper(internal.Underscore("OsUsername")))
        authPassword = os.Getenv(strings.ToUpper(internal.Underscore("OsPassword")))
        authUserDomainName = os.Getenv(strings.ToUpper(internal.Underscore("OsUserDomainName")))
        authProjectName = os.Getenv(strings.ToUpper(internal.Underscore("OsProjectName")))
        authProjectDomainName = os.Getenv(strings.ToUpper(internal.Underscore("OsProjectDomainName")))
        authRegion = os.Getenv(strings.ToUpper(internal.Underscore("OsRegionName")))

        SwiftConnection(
            authVersion, authEndpoint, authUsername, authPassword, authUserDomainName, authProjectName, authProjectDomainName, authRegion, containerPrefix,
        )
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
    return nil
}

func appQuest1() error {

    list, err := SwiftListPrefixFiles(containerPrefix)

    if err != nil {
        return cli.NewExitError("-- E: 200.050 --", 12)
    }

    list2 = deleteNoGzSuffix(deleteEmpty(list))
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
        if len(list) > listInt && listInt > 0 {
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

    err := SwiftDownloadPrefix(strings.Join([]string{os.Getenv("OS_REGION_NAME"), os.Getenv("MY_POD_NAMESPACE"), os.Getenv("MY_POD_NAME"), slicedStr[3], "backup", backupType, "base"}, "/"))
    if err != nil {
        log.Fatal(err)
    }

    // change workingdir to  /newbackup
    if err := os.Chdir(backupPath); err != nil {
        log.Fatal(err)
    }

    files, _ := ioutil.ReadDir(backupPath)
    objects := make([]string, 0)
    for _, file := range files {
        objects = append(objects, file.Name())
    }
    err = UnpackFiles(objects)
    if err != nil {
        log.Fatal(err)
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
