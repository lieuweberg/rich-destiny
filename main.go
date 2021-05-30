package main

import (
	"archive/zip"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"time"

	richgo "github.com/hugolgst/rich-go/client"
	"github.com/kardianos/service"
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
	quitPresenceTicker  chan (struct{})
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
	manifest.Close()
	close(quitPresenceTicker)
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
		defer func() {
			fmt.Println("\n Press ENTER to close this window.")
			fmt.Scanln()
		}()

		fmt.Print("         _      _              _           _   _\n        (_)    | |            | |         | | (_)\n    _ __ _  ___| |__ ______ __| | ___  ___| |_ _ _ __  _   _\n   | '__| |/ __| '_ \\______/ _` |/ _ \\/ __| __| | '_ \\| | | |\n   | |  | | (__| | | |    | (_| |  __/\\__ \\ |_| | | | | |_| |\n   |_|  |_|\\___|_| |_|     \\__,_|\\___||___/\\__|_|_| |_|\\__, |\n                                                        __/ |\n                                                       |___/    ",
			version, "\n\n\n")
		log.SetFlags(log.Lshortfile)

		// This works because it deletes/starts/detects by name. Path does not matter, except when installing.
		createService()
		status, err := s.Status()
		if err != nil && !errors.Is(err, service.ErrNotInstalled) {
			log.Printf("Error trying to detect service status: %s", err)
		}

		if status == service.StatusRunning {
			fmt.Println(" rich-destiny is already installed and running. Open the control panel at:  https://richdestiny.app/cp")
			return
		} else if status == service.StatusStopped {
			fmt.Println(" rich-destiny is already installed but not running. Attempting to start it...")
			started, err := successfullyStartService()
			if err != nil {
				if err.Error() == "The system cannot find the file specified." {
					fmt.Println(" Windows can't find the file where you installed rich-destiny previously.\n\n Do you want to uninstall the original location so you can reinstall here?")
					for {
						fmt.Print("\n Choose: [Yes/No]: ")
						r, err := readUserInput()
						if err != nil {
							log.Printf(" Unable to read your input...: %s", err)
							return
						}
						if strings.Contains(r, "y") {
							err = s.Uninstall()
							if err != nil {
								log.Printf("Error trying to uninstall the service: %s", err)
							}
							fmt.Print("\n Uninstalled. Starting new installation now.\n\n")
							break
						} else if strings.Contains(r, "n") {
							fmt.Println(" Okay. If you need help, please join the support server!  https://discord.gg/UNU4UXp")
							return
						} else {
							fmt.Println(" Invalid response.")
						}
					}
				}
			} else if started {
				fmt.Println(" rich-destiny was successfully started! Find the control panel at:  https://richdestiny.app/cp")
				return
			} else { // it didn't start, but that message was already printed so just return here
				return
			}
		}

		fmt.Println(" Welcome to the rich-destiny setup!")

		home, err := os.UserHomeDir()
		if err != nil {
			log.Printf("Could not get home directory...: %s", err)
		}
		var downloadDir bool
		if currentDirectory == filepath.Join(home, "Downloads") {
			downloadDir = true
			fmt.Println("\n This program will refuse to install in the Downloads folder as it often will not work from there.")
		} else {
			fmt.Println("\n You can install the program in either the default (recommended) folder or the current folder.")
		}

		defaultDirectory := filepath.Join(home, "rich-destiny")
		fmt.Printf(" If you want to install the program in another folder: Exit, move this file to the desired folder and then run it again.\n\n → Default: %s", defaultDirectory)
		if !downloadDir {
			fmt.Printf("\n → Current: %s", currentDirectory)
		}

		for {
			fmt.Print("\n\n  > Choose a location: [Default/")
			if !downloadDir {
				fmt.Print("Current/")
			}
			fmt.Print("Exit]: ")

			r, err := readUserInput()
			if err != nil {
				log.Printf(" Unable to read your input...: %s", err)
				return
			}
			if strings.Contains(r, "d") {
				fmt.Println(" Okay, attempting to move there...")

				err = os.Mkdir(defaultDirectory, os.ModePerm)
				if err != nil && !errors.Is(err, os.ErrExist) {
					log.Printf("Error trying to create %s\\rich-destiny folder: %s", home, err)
					return
				}

				oldExe := exe
				exe = filepath.Join(defaultDirectory, "rich-destiny.exe")

				if oldExe[0] != exe[0] {
					newFile, err := os.Create(exe)
					if err != nil {
						log.Printf("Error creating file at new location: %s", err)
						return
					}
					oldFile, err := os.Open(oldExe)
					if err != nil {
						log.Printf("Error opening old file: %s", err)
						return
					}
					_, err = io.Copy(newFile, oldFile)
					if err != nil {
						log.Printf("Error copying file: %s", err)
						return
					}
					newFile.Close()
					oldFile.Close()
				} else {
					err = os.Rename(oldExe, exe)
					if err != nil {
						log.Printf("Error moving rich-destiny.exe to new location: %s", err)
						return
					}
				}

				fmt.Println(" Successfully moved.")
				break
			} else if strings.Contains(r, "c") {
				if downloadDir {
					fmt.Println(" Okay, move this program to a different directory manually and run it from there.")
					return
				}
				fmt.Println(" Okay, installing at the current location.")
				break
			} else if strings.Contains(r, "x") {
				fmt.Println(" Okay, exiting. If you intend to move this program to another folder, make sure to close this window first.")
				return
			} else {
				fmt.Println(" Invalid response. Please reply with Default, Current or Exit.")
			}
		}

		createService()
		err = s.Install()
		if err != nil {
			log.Printf("Error adding rich-destiny to the service manager: %s", err)
			return
		}

		fmt.Println(" Done! Waiting for rich-destiny to start...")

		started, err := successfullyStartService()
		if err != nil {
			log.Printf("Error trying to start rich-destiny: %s", err)
			return
		} else if !started {
			return
		}

		fmt.Println(" Done! Opening a browser tab to log in with Bungie.net. Setup is now complete and you can close this window.")
		openOauthTab()
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

	go startWebServer()

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
					log.Print(err)
				}
			}()
		}
	}

	var manifestExists bool
	// The following section returns on most errors, so defer this function (long manifest downloads can cause issues for initPresence, too)
	defer func() {
		if manifestExists {
			manifest, err = sql.Open("sqlite3", makePath("manifest.db"))
			if err != nil {
				log.Printf("Error opening manifest.db. This program will now exit since without a manifest, it can't do anything: %s", err)
				s.Stop()
				return
			}
		} else {
			log.Printf("No manifest exists and could not download new one. See errors above. This program will now exit since without a manifest, it can't do anything.")
			s.Stop()
			return
		}

		initPresence()
	}()

	// Check if a new manifest has to be downloaded, if so do that, then open the db
	manifestRes, err := getManifestData()
	if err != nil {
		log.Printf("Error getting manifest data: %s", err)
	}

	var lastManifestURL string
	err = db.QueryRow("SELECT value FROM data WHERE key='lastManifestURL'").Scan(&lastManifestURL)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Error querying database for lastManifestURL. Obtaining new manifest: %s", err)
		}
	}

	if _, err := os.Stat(makePath("manifest.db")); os.IsNotExist(err) || manifestRes.Response.MobileWorldContentPaths.En != lastManifestURL {
		if os.IsNotExist(err) {
			log.Print("Manifest doesn't exist, downloading one...")
		} else {
			manifestExists = true
			log.Print("Manifest is outdated, downloading a new one...")
		}

		res, err := http.Get("https://www.bungie.net" + manifestRes.Response.MobileWorldContentPaths.En)
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
		manifestExists = true

		err = os.Remove(makePath("manifest.zip"))
		if err != nil {
			log.Printf("Error deleting manifest.zip: %s", err)
			return
		}
		log.Print("Deleted temporary file manifest.zip")

		err = storeData("lastManifestURL", manifestRes.Response.MobileWorldContentPaths.En)
		if err != nil {
			log.Printf("Error setting lastManifestURL to storage.db: %s", err)
			return
		}
	} else {
		manifestExists = true
	}
}

func startWebServer() {
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		enableCors(&res, req)
		fmt.Fprint(res, "hello")
	})

	var generatedState string
	http.HandleFunc("/login", func(res http.ResponseWriter, req *http.Request) {
		generatedState = randomString(20)
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

		res.Header().Set("Content-Type", "text/html")
		fmt.Fprint(res, "Success! You are now logged in and may close this tab (and head to <a href=\"https://richdestiny.app/cp\">the control panel</a>).")
	})

	http.HandleFunc("/action", func(res http.ResponseWriter, req *http.Request) {
		enableCors(&res, req)
		if req.Method == http.MethodOptions {
			return
		}
		res.Header().Set("Content-Type", "application/json")
		action := req.URL.Query().Get("a")

		switch action {
		case "":
			res.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(res, "error 400: Bad Request")
			return

		case "current":
			d := currentProgramStatus{
				Version:  version,
				Debug:    "NA",
				Status:   "Not logged in",
				Presence: previousActivity,
			}

			if storage == nil {
				returnStructAsJSON(res, d)
				return
			}

			d.Name = storage.DisplayName
			d.OrbitText = storage.OrbitText
			d.AutoUpdate = storage.AutoUpdate
			d.JoinGameButton = storage.JoinGameButton
			d.JoinOnlySocial = storage.JoinOnlySocial

			if previousActivity.Details == "" {
				d.Status = "Not playing Destiny 2"
				returnStructAsJSON(res, d)
				return
			}

			status := previousActivity.Details
			if previousActivity.State != "" {
				status += fmt.Sprintf(" | %s", previousActivity.State)
			}
			if previousActivity.SmallText != "" {
				status += fmt.Sprintf(" | %s", previousActivity.SmallText)
			}
			d.Status = status
			d.Debug = debugText
			returnStructAsJSON(res, d)
		case "save":
			if req.Method != http.MethodPost {
				return
			}
			data, err := ioutil.ReadAll(req.Body)
			req.Body.Close()
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(res, "error 500: %s", err)
				return
			}
			err = json.Unmarshal(data, storage)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(res, "error 500: %s", err)
				return
			}

			err = storeData("storage", storage)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(res, "error 500: could not save data: %s", err)
				return
			}

			fmt.Fprint(res, "OK")

			if previousActivity.Details != "" {
				forcePresenceUpdate = true
			}
		case "update":
			newVersion, err := attemptApplicationUpdate()
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(res, err)
			} else {
				fmt.Fprintf(res, "Update installed successfully; will be applied next startup (or restart rich-destiny from the Services manager). New version: %s", newVersion)
			}
		case "uninstall":
			err := s.Uninstall()
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(res, "Error trying to uninstall: %s", err)
				return
			}
			err = s.Stop()
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(res, "Error trying to stop service: %s", err)
				return
			}
		case "reconnect":
			if previousActivity.Details != "" {
				richgo.Logout()
				err := richgo.Login("726090012877258762")
				if err != nil {
					fmt.Fprintf(res, "Couldn't connect to Discord: %s", err)
					return
				}
				forcePresenceUpdate = true
			} else {
				fmt.Fprintf(res, "Can't reconnect when there's nothing to connect to! You're not playing Destiny 2.")
				return
			}

			fmt.Fprint(res, "Successfully reconnected.")
		// case "restart":
		// 	err := s.Restart()
		// 	if err != nil {
		// 		res.WriteHeader(http.StatusInternalServerError)
		// 		fmt.Fprintf(res, "Error trying to restart: %s", err)
		// 	}

		// 	fmt.Fprintf(res, "OK")
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

func readUserInput() (string, error) {
	var r string
	_, err := fmt.Scanln(&r)
	return strings.ToLower(r), err
}

func enableCors(res *http.ResponseWriter, req *http.Request) {
	origin := req.Header.Get("Origin")
	allowedOrigins := [...]string{"https://lieuweberg.com", "http://localhost:1234", `https://richdestiny.app`}
	for _, o := range allowedOrigins {
		if o == origin {
			(*res).Header().Set("Access-Control-Allow-Origin", origin)
			(*res).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			(*res).Header().Set("Access-Control-Allow-Headers", "*")
			break
		}
	}
}

func returnStructAsJSON(res http.ResponseWriter, data interface{}) {
	d, err := json.Marshal(data)
	if err != nil {
		fmt.Fprintf(res, "error 500: marshaling struct: %s", err)
		return
	}
	fmt.Fprint(res, string(d))
}

func randomString(length uint8) string {
	rand.Seed(time.Now().UnixNano())
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
