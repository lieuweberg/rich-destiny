package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	richgo "github.com/lieuweberg/rich-go/client"
)

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
			if storage == nil {
				res.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(res, "Please Authenticate before saving settings.")
				return
			}
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
				log.Printf("Error after hitting Update button: %s", err)
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
