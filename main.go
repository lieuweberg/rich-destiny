package main

import (
	"archive/zip"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var generatedState string
var db *sql.DB
var manifest *sql.DB

func init() {
	// State query param
	rand.Seed(time.Now().UnixNano())
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 20)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	generatedState = string(b)

	// Open the storage db (access token, etc.)
	var err error
	db, err = sql.Open("sqlite3", "./storage.db")
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
	if _, err := os.Stat("./manifest.db"); os.IsNotExist(err) ||  manifestRes.Response.MobileWorldContentPaths.En != lastManifestURL {
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
		out, err := os.Create("manifest.zip")
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
			out, err := os.Create("manifest.db")
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

		err = os.Remove("manifest.zip")
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

	manifest, err = sql.Open("sqlite3", "./manifest.db")
	if err != nil {
		printErr(err)
		return
	}

	initPresence()
}

func main() {
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
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

	server := &http.Server{Addr: ":35893", Handler: nil}
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
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGTERM, syscall.SIGINT, os.Interrupt, os.Kill)
	<-sc
	log.Print("OS termination received")
	server.Shutdown(context.Background())
	db.Close()
	manifest.Close()
	close(quitExeCheckTicker)
	log.Print("Gracefully exited, bye bye")
}

func printErr(err error) {
	log.Printf("{error: \"%s\"}", err)
}