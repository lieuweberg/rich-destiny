package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// requestAccessToken requests an access token from bungie.
// code is the code param if refresh is false.
// code is the refresh token param is refresh is true.
func requestAccessToken(code string, refresh bool) (err error) {
	data := url.Values{}
	if refresh {
		logInfoIfNoErrorSpam(errorOriginAuth, "Refreshing access token")
		data.Set("grant_type", "refresh_token")
		data.Set("refresh_token", code)
	} else {
		logInfoIfNoErrorSpam(errorOriginAuth, "Requesting access token with code: "+code)
		data.Set("grant_type", "authorization_code")
		data.Set("code", code)
	}
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)

	authReq, err := http.NewRequest("POST", "https://www.bungie.net/platform/app/oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("Error making NewRequest to get an access token: %s", err)
	}
	addUserAgent(authReq)

	authReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	authRes, err := http.DefaultClient.Do(authReq)
	if err != nil {
		return fmt.Errorf("Error doing http request: %s", err)
	}
	body, err := io.ReadAll(authRes.Body)
	authRes.Body.Close()
	if err != nil {
		return fmt.Errorf("Error reading response body: %s", err)
	}

	err = setAuth(body)
	if err != nil {
		return fmt.Errorf("Error setting auth details: %s", err)
	}

	if refresh {
		log.Print("Refreshed access token")
	} else {
		log.Print("Successfully authorised")
	}
	return
}

// setAuth sets the auth response from the bungie api (to memory and the database)
func setAuth(data []byte) (err error) {
	var errorResponse *oauthError
	err = json.Unmarshal(data, &errorResponse)
	if err != nil {
		return
	}

	if errorResponse.ErrorDescription != "" {
		return fmt.Errorf("Error response to the request: %s", errorResponse.ErrorDescription)
	}

	err = json.Unmarshal(data, &storage)
	if err != nil {
		return
	}

	// Subtracting five to make sure tokens are refreshed on-time, and not used a few milliseconds late
	storage.RefreshAt = time.Now().Unix() + storage.ExpiresIn - 5
	storage.ReAuthAt = time.Now().Unix() + storage.RefreshExpiresIn - 5

	var lp *linkedProfiles
	err = requestComponents(fmt.Sprintf("/Destiny2/254/Profile/%s/LinkedProfiles/", storage.BungieMSID), &lp)
	if err != nil {
		return fmt.Errorf("Error requesting linked profiles: %s", err)
	}
	if lp.ErrorStatus != "Success" {
		storage = nil
		return fmt.Errorf("Bungie returned an error status %s when trying to find your profiles, message: %s", lp.ErrorStatus, lp.Message)
	}

	for _, profile := range lp.Response.Profiles {
		for _, membershipType := range profile.MembershipTypes {
			if membershipType == 3 || membershipType == 6 {
				storage.BungieName = profile.BungieGlobalDisplayName
				storage.ActualMSID = profile.MembershipID
				storage.MSType = profile.MembershipType

				code := strconv.Itoa(int(profile.BungieGlobalDisplayNameCode))
				// 0 is removed from the start of codes since the data type is an int, so we have to add them back manually
				code = strings.Repeat("0", 4-len(code)) + code
				storage.BungieCode = code

				// var cta *credentialsTargetAccount
				// err = requestComponents(fmt.Sprintf("/User/GetCredentialTypesForTargetAccount/%s/", storage.ActualMSID), &cta)
				// if err != nil {
				// 	return
				// }
				// for _, cred := range cta.Response {
				// 	if cred.CredentialType == 12 {
				// 		storage.SteamID64 = cred.CredentialAsString
				// 	}
				// }

				break
			}
		}
		if storage.ActualMSID != "" {
			break
		}
	}

	var dud string
	err = db.QueryRow("SELECT value FROM data WHERE key='storage'").Scan(&dud)
	if err != nil {
		if err == sql.ErrNoRows {
			// Default settings
			storage.AutoUpdate = true
		} else {
			logInfoIfNoErrorSpam(errorOriginAuth, "Error trying to query the database: "+err.Error())
		}
	}

	err = storeData("storage", storage)
	return
}

// storeData inserts data into the db by key-value pair. Only store strings or the storage struct.
func storeData(key string, data interface{}) (err error) {
	var d string
	switch data.(type) {
	case string:
		d = data.(string)
	default:
		jsonBytes, err := json.Marshal(storage)
		if err != nil {
			return err
		}
		d = string(jsonBytes)
	}
	_, err = db.Exec("INSERT OR REPLACE INTO data(key, value) VALUES($1, $2)", key, d)
	return
}

// getStorage gets the AuthResponse and storage values from the database, or if there is none
// (or the refresh token is expired) tries to initiate an oauth tab in the browser.
// This function also refreshes the auth token.
func getStorage() (s *storageStruct, err error) {
	if storage == nil {
		var r string
		err = db.QueryRow("SELECT value FROM data WHERE key='storage'").Scan(&r)
		if err == sql.ErrNoRows {
			err = fmt.Errorf("no existing storage found, please go to https://richdestiny.app/cp to log in (click Reauthenticate)")
			return
		} else if err != nil {
			return
		}
		err = json.Unmarshal([]byte(r), &storage)
		if err != nil {
			return
		}
		s, err = getStorage()
		if err != nil {
			return nil, err
		}
	} else if time.Now().Unix() >= storage.ReAuthAt {
		logErrorIfNoErrorSpam(errorOriginAuth, "Your authentication details have expired. Please go to https://rich-destiny.app/cp to Reauthenticate again.")
		return
	} else if time.Now().Unix() >= storage.RefreshAt {
		err = requestAccessToken(storage.RefreshToken, true)
		if err != nil {
			return
		}
	}

	return storage, nil
}

// openTab opens a browser tab with the given URL
func openTab(url string) {
	err := exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	if err != nil {
		log.Printf("Error executing browser open command: %s", err)
	}
}

// openOauthTab tries to open the browser with the bungie oauth authorisation page.
// It is used when first launching.
func openOauthTab() {
	openTab("http://localhost:35893/login")
}

// requestComponents is a helper function to request an endpoint/component from the bungie api.
// You MUST make sure that storage is populated.
// endpoint MUST start with a '/'!
func requestComponents(endpoint string, responseStruct interface{}) (err error) {
	if bungieHTTPClient == nil {
		cookieJar, err := cookiejar.New(nil)
		if err == nil {
			bungieHTTPClient = &http.Client{
				Jar:     cookieJar,
				Timeout: time.Second * 10,
			}
		} else {
			bungieHTTPClient = &http.Client{}
			log.Printf("Couldn't create cookie jar, resolving to http client without it. This can cause some \"stuttery\" presence: %s", err)
		}
	}

	req, err := http.NewRequest("GET", "https://www.bungie.net/Platform"+endpoint, nil)
	if err != nil {
		return
	}
	addUserAgent(req)

	req.Header.Set("X-API-Key", config.APIKey)
	if endpoint != "/Destiny2/Manifest/" {
		req.Header.Set("Authorization", "Bearer "+storage.AccessToken)
	}

	res, err := bungieHTTPClient.Do(req)
	if err != nil {
		return
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &responseStruct)

	return
}

func addUserAgent(req *http.Request) {
	req.Header.Set("User-Agent", fmt.Sprintf("rich-destiny/%s AppId/%s (+richdestiny.app;@lieuwe_berg)", version, config.ClientID))
}
