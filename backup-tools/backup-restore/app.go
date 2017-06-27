package main

import (
    "bufio"
    b64 "encoding/base64"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "strconv"
    "strings"
    "time"

    "./internal"

    "github.com/ncw/swift"
    "gopkg.in/urfave/cli.v1"
)

var clientSwift swift.Connection

func appQuit() error {

    fmt.Println("Clearing " + backupPath + " ...")

    _ = os.RemoveAll(backupPath)

    fmt.Println("Clearing " + backupPath + " done!")

    fmt.Println("You have request the Exit - Good Bye!")

    return cli.NewExitError("All Okay", 0)
}

func startCrossregionInit() error {

    group := environmentStruct{
        ContainerPrefix:      strings.Join([]string{os.Getenv("BACKUP_REGION_NAME"), os.Getenv("MY_POD_NAMESPACE"), os.Getenv("MY_POD_NAME")}, "/"),
        OsAuthURL:            os.Getenv(strings.ToUpper(internal.Underscore("OsAuthURL"))),
        OsAuthVersion:        os.Getenv(strings.ToUpper(internal.Underscore("OsAuthVersion"))),
        OsIdentityAPIVersion: os.Getenv(strings.ToUpper(internal.Underscore("OsIdentityAPIVersion"))),
        OsUsername:           os.Getenv(strings.ToUpper(internal.Underscore("OsUsername"))),
        OsUserDomainName:     os.Getenv(strings.ToUpper(internal.Underscore("OsUserDomainName"))),
        OsProjectName:        os.Getenv(strings.ToUpper(internal.Underscore("OsProjectName"))),
        OsProjectDomainName:  os.Getenv(strings.ToUpper(internal.Underscore("OsProjectDomainName"))),
        OsRegionName:         os.Getenv(strings.ToUpper(internal.Underscore("OsRegionName"))),
        OsPassword:           os.Getenv(strings.ToUpper(internal.Underscore("OsPassword"))),
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

func startRestoreInit(cc bool) error {
    if !cc {
        fmt.Println("Welcome to the Backup-Restore process!")
        fmt.Println("Please follow the instructions to the end to restore your backup.")
        fmt.Println("With the \"QUIT\" command on user-input requests, you can stop the backup process.")
        fmt.Println("\n\nPress 'Enter' to continue...")
        bufio.NewReader(os.Stdin).ReadBytes('\n')
    }
    if os.Getenv("BACKUP_PGSQL_FULL") != "" {
        backupType = "pgsql"
    } else if os.Getenv("BACKUP_MYSQL_FULL") != "" {
        backupType = "mysql"
    }

    if backupType == "" {
        fmt.Println("\n\nNo System for the backup restore found.")
        fmt.Println("\n\n******** * * EXIT NO SUPPORTED SYSTEM FOUND * * ********")
        return cli.NewExitError("-- E: 1920291 --", 12)
    } else if cc == true {
    jumpToCJOFj30g2:

        reader := bufio.NewReader(os.Stdin)
        fmt.Print("Please enter the string from the backup container command \"backup-restore cc\" as a single-line: ")
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

        // fmt.Printf("%q\n", data)
        var jsonReturn environmentStruct
        err = json.Unmarshal(data, &jsonReturn)
        if err != nil {
            return cli.NewExitError("-- Fatal Config Decoding E: 39343 --", 1)
        }
        // fmt.Printf("%+v", jsonReturn)

        if jsonReturn.ContainerPrefix != "" {
            os.Setenv(strings.ToUpper(internal.Underscore("ContainerPrefix")), jsonReturn.ContainerPrefix)
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

    }
    if containerPrefix := os.Getenv(strings.ToUpper(internal.Underscore("ContainerPrefix"))); containerPrefix == "" {
        containerPrefix = strings.Join([]string{os.Getenv("OS_REGION_NAME"), os.Getenv("MY_POD_NAMESPACE"), os.Getenv("MY_POD_NAME")}, "/")
    }
    authVersion = os.Getenv(strings.ToUpper(internal.Underscore("OsAuthVersion")))
    authEndpoint = os.Getenv(strings.ToUpper(internal.Underscore("OsAuthURL")))
    authUsername = os.Getenv(strings.ToUpper(internal.Underscore("OsUsername")))
    authPassword = os.Getenv(strings.ToUpper(internal.Underscore("OsPassword")))
    authUserDomainName = os.Getenv(strings.ToUpper(internal.Underscore("OsUserDomainName")))
    authProjectName = os.Getenv(strings.ToUpper(internal.Underscore("OsProjectName")))
    authProjectDomainName = os.Getenv(strings.ToUpper(internal.Underscore("OsProjectDomainName")))
    authRegion = os.Getenv(strings.ToUpper(internal.Underscore("OsRegionName")))
    mysqlRootPassword = os.Getenv(strings.ToUpper(internal.Underscore("MysqlRootPassword")))

    clientSwift = SwiftConnection(
        authVersion, authEndpoint, authUsername, authPassword, authUserDomainName, authProjectName, authProjectDomainName, authRegion, containerPrefix,
    )

    os.Mkdir(backupPath, 0777)
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

    _ = appQuest1(false)

    return nil
    return nil
}

func appQuest1(full bool) error {

    list, err := SwiftListPrefixFiles(clientSwift, containerPrefix)

    if err != nil {
        return cli.NewExitError("-- E: 200.050 --", 12)
    }

    list2 = makePrefixPathOnly(deleteNoGzSuffix(deleteEmpty(list)))

    // Last 5 Backup List of backups
    if !full {
        length := len(list2)
        start := length - 5

        if start < 0 {
            start = 0
        }

        if start > length {
            start = length
        }

        list2 = list2[start:]
    }

    for id, str := range list2 {
        myStr := strings.Split(str, "/")
        // fmt.Printf("%q\n", myStr)

        if myStr[3] != "" {
            t, _ := time.Parse(longForm, myStr[3])
            fmt.Println(leftPad(strconv.Itoa(id+1), 3, "0"), ") ", myStr[0], "/", myStr[1], "/", myStr[2], " at ", t)
        }
    }

    reader := bufio.NewReader(os.Stdin)
    fmt.Println("Cross-Region Backup (need Config-String) with \"crossregion\"")
    fmt.Println("Full backup-list with \"full-list\"")
    fmt.Println("Manual backup restore from /newbackup/ with \"manual\"")
    fmt.Print("Enter ID of backup to restore or \"QUIT\" to Exit: ")
    text, _ := reader.ReadString('\n')
    text = strings.TrimRight(text, "\n")
    //fmt.Println(text)

    if listInt, err := strconv.Atoi(text); err == nil {
        if len(list) > listInt && listInt > 0 {
            fmt.Println("The next step can take a while... please wait...")
            // ToDo: add next step - download backup data
            _ = appQuest2(listInt)
        } else {
            fmt.Printf("%v is no backup ID\n", listInt)
            fmt.Println("Restart - Backup List")
            time.Sleep(2 * time.Second)
            _ = appQuest1(false)
        }
        return nil
    } else if strings.ToLower(text) == "full-list" {
        // do nothing
        return appQuest1(true)
    } else if strings.ToLower(text) == "quit" {
        // do nothing
        return appQuit()
    } else if strings.ToLower(text) == "manual" {
        // do nothing
        return appQuestManual()
    } else if strings.ToLower(text) == "crossregion" {
        // do nothing
        return startRestoreInit(true)
    }

    return appQuest1(false)
}

// download backup
func appQuest2(index int) error {
    // normalize index
    index = index - 1

    slicedStr := strings.Split(list2[index], "/")

    fmt.Println("Download: " + list2[index])

    _, err := SwiftDownloadPrefix(clientSwift, strings.Join([]string{containerPrefix, slicedStr[3], "backup", backupType, "base"}, "/"))
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

    return appProcessRestore()
}

// download backup
func appQuestManual() error {
    fmt.Println("Backup Manual from " + backupPath)

    // change workingdir to  /newbackup
    if err := os.Chdir(backupPath); err != nil {
        log.Fatal(err)
    }

    files, _ := ioutil.ReadDir(backupPath)
    objects := make([]string, 0)
    for _, file := range files {
        objects = append(objects, file.Name())
    }
    err := UnpackFiles(objects)
    if err != nil {
        log.Fatal(err)
    }

    return appProcessRestore()
}

func appProcessRestore() error {

    files, _ := ioutil.ReadDir(backupPath)
    for _, f := range files {
        if f.Name() == "." || f.Name() == ".." {
            continue
        } else if strings.HasPrefix(f.Name(), "mysql.") {
            continue
        } else if !f.IsDir() {
            if strings.HasSuffix(f.Name(), ".sql") {

                table := strings.TrimSuffix(f.Name(), ".sql")

                if backupType == "mysql" {
                    appMysqlDB(table)
                } else if backupType == "pgsql" {
                    appPgsqlDB(table)
                }
            }

        }
    }
    return appQuit()
}

func appMysqlDB(table string) error {

    //log.Println("mysql -u root -p'" + os.Getenv("MYSQL_ROOT_PASSWORD") + "' --socket /db/socket/mysqld.sock " + table + " < " + backupPath + "/" + table + ".sql")
    log.Println("mysql -u root -p'" + mysqlRootPassword + "' --socket /db/socket/mysqld.sock < " + backupPath + "/" + table + ".sql")

    //_ = exeCmdBashC("mysql -u root -p'" + os.Getenv("MYSQL_ROOT_PASSWORD") + "' --socket /db/socket/mysqld.sock " + table + " < " + backupPath + "/" + table + ".sql")
    _ = exeCmdBashC("mysql -u root -p'" + mysqlRootPassword + "' --socket /db/socket/mysqld.sock < " + backupPath + "/" + table + ".sql")

    fmt.Println(">> database restore done: " + table)
    return nil
}

func appPgsqlDB(table string) error {

    log.Println("psql -U postgres -h localhost -d " + table + " -f " + backupPath + "/" + table + ".sql")

    _ = exeCmd("psql -U postgres -h localhost -d " + table + " -f " + backupPath + "/" + table + ".sql")

    fmt.Println(">> database restore done: " + table)
    return nil
}
