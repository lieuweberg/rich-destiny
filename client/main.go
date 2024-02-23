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
	"strings"
	"syscall"

	"time"

	"github.com/kardianos/service"
	richgo "github.com/lieuweberg/rich-go/client"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mitchellh/go-ps"
)

var (
	// Injected by the go linker or when flagDev is true
	version string

	// Command line flags
	flagDaemon bool
	flagDev    bool

	// Other
	logFile          *os.File
	s                service.Service
	db               *sql.DB
	manifest         *sql.DB
	server           = &http.Server{Addr: "localhost:35893", Handler: nil}
	currentDirectory string
	exe              string
	exitChannel      chan os.Signal
	windowsUsers     []string

	veryImportantStatusActive bool

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

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	var err error

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
		logFile, err = os.Create(makePath(fmt.Sprintf("logs/%d-%d-%d %dh%dm%ds.log", y, m, d, h, min, sec)))
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

	var lastVersion string
	err = db.QueryRow("SELECT value FROM data WHERE key='lastVersion'").Scan(&lastVersion)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Error querying database for lastVersion: %s", err)
		}
	}

	err = storeData("lastVersion", version)

	if !service.Interactive() {
		err = tryServicelessTransition()
		if err != nil {
			log.Printf("Error trying serviceless transition: %s", err)
		}

		// This is not the first time launching this version, so if previously the transition worked we should now check for that and otherwise notify the user
		if lastVersion == version {
			log.Printf("Checking for different running rich-destiny.exe...")

			for i := 0; i < 6*5; i++ { // We check for 5 minutes
				pl, _ := ps.Processes()
				for _, p := range pl {
					if p.Executable() == "rich-destiny.exe" && p.Pid() != os.Getpid() {
						log.Printf("Running rich-destiny instance found, trying to uninstall service")
						err = s.Uninstall()
						if err != nil && !strings.Contains(err.Error(), "RemoveEventLogSource() failed") {
							log.Printf("Error uninstalling service: %s", err)
							return
						}
						err = s.Stop()
						if err != nil {
							log.Printf("Error stopping service: %s", err)
						}
						return
					}
				}

				time.Sleep(10 * time.Second)
			}

			log.Printf("Can't find running rich-destiny after 5 minutes. Assuming none exists. Starting but will only display very important status message.")
			setVeryImportantStatus(richgo.Activity{
				Details: "Please reinstall rich-destiny!",
				State:   "Your installation is broken.",
				Buttons: []*richgo.Button{
					{
						Label: "More Info (opens in browser)",
						Url:   "https://richdestiny.app/cp", // TODO: actual link
					},
				},
			})
		}
	}

	startWebServer()

	debugText = "Waiting for internet connection..."
	c := &http.Client{
		Timeout: 3 * time.Second,
	}
	var dnsError bool
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
				logErrorIfNoErrorSpam(errorOriginInternet, "Error trying to check internet/bungie connection: "+err.Error())
			}
			time.Sleep(10 * time.Second)
		} else {
			debugText = ""
			log.Println("Internet/Bungie connection seems ok!")
			resolveErrorSpam(errorOriginInternet)
			break
		}
	}

	// Kinda useless since browser tabs cannot be opened from a service, but leaving it in
	if _, err = getStorage(); err != nil {
		log.Printf("Error getting storage: %s", err)
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

	err = getDefinitions()
	if err != nil {
		log.Printf("Error getting definitions, will try again when the game is started: %s", err)
	}

	initPresence()
}

func stopApplication() {
	log.Print("OS termination received")
	if db != nil {
		db.Close()
		log.Print("Database closed")
	} else {
		log.Print("Storage was not opened, didn't close")
	}
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
	logFile.Close()
}

func makePath(e string) string {
	return filepath.Join(currentDirectory, e)
}

func setVeryImportantStatus(a richgo.Activity) {
	a.LargeImage = "important"
	setActivity(a, time.Now(), nil)
	veryImportantStatusActive = true
}
