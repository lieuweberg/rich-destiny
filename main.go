//go:generate goversioninfo -64
package main

import (
	"archive/zip"
	"database/sql"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"

	"time"

	"github.com/kardianos/service"
	_ "github.com/mattn/go-sqlite3"
)

var generatedState string
var db *sql.DB
var manifest *sql.DB
var server = &http.Server{Addr: ":35893", Handler: nil}
var currentDirectory string
var programStatus programStatusStruct

type program struct{}

func (p *program) Start(s service.Service) (err error) {
	go p.run()
	return
}

func (p *program) Stop(s service.Service) (err error) {
	log.Print("OS termination received")
	db.Close()
	manifest.Close()
	server.Close()
	close(quitExeCheckTicker)
	log.Print("Gracefully exited, bye bye")
	return
}

func main() {
	if _, err := os.Stat(makePath("logs")); os.IsNotExist(err) {
		err = os.Mkdir(makePath("logs"), os.ModePerm)
		if err != nil {
			log.Printf("Couldn't create logs directory: %s", err)
		}
	}

	svcConfig := &service.Config{
		Name:        "rich-destiny",
		Description: "Discord rich presence tool for Destiny 2",
	}
	prg := &program{}

	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	if service.Interactive() {
		log.Print("Detected that this program is being run manually, installing into the service manager...")
		err = s.Install()
		if err != nil {
			log.Printf("Something went wrong: %s", err)
			return
		}
		log.Print("Program installed. The service should automatically start (check the installation directory for new files).")
		err = s.Start()
		if err != nil {
			log.Printf("Error starting service, aborting further execution: %s", err)
			return
		}
		log.Print("Assuming this is first-time installation; this program will now attempt to open a browser tab to log in with Bungie.net.")
		openOauthTab()
		
		log.Print("Press ENTER to close this window.")
		fmt.Scanln()
	} else {
		err = s.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (p *program) run() {
	setPrgStatus(true, "Initialising...")
	// Wait for a decent computer to have booted, no internet connection means trouble
	// TODO: Way better way of handling internet connection status, this is pretty terrible
	time.Sleep(20 * time.Second)

	exe, err := os.Executable()
	if err != nil {
		log.Fatalf("Couldn't find current path: %s", err)
	}
	currentDirectory = filepath.Dir(exe)

	logFile, err := os.Create(makePath(fmt.Sprintf("logs/%d.log", time.Now().Unix())))
	if err != nil {
		log.Printf("Couldn't create log file: %s", err)
	} else {
		log.SetOutput(logFile)
	}

	// State query param
	rand.Seed(time.Now().UnixNano())
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 20)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	generatedState = string(b)

	// Open the storage db (access token, etc.)
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

	if _, err = getAuth(); err != nil {
		log.Printf("Error getting auth: %s", err)
		setPrgStatus(false, "Unable to get auth.")
	}

	// Check if a new manifest has to be downloaded, if so do that, then open the db
	manifestRes, err := getManifestData()
	if err != nil {
		log.Printf("Error getting manifest data: %s", err)
		setPrgStatus(false, "Unable to get new manifest.")
	}

	startWebServer()
	initPresence()

	var lastManifestURL string
	db.QueryRow("SELECT value FROM data WHERE key='lastManifestURL'").Scan(&lastManifestURL)
	if _, err := os.Stat(makePath("manifest.db"));  (os.IsNotExist(err) || manifestRes.Response.MobileWorldContentPaths.En != lastManifestURL) {
		if os.IsNotExist(err) {
			log.Print("Manifest doesn't exist, downloading one...")
		} else {
			log.Print("Manifest is outdated, downloading a new one...")
		}

		res, err := http.Get("https://bungie.net" + manifestRes.Response.MobileWorldContentPaths.En)
		if err != nil {
			log.Printf("Error getting manifest database: %s", err)
			return
		}
		out, err := os.Create(makePath("manifest.zip"))
		if err != nil {
			log.Printf("Error creating manifest.zip: %s", err)
			return
		}
		_, err = io.Copy(out, res.Body)
		res.Body.Close()
		log.Print("Manifest downloaded, unzipping...")

		z, err := zip.OpenReader(out.Name())
		out.Close()
		if err != nil {
			log.Printf("Error writing/unzipping manifest.zip: %s", err)
			return
		}
		var success bool
		for _, f := range z.File {
			file, err := f.Open()
			if err != nil {
				log.Printf("Error opening file: %s", err)
				break
			}
			out, err := os.Create(makePath("manifest.db"))
			if err != nil {
				log.Printf("Error creating manifest.db: %s", err)
				break
			}
			_, err = io.Copy(out, file)
			if err != nil {
				log.Printf("Error writing manifest.db: %s", err)
				return
			}
			file.Close()
			out.Close()
			success = true
		}
		z.Close()
		if !success {
			log.Print("Something went wrong downloading and unzipping the manifest. There may be errors above.")
			return
		}
		log.Print("Manifest downloaded and unzipped!")

		err = os.Remove(makePath("manifest.zip"))
		if err != nil {
			log.Printf("Error deleting manifest.zip: %s", err)
			return
		}
		log.Print("Deleted temporary file manifest.zip")

		_, err = db.Exec("INSERT OR REPLACE INTO data(key, value) VALUES('lastManifestURL', $1)", manifestRes.Response.MobileWorldContentPaths.En)
		if err != nil {
			log.Printf("Error setting lastManifestURL to storage.db: %s", err)
			return
		}
	}

	manifest, err = sql.Open("sqlite3", makePath("manifest.db"))
	if err != nil {
		log.Printf("Error opening manifest.db: %s", err)
		return
	}
}

func startWebServer() {
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		enableCors(&res)
		fmt.Fprint(res, "{message: \"Hello\"}")
	})

	http.HandleFunc("/login", func(res http.ResponseWriter, req *http.Request) {
		http.Redirect(res, req, fmt.Sprintf("https://www.bungie.net/en/oauth/authorize?response_type=code&client_id=%s&redirect_uri=%s&state=%s", config.ClientID, config.RedirectURI, generatedState), http.StatusFound)
	})

	http.HandleFunc("/callback", func(res http.ResponseWriter, req *http.Request) {
		code := req.URL.Query().Get("code")
		state := req.URL.Query().Get("state")
		if code == "" || state != generatedState {
			res.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(res, "{error: \"400: Bad Request\"}")
			return
		}

		err := requestAccessToken(code, false)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(res, "{error: \"500: %s\"}", err)
		}

		fmt.Fprint(res, "{message: \"Success! You may now close this tab.\"}")
		browserOpened = false
	})

	go func() {
		log.Print("If no further errors, listening on port http://localhost:35893")
		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				log.Print("http server closed")
				return
			}
			log.Printf("Error with http server: %s", err)
		}
	}()
}

func setPrgStatus(ok bool, message string) {
	programStatus.OK = ok
	programStatus.Message = message
}

func makePath(e string) string {
	return filepath.Join(currentDirectory, e)
}

func enableCors(res *http.ResponseWriter) {
	(*res).Header().Set("Access-Control-Allow-Origin", "http://localhost:5500")
}
