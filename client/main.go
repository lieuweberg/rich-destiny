package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"syscall"

	"time"

	"github.com/kardianos/service"
	richgo "github.com/lieuweberg/rich-go/client"
	_ "github.com/mattn/go-sqlite3"
)

var (
	// Injected by the go linker
	version string

	s                service.Service
	db               *sql.DB
	manifest         *sql.DB
	server           = &http.Server{Addr: "localhost:35893", Handler: nil}
	currentDirectory string
	exe              string

	storage *storageStruct
	// Generally don't use this, use http.DefaultClient. If you want to make a component request, use requestComponents.
	// All other requests to bungie should probably also use the DefaultClient.
	bungieHTTPClient *http.Client

	// Close this channel to stop the presence loop
	quitPresenceTicker  chan bool
	previousActivity    richgo.Activity
	forcePresenceUpdate bool
	debugText           string
)

type program struct{}

func (p *program) Start(s service.Service) (err error) {
	go p.run()
	return
}

func (p *program) Stop(s service.Service) (err error) {
	log.Print("OS termination received")
	db.Close()
	log.Print("Database closed")
	if manifest != nil {
		manifest.Close()
		log.Print("Manifest closed")
	} else {
		log.Print("Definitions didn't exist, didn't close")
	}
	if quitPresenceTicker != nil {
		quitPresenceTicker <- true
		log.Print("Presence loop stopped")
	} else {
		log.Print("Presence loop wasn't running, didn't stop")
	}

	server.Close()
	log.Print("Gracefully exited, bye bye")
	return
}

func createService() {
	svcConfig := &service.Config{
		Name:        "rich-destiny",
		Description: "discord rich presence tool for destiny 2 ( https://richdestiny.app )",
		Executable:  exe,
	}
	prg := &program{}

	var err error
	s, err = service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
}

func successfullyStartService() (success bool, err error) {
	for i := 0; i <= 10; i++ {
		if i == 0 || i == 5 {
			err = s.Start()
			if err != nil {
				return
			}
		}
		_, err = http.Get("http://localhost:35893")
		if err != nil {
			time.Sleep(3 * time.Second)
		} else {
			success = true
			break
		}
	}

	if !success {
		fmt.Println(" It seems rich-destiny didn't want to start at all..." +
			"Try seeing if there is any information in the logs folder where rich-destiny was installed or head to the support server for help ( https://discord.gg/UNU4UXp ).")
		return
	}

	return
}

func main() {
	var err error
	exe, err = os.Executable()
	if err != nil {
		log.Fatalf("Couldn't find current path: %s", err)
	}
	currentDirectory = filepath.Dir(exe)

	if service.Interactive() {
		installProgram()
	} else {
		createService()
		err = s.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (p *program) run() {
	debugText = "Starting up..."

	if _, err := os.Stat(makePath("logs")); os.IsNotExist(err) {
		err = os.Mkdir(makePath("logs"), os.ModePerm)
		if err != nil {
			// Logs are voided. Return as the application is probably lacking permissions.
			log.Printf("Couldn't create logs directory: %s", err)
			return
		}
	}

	y, m, d := time.Now().Date()
	h, min, sec := time.Now().Clock()
	logFile, err := os.Create(makePath(fmt.Sprintf("logs/%d-%d-%d %dh%dm%ds.log", y, m, d, h, min, sec)))
	if err != nil {
		log.Printf("Couldn't create log file: %s", err)
	} else {
		log.SetOutput(logFile)

		if runtime.GOOS == "windows" {
			stdErrorHandle := syscall.STD_ERROR_HANDLE
			r0, _, e1 := syscall.Syscall(syscall.MustLoadDLL("kernel32").MustFindProc("SetStdHandle").Addr(),
				2, uintptr(stdErrorHandle), logFile.Fd(), 0)
			if r0 == 0 {
				log.Printf("Couldn't set stderr handle: %d", e1)
			}
		}
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	db, err = sql.Open("sqlite3", makePath("storage.db"))
	if err != nil {
		log.Printf("Error opening storage.db: %s", err)
	}

	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS data(
		key		STRING	PRIMARY KEY NOT NULL,
		value	STRING	NOT NULL
	)`); err != nil {
		log.Printf("Error creating storage.db table: %s", err)
	}

	startWebServer()

	// Wait for a decent computer to have booted, no internet connection means trouble
	// TODO: Way better way of handling internet connection status; this is pretty terrible
	time.Sleep(10 * time.Second)

	debugText = ""

	// Kinda useless since browser tabs cannot be opened from a service, but leaving it in
	if _, err = getStorage(); err != nil {
		log.Printf("Error getting auth: %s", err)
	} else {
		if storage.AutoUpdate {
			go func() {
				_, err = attemptApplicationUpdate()
				if err != nil {
					log.Printf("Error trying to update: %s", err)
				}
			}()
		}
	}

	getDefinitions()
}

func makePath(e string) string {
	return filepath.Join(currentDirectory, e)
}
