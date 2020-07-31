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

	richgo "github.com/hugolgst/rich-go/client"
	"github.com/kardianos/service"
	_ "github.com/mattn/go-sqlite3"
)

// Injected by the go linker
var version string

var generatedState string
var db *sql.DB
var manifest *sql.DB
var server = &http.Server{Addr: ":35893", Handler: nil}
var currentDirectory string

var auth *authResponse
var browserOpened bool
// Generally don't use this, use http.DefaultClient. If you want to make a component request, use requestComponents.
// All other requests to bungie should probably also use the DefaultClient.
var bungieHTTPClient *http.Client

// Close this channel to stop the presence loop
var quitPresenceTicker chan(struct{})
var previousActivity richgo.Activity
var debugHashes string

type program struct{}

func (p *program) Start(s service.Service) (err error) {
	fmt.Print("hi2")
	go p.run()
	return
}

func (p *program) Stop(s service.Service) (err error) {
	log.Print("OS termination received")
	db.Close()
	manifest.Close()
	close(quitPresenceTicker)
	server.Close()
	log.Print("Gracefully exited, bye bye")
	return
}

func main() {
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
		fmt.Print("         _      _              _           _   _\n        (_)    | |            | |         | | (_)\n    _ __ _  ___| |__ ______ __| | ___  ___| |_ _ _ __  _   _\n   | '__| |/ __| '_ \\______/ _` |/ _ \\/ __| __| | '_ \\| | | |\n   | |  | | (__| | | |    | (_| |  __/\\__ \\ |_| | | | | |_| |\n   |_|  |_|\\___|_| |_|     \\__,_|\\___||___/\\__|_|_| |_|\\__, |\n                                                        __/ |\n                                                       |___/    ",
			version, "\n\n")
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

		log.Print("Assuming this is first-time installation; this program will shortly attempt to open a browser tab to log in with Bungie.net.")
		time.Sleep(5 * time.Second)
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
	exe, err := os.Executable()
	if err != nil {
		log.Fatalf("Couldn't find current path: %s", err)
	}
	currentDirectory = filepath.Dir(exe)

	if _, err := os.Stat(makePath("logs")); os.IsNotExist(err) {
		err = os.Mkdir(makePath("logs"), os.ModePerm)
		if err != nil {
			// Logs are voided from here on out. Return as the application is probably lacking permissions.
			log.Printf("Couldn't create logs directory: %s", err)
			return
		}
	}

	y, m, d := time.Now().Date()
	h, min, s := time.Now().Clock()
	logFile, err := os.Create(makePath(fmt.Sprintf("logs/%d-%d-%d %dh%dm%ds.log", y, m, d, h, min, s)))
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

	go startWebServer()
	// The following section returns on most errors, so defer this function (long manifest downloads can cause issues for initPresence, too)
	defer func() {
		initPresence()
	}()

	// Wait for a decent computer to have booted, no internet connection means trouble
	// TODO: Way better way of handling internet connection status; this is pretty terrible
	time.Sleep(10 * time.Second)

	// Kinda useless since browser tabs cannot be opened from a service, but leaving it in
	if _, err = getAuth(); err != nil {
		log.Printf("Error getting auth: %s", err)
	}

	// Check if a new manifest has to be downloaded, if so do that, then open the db
	manifestRes, err := getManifestData()
	if err != nil {
		log.Printf("Error getting manifest data: %s", err)
	}

	var lastManifestURL string
	db.QueryRow("SELECT value FROM data WHERE key='lastManifestURL'").Scan(&lastManifestURL)
	if _, err := os.Stat(makePath("manifest.db")); os.IsNotExist(err) || manifestRes.Response.MobileWorldContentPaths.En != lastManifestURL {
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
		enableCors(&res, req)
		res.Header().Set("Content-Type", "application/json")
		fmt.Fprint(res, "{\"message\": \"hello\"}")
	})

	http.HandleFunc("/login", func(res http.ResponseWriter, req *http.Request) {
		http.Redirect(res, req, fmt.Sprintf("https://www.bungie.net/en/oauth/authorize?response_type=code&client_id=%s&redirect_uri=%s&state=%s",
			config.ClientID, config.RedirectURI, generatedState), http.StatusFound)
	})

	http.HandleFunc("/callback", func(res http.ResponseWriter, req *http.Request) {
		code := req.URL.Query().Get("code")
		state := req.URL.Query().Get("state")
		if code == "" || state != generatedState {
			res.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(res, "error: 400: Bad Request")
			return
		}

		err := requestAccessToken(code, false)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(res, "error: 500: %s", err)
		}

		fmt.Fprint(res, "Success! You are now logged in and may close this tab.")
		browserOpened = false
	})

	http.HandleFunc("/action", func(res http.ResponseWriter, req *http.Request) {
		enableCors(&res, req)
		res.Header().Set("Content-Type", "application/json")
		action := req.URL.Query().Get("a")
		
		switch action {
		case "":
			res.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(res, "{\"error\": \"400: Bad Request\"}")
			return
		
		case "current":
			returnJSON := "{\"status\": \"%s\", \"version\": \"%s\", \"name\": \"%s\", \"debug\": \"%s\"}"

			if auth == nil {
				fmt.Fprintf(res, returnJSON, "Not logged in", version, "", "NA")
				return
			}
			if previousActivity.Details == "" {
				fmt.Fprintf(res, returnJSON, "Not playing Destiny 2", version, auth.DisplayName, "NA")
				return
			}

			status := previousActivity.Details
			if previousActivity.State != "" {
				status += fmt.Sprintf(" | %s", previousActivity.State)
			}
			if previousActivity.SmallText != "" {
				status += fmt.Sprintf(" | %s", previousActivity.SmallText)
			}

			fmt.Fprintf(res, returnJSON, status, version, auth.DisplayName, debugHashes)
		}
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

func makePath(e string) string {
	return filepath.Join(currentDirectory, e)
}

func enableCors(res *http.ResponseWriter, req *http.Request) {
	origin := req.Header.Get("Origin")
	allowedOrigins := [...]string{"http://localhost:5500", "https://lieuweberg.com"}
	for _, o := range allowedOrigins {
		if o == origin {
			(*res).Header().Set("Access-Control-Allow-Origin", origin)
			break
		}
	}
}
