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

	"github.com/sapcc/containers/backup-tools/go-src/configuration"
	"github.com/sapcc/containers/backup-tools/go-src/swiftcli"
	"github.com/sapcc/containers/backup-tools/go-src/underscore"
	"github.com/sapcc/containers/backup-tools/go-src/utils"

	"github.com/ncw/swift"
	"gopkg.in/urfave/cli.v1"
)

var (
	clientSwift *swift.Connection
)

func appQuit() error {

	fmt.Println("Clearing " + utils.BackupPath + " ...")

	_ = os.RemoveAll(utils.BackupPath)

	fmt.Println("Clearing " + utils.BackupPath + " done!")

	fmt.Println("You have request the Exit - Good Bye!")

	return cli.NewExitError("All Okay", 0)
}

func startCrossregionInit() error {

	group := configuration.EnvironmentStruct{
		ContainerPrefix:      strings.Join([]string{os.Getenv("BACKUP_REGION_NAME"), os.Getenv("MY_POD_NAMESPACE"), os.Getenv("MY_POD_NAME")}, "/"),
		OsAuthURL:            os.Getenv(strings.ToUpper(underscore.Underscore("OsAuthURL"))),
		OsAuthVersion:        os.Getenv(strings.ToUpper(underscore.Underscore("OsAuthVersion"))),
		OsIdentityAPIVersion: os.Getenv(strings.ToUpper(underscore.Underscore("OsIdentityAPIVersion"))),
		OsUsername:           os.Getenv(strings.ToUpper(underscore.Underscore("OsUsername"))),
		OsUserDomainName:     os.Getenv(strings.ToUpper(underscore.Underscore("OsUserDomainName"))),
		OsProjectName:        os.Getenv(strings.ToUpper(underscore.Underscore("OsProjectName"))),
		OsProjectDomainName:  os.Getenv(strings.ToUpper(underscore.Underscore("OsProjectDomainName"))),
		OsRegionName:         os.Getenv(strings.ToUpper(underscore.Underscore("OsRegionName"))),
		OsPassword:           os.Getenv(strings.ToUpper(underscore.Underscore("OsPassword"))),
	}

	data, err := json.Marshal(group)
	if err != nil {
		fmt.Println("error:", err)
	}

	str := b64.StdEncoding.WithPadding(-1).EncodeToString(data)
	fmt.Println(str)

	//fmt.Println(strings.ToUpper(underscore.Underscore("OsAuthURL")))

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
		utils.BackupType = "pgsql"
	} else if os.Getenv("BACKUP_MYSQL_FULL") != "" {
		utils.BackupType = "mysql"
	}

	if utils.BackupType == "" {
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
		var jsonReturn configuration.EnvironmentStruct
		err = json.Unmarshal(data, &jsonReturn)
		if err != nil {
			return cli.NewExitError("-- Fatal Config Decoding E: 39343 --", 1)
		}
		// fmt.Printf("%+v", jsonReturn)

		if jsonReturn.ContainerPrefix != "" {
			os.Setenv(strings.ToUpper(underscore.Underscore("ContainerPrefix")), jsonReturn.ContainerPrefix)
		}
		if jsonReturn.OsAuthURL != "" {
			os.Setenv(strings.ToUpper(underscore.Underscore("OsAuthURL")), jsonReturn.OsAuthURL)
		}
		if jsonReturn.OsAuthVersion != "" {
			os.Setenv(strings.ToUpper(underscore.Underscore("OsAuthVersion")), jsonReturn.OsAuthVersion)
		}
		if jsonReturn.OsIdentityAPIVersion != "" {
			os.Setenv(strings.ToUpper(underscore.Underscore("OsIdentityAPIVersion")), jsonReturn.OsIdentityAPIVersion)
		}
		if jsonReturn.OsUsername != "" {
			os.Setenv(strings.ToUpper(underscore.Underscore("OsUsername")), jsonReturn.OsUsername)
		}
		if jsonReturn.OsUserDomainName != "" {
			os.Setenv(strings.ToUpper(underscore.Underscore("OsUserDomainName")), jsonReturn.OsUserDomainName)
		}
		if jsonReturn.OsProjectName != "" {
			os.Setenv(strings.ToUpper(underscore.Underscore("OsProjectName")), jsonReturn.OsProjectName)
		}
		if jsonReturn.OsProjectDomainName != "" {
			os.Setenv(strings.ToUpper(underscore.Underscore("OsProjectDomainName")), jsonReturn.OsProjectDomainName)
		}
		if jsonReturn.OsRegionName != "" {
			os.Setenv(strings.ToUpper(underscore.Underscore("OsRegionName")), jsonReturn.OsRegionName)
		}
		if jsonReturn.OsPassword != "" {
			os.Setenv(strings.ToUpper(underscore.Underscore("OsPassword")), jsonReturn.OsPassword)
		}

	}
	configuration.ContainerPrefix = os.Getenv(strings.ToUpper(underscore.Underscore("ContainerPrefix")))

	if configuration.ContainerPrefix == "" {
		configuration.ContainerPrefix = strings.Join([]string{os.Getenv(strings.ToUpper(underscore.Underscore("OsRegionName"))), os.Getenv("MY_POD_NAMESPACE"), os.Getenv("MY_POD_NAME")}, "/")
		os.Setenv(strings.ToUpper(underscore.Underscore("ContainerPrefix")), configuration.ContainerPrefix)
	}

	configuration.AuthVersion = os.Getenv(strings.ToUpper(underscore.Underscore("OsAuthVersion")))
	configuration.AuthEndpoint = os.Getenv(strings.ToUpper(underscore.Underscore("OsAuthURL")))
	configuration.AuthUsername = os.Getenv(strings.ToUpper(underscore.Underscore("OsUsername")))
	configuration.AuthPassword = os.Getenv(strings.ToUpper(underscore.Underscore("OsPassword")))
	configuration.AuthUserDomainName = os.Getenv(strings.ToUpper(underscore.Underscore("OsUserDomainName")))
	configuration.AuthProjectName = os.Getenv(strings.ToUpper(underscore.Underscore("OsProjectName")))
	configuration.AuthProjectDomainName = os.Getenv(strings.ToUpper(underscore.Underscore("OsProjectDomainName")))
	configuration.AuthRegion = os.Getenv(strings.ToUpper(underscore.Underscore("OsRegionName")))
	configuration.MysqlRootPassword = os.Getenv(strings.ToUpper(underscore.Underscore("MysqlRootPassword")))

	clientSwift = swiftcli.SwiftConnection(
		configuration.AuthVersion,
		configuration.AuthEndpoint,
		configuration.AuthUsername,
		configuration.AuthPassword,
		configuration.AuthUserDomainName,
		configuration.AuthProjectName,
		configuration.AuthProjectDomainName,
		configuration.AuthRegion,
		configuration.ContainerPrefix,
	)

	os.Mkdir(utils.BackupPath, 0777)
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
}

func appQuest1(full bool) error {
	var err error
	utils.List, err = swiftcli.SwiftListPrefixFiles(clientSwift, configuration.ContainerPrefix)

	if err != nil {
		return cli.NewExitError("-- E: 200.050 --", 12)
	}

	utils.List2 = utils.MakePrefixPathOnly(utils.DeleteNoGzSuffix(utils.DeleteEmpty(utils.List)))

	// Last 5 Backup List of backups
	if !full {
		length := len(utils.List2)
		start := length - 5

		if start < 0 {
			start = 0
		}

		if start > length {
			start = length
		}

		utils.List2 = utils.List2[start:]
	}

	for id, str := range utils.List2 {
		myStr := strings.Split(str, "/")
		// fmt.Printf("%q\n", myStr)

		if myStr[3] != "" {
			t, _ := time.Parse(configuration.LongDateForm, myStr[3])
			fmt.Println(utils.LeftPad(strconv.Itoa(id+1), 3, "0"), ") ", myStr[0], "/", myStr[1], "/", myStr[2], " at ", t)
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
		if len(utils.List) >= listInt && listInt > 0 {
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
	var err error
	// normalize index
	index = index - 1

	slicedStr := strings.Split(utils.List2[index], "/")

	fmt.Println("Download: " + utils.List2[index])

	_, err = swiftcli.SwiftDownloadPrefix(clientSwift, strings.Join([]string{configuration.ContainerPrefix, slicedStr[3], "backup", utils.BackupType, "base"}, "/"))
	if err != nil {
		log.Fatal(err)
	}

	// change workingdir to  /newbackup
	if err = os.Chdir(utils.BackupPath); err != nil {
		log.Fatal(err)
	}

	files, _ := ioutil.ReadDir(utils.BackupPath)
	objects := make([]string, 0)
	for _, file := range files {
		objects = append(objects, file.Name())
	}
	err = swiftcli.UnpackFiles(objects)
	if err != nil {
		log.Fatal(err)
	}

	return appProcessRestore()
}

// download backup
func appQuestManual() error {
	fmt.Println("Backup Manual from " + utils.BackupPath)

	// change workingdir to  /newbackup
	if err := os.Chdir(utils.BackupPath); err != nil {
		log.Fatal(err)
	}

	files, _ := ioutil.ReadDir(utils.BackupPath)
	objects := make([]string, 0)
	for _, file := range files {
		objects = append(objects, file.Name())
	}
	err := swiftcli.UnpackFiles(objects)
	if err != nil {
		log.Fatal(err)
	}

	return appProcessRestore()
}

func appProcessRestore() error {

	files, _ := ioutil.ReadDir(utils.BackupPath)
	for _, f := range files {
		if f.Name() == "." || f.Name() == ".." {
			continue
		} else if strings.HasPrefix(f.Name(), "mysql.") {
			continue
		} else if !f.IsDir() {
			if strings.HasSuffix(f.Name(), ".sql") {

				table := strings.TrimSuffix(f.Name(), ".sql")

				if utils.BackupType == "mysql" {
					appMysqlDB(table)
				} else if utils.BackupType == "pgsql" {
					appPgsqlDB(table)
				}
			}

		}
	}
	return appQuit()
}

func appMysqlDB(table string) error {

	//log.Println("mysql -u root -p'" + os.Getenv("MYSQL_ROOT_PASSWORD") + "' --socket /db/socket/mysqld.sock " + table + " < " + utils.BackupPath + "/" + table + ".sql")
	log.Println("mysql -u root -p'" + configuration.MysqlRootPassword + "' --socket /db/socket/mysqld.sock < " + utils.BackupPath + "/" + table + ".sql")

	//_ = exeCmdBashC("mysql -u root -p'" + os.Getenv("MYSQL_ROOT_PASSWORD") + "' --socket /db/socket/mysqld.sock " + table + " < " + utils.BackupPath + "/" + table + ".sql")
	_ = utils.ExeCmdBashC("mysql -u root -p'" + configuration.MysqlRootPassword + "' --socket /db/socket/mysqld.sock < " + utils.BackupPath + "/" + table + ".sql")

	fmt.Println(">> database restore done: " + table)
	return nil
}

func appPgsqlDB(table string) error {

	log.Println("psql -U postgres -h localhost -d " + table + " -f " + utils.BackupPath + "/" + table + ".sql")

	_ = utils.ExeCmd("psql -U postgres -h localhost -d " + table + " -f " + utils.BackupPath + "/" + table + ".sql")

	fmt.Println(">> database restore done: " + table)
	return nil
}
