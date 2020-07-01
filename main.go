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
	svcConfig := &service.Config{
		Name: "rich-destiny",
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
		} else {
			log.Print("Program installed.")
			s.Start()
		}
		log.Print("Press ENTER to close this window. The service should automatically start :D")
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
	currentDirectory = filepath.Dir(exe);

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
		printErr(err)
	}

	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS data(
		key		STRING	PRIMARY KEY NOT NULL,
		value	STRING	NOT NULL
	)`); err != nil {
		printErr(err)
	}
	if _, err = getAuth(); err != nil {
		printErr(err)
	}

	// Check if a new manifest has to be downloaded, if so do that, then open the db
	manifestRes, err := getManifestData()
	if err != nil {
		printErr(err)
		return
	}
	var lastManifestURL string
	db.QueryRow("SELECT value FROM data WHERE key='lastManifestURL'").Scan(&lastManifestURL)
	if _, err := os.Stat(makePath("manifest.db")); os.IsNotExist(err) ||  manifestRes.Response.MobileWorldContentPaths.En != lastManifestURL {
		if os.IsNotExist(err) {
			log.Print("Manifest doesn't exist, downloading one...")
		} else {
			log.Print("Manifest is outdated, downloading a new one...")
		}

		res, err := http.Get("https://bungie.net" + manifestRes.Response.MobileWorldContentPaths.En);
		if err != nil {
			printErr(err)
			return
		}
		out, err := os.Create(makePath("manifest.zip"))
		if err != nil {
			printErr(err)
			return
		}
		_, err = io.Copy(out, res.Body)
		res.Body.Close()
		log.Print("Manifest downloaded, unzipping...")
		
		z, err := zip.OpenReader(out.Name())
		out.Close()
		if err != nil {
			printErr(err)
			return
		}
		var success bool
		for _, f := range z.File {
			file, err := f.Open()
			if err != nil {
				printErr(err)
				break
			}
			out, err := os.Create(makePath("manifest.db"))
			if err != nil {
				printErr(err)
				break
			}
			_, err = io.Copy(out, file)
			if err != nil {
				printErr(err)
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
			printErr(err)
			return
		}
		log.Print("Deleted temporary file manifest.zip")

		_, err = db.Exec("INSERT OR REPLACE INTO data(key, value) VALUES('lastManifestURL', $1)", manifestRes.Response.MobileWorldContentPaths.En)
		if err != nil {
			printErr(err)
			return
		}
	}

	manifest, err = sql.Open("sqlite3", makePath("manifest.db"))
	if err != nil {
		printErr(err)
		return
	}

	startWebServer()
	initPresence()
}

func startWebServer() {
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		enableCors(&res)
		fmt.Fprint(res, "{message: \"hello\"}")
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
			printErr(err)
		}
	}()
}

func makePath(e string) string {
	return filepath.Join(currentDirectory, e)
}

func enableCors(res *http.ResponseWriter) {
	(*res).Header().Set("Access-Control-Allow-Origin", "http://localhost:5500")
}

func printErr(err error) {
	log.Printf("{error: \"%s\"}", err)
}