package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"time"

	"github.com/kardianos/service"
	richgo "github.com/lieuweberg/rich-go/client"
	_ "github.com/mattn/go-sqlite3"
)

var (
	// Injected by the go linker or when flagDev is true
	version string

	// Command line flags
	flagDaemon bool
	flagDev    bool

	// Other
	s                service.Service
	db               *sql.DB
	manifest         *sql.DB
	server           = &http.Server{Addr: "localhost:35893", Handler: nil}
	currentDirectory string
	exe              string
	exitChannel      chan os.Signal

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

func init() {
	flag.BoolVar(&flagDaemon, "daemon", false, "run the program, not the install sequence")
	flag.BoolVar(&flagDev, "dev", false, "don't free the console and don't create a log file")
}

type program struct{}

func (p *program) Start(s service.Service) (err error) {
	go startApplication()
	return
}

func (p *program) Stop(s service.Service) (err error) {
	stopApplication()
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

func main() {
	// Here and not in init due to testing package erroring when this is run in init
	flag.Parse()

	var err error
	exe, err = os.Executable()
	if err != nil {
		log.Fatalf("Couldn't find current path: %s", err)
	}
	currentDirectory = filepath.Dir(exe)

	if service.Interactive() {
		if flagDaemon {
			if !flagDev {
				// https://docs.microsoft.com/en-us/windows/console/freeconsole
				r1, _, err := syscall.Syscall(syscall.MustLoadDLL("kernel32").MustFindProc("FreeConsole").Addr(), 0, 0, 0, 0)
				if r1 == 0 {
					log.Printf("Couldn't free console. This is probably important and should be sent in the support server ( https://discord.gg/UNU4UXp ): %s", err)
				}
			} else {
				version = "dev"
			}

			startApplication()

			exitChannel = make(chan os.Signal, 1)
			signal.Notify(exitChannel, syscall.SIGTERM, syscall.SIGINT, os.Interrupt, os.Kill)

			<-exitChannel
			stopApplication()
		} else {
			installProgram()
		}
	} else {
		createService()
		err = s.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func startApplication() {
	debugText = "Starting up..."

	if !flagDev {
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
				// https://docs.microsoft.com/en-us/windows/console/setstdhandle
				r1, _, err := syscall.Syscall(syscall.MustLoadDLL("kernel32").MustFindProc("SetStdHandle").Addr(), 2, uintptr(stdErrorHandle), logFile.Fd(), 0)
				if r1 == 0 {
					log.Printf("Couldn't set stderr handle: %d", err)
				}
			}
		}
	} else {
		version = "dev"
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	var err error
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

	debugText = "Waiting for internet connection..."
	c := &http.Client{
		Timeout: 3 * time.Second,
	}
	var dnsError bool
	var errCount int
	for {
		_, err = c.Get("https://www.bungie.net/Platform/GlobalAlerts/")
		if err != nil {
			var e *net.DNSError
			if errors.As(err, &e) {
				if !dnsError {
					log.Printf("DNS error trying to check internet/bungie connection. You can probably ignore this if you don't have internet "+
						"(e.g. no such host error). Muting further DNS errors: %s", err)
					dnsError = true
				}
			} else {
				if errCount <= 3 {
					log.Printf("Error trying to check internet/bungie connection: %s", err)
				}
				errCount++
			}
			time.Sleep(10 * time.Second)
		} else {
			debugText = ""
			log.Printf("Internet/Bungie connection seems ok! Errors: %d", errCount)
			break
		}
	}

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

func stopApplication() {
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
}

func makePath(e string) string {
	return filepath.Join(currentDirectory, e)
}
