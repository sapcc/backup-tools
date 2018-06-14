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
	"github.com/sapcc/containers/backup-tools/go-src/utils"

	"github.com/ncw/swift"
	"gopkg.in/urfave/cli.v1"
)

var (
	clientSwift *swift.Connection
	backupPath  = utils.NewBackupPath
)

func appQuit() error {

	fmt.Println("Clearing " + backupPath + " ...")

	_ = os.RemoveAll(backupPath)

	fmt.Println("Clearing " + backupPath + " done!")

	fmt.Println("You have request the Exit - Good Bye!")

	return cli.NewExitError("All Okay", 0)
}

func startCrossregionInit() error {

	group := configuration.EnvironmentStruct{
		ContainerPrefix:      strings.Join([]string{os.Getenv("BACKUP_REGION_NAME"), os.Getenv("MY_POD_NAMESPACE"), os.Getenv("MY_POD_NAME")}, "/"),
		OsAuthURL:            os.Getenv("OS_AUTH_URL"),
		OsAuthVersion:        os.Getenv("OS_AUTH_VERSION"),
		OsIdentityAPIVersion: os.Getenv("OS_IDENTITY_API_VERSION"),
		OsUsername:           os.Getenv("OS_USERNAME"),
		OsUserDomainName:     os.Getenv("OS_USER_DOMAIN_NAME"),
		OsProjectName:        os.Getenv("OS_PROJECT_NAME"),
		OsProjectDomainName:  os.Getenv("OS_PROJECT_DOMAIN_NAME"),
		OsRegionName:         os.Getenv("OS_REGION_NAME"),
		OsPassword:           os.Getenv("OS_PASSWORD"),
	}

	data, err := json.Marshal(group)
	if err != nil {
		fmt.Println("error:", err)
	}

	str := b64.StdEncoding.WithPadding(-1).EncodeToString(data)
	fmt.Println(str)

	//fmt.Println("OS_AUTH_URL")

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
			os.Setenv("CONTAINER_PREFIX", jsonReturn.ContainerPrefix)
		}
		if jsonReturn.OsAuthURL != "" {
			os.Setenv("OS_AUTH_URL", jsonReturn.OsAuthURL)
		}
		if jsonReturn.OsAuthVersion != "" {
			os.Setenv("OS_AUTH_VERSION", jsonReturn.OsAuthVersion)
		}
		if jsonReturn.OsIdentityAPIVersion != "" {
			os.Setenv("OS_IDENTITY_API_VERSION", jsonReturn.OsIdentityAPIVersion)
		}
		if jsonReturn.OsUsername != "" {
			os.Setenv("OS_USERNAME", jsonReturn.OsUsername)
		}
		if jsonReturn.OsUserDomainName != "" {
			os.Setenv("OS_USER_DOMAIN_NAME", jsonReturn.OsUserDomainName)
		}
		if jsonReturn.OsProjectName != "" {
			os.Setenv("OS_PROJECT_NAME", jsonReturn.OsProjectName)
		}
		if jsonReturn.OsProjectDomainName != "" {
			os.Setenv("OS_PROJECT_DOMAIN_NAME", jsonReturn.OsProjectDomainName)
		}
		if jsonReturn.OsRegionName != "" {
			os.Setenv("OS_REGION_NAME", jsonReturn.OsRegionName)
		}
		if jsonReturn.OsPassword != "" {
			os.Setenv("OS_PASSWORD", jsonReturn.OsPassword)
		}

	}
	configuration.ContainerPrefix = os.Getenv("CONTAINER_PREFIX")

	if configuration.ContainerPrefix == "" {
		configuration.ContainerPrefix = strings.Join([]string{os.Getenv("OS_REGION_NAME"), os.Getenv("MY_POD_NAMESPACE"), os.Getenv("MY_POD_NAME")}, "/")
		os.Setenv("CONTAINER_PREFIX", configuration.ContainerPrefix)
	}

	configuration.AuthVersion = os.Getenv("OS_AUTH_VERSION")
	configuration.AuthEndpoint = os.Getenv("OS_AUTH_URL")
	configuration.AuthUsername = os.Getenv("OS_USERNAME")
	configuration.AuthPassword = os.Getenv("OS_PASSWORD")
	configuration.AuthUserDomainName = os.Getenv("OS_USER_DOMAIN_NAME")
	configuration.AuthProjectName = os.Getenv("OS_PROJECT_NAME")
	configuration.AuthProjectDomainName = os.Getenv("OS_PROJECT_DOMAIN_NAME")
	configuration.AuthRegion = os.Getenv("OS_REGION_NAME")
	configuration.MysqlRootPassword = os.Getenv("MYSQL_ROOT_PASSWORD")

	var err error

	clientSwift, err = swiftcli.SwiftConnection(
		configuration.AuthVersion,
		configuration.AuthEndpoint,
		configuration.AuthUsername,
		configuration.AuthPassword,
		configuration.AuthUserDomainName,
		configuration.AuthProjectName,
		configuration.AuthProjectDomainName,
		configuration.AuthRegion)

	if err != nil {
		log.Println("Error can't connect swift for", configuration.AuthRegion, err)
		return err
	}

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
	var backupPath = backupPath
	// normalize index
	index = index - 1

	slicedStr := strings.Split(utils.List2[index], "/")

	fmt.Println("Download: " + utils.List2[index])

	_, err = swiftcli.SwiftDownloadPrefix(clientSwift, strings.Join([]string{configuration.ContainerPrefix, slicedStr[3], "backup", utils.BackupType, "base"}, "/"), &backupPath, false)
	if err != nil {
		log.Fatal(err)
	}

	// change workingdir to  /newbackup
	if err = os.Chdir(backupPath); err != nil {
		log.Fatal(err)
	}

	files, _ := ioutil.ReadDir(backupPath)
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
	err := swiftcli.UnpackFiles(objects)
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

				database := strings.TrimSuffix(f.Name(), ".sql")

				if utils.BackupType == "mysql" {
					appMysqlDB(database)
				} else if utils.BackupType == "pgsql" {
					appPgsqlDB(database)
				}
			}

		}
	}
	return appQuit()
}

func appMysqlDB(database string) error {

	//log.Println("mysql -u root -p'" + os.Getenv("MYSQL_ROOT_PASSWORD") + "' --socket /db/socket/mysqld.sock " + database + " < " + backupPath + "/" + database + ".sql")
	log.Println("mysql -u root -p'" + configuration.MysqlRootPassword + "' --socket /db/socket/mysqld.sock < " + backupPath + "/" + database + ".sql")

	//_ = exeCmdBashC("mysql -u root -p'" + os.Getenv("MYSQL_ROOT_PASSWORD") + "' --socket /db/socket/mysqld.sock " + database + " < " + backupPath + "/" + database + ".sql")
	_ = utils.ExeCmdBashC("mysql -u root -p'" + configuration.MysqlRootPassword + "' --socket /db/socket/mysqld.sock < " + backupPath + "/" + database + ".sql")

	fmt.Println(">> database restore done: " + database)
	return nil
}

func appPgsqlDB(database string) error {

	log.Println("psql -U postgres -h localhost -f " + backupPath + "/" + database + ".sql")

	_ = utils.ExeCmd("psql -U postgres -h localhost -f " + backupPath + "/" + database + ".sql")

	fmt.Println(">> database restore done: " + database)
	return nil
}
