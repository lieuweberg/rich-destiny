package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

// requestAccessToken requests an access token from bungie.
// code is the code param if refresh is false.
// code is the refresh token param is refresh is true.
func requestAccessToken(code string, refresh bool) (err error) {
	data := url.Values{}
	if refresh {
		log.Print("Refreshing access token")
		data.Set("grant_type", "refresh_token")
		data.Set("refresh_token", code)
	} else {
		log.Print("Requesting access token with code: " + code)
		data.Set("grant_type", "authorization_code")
		data.Set("code", code)
	}
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)

	authReq, err := http.NewRequest("POST", "https://www.bungie.net/platform/app/oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		log.Printf("Trouble making NewRequest to get an access token: %s", err)
		return
	}
	authReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	authRes, err := http.DefaultClient.Do(authReq)
	if err != nil {
		log.Printf("http client failed to do request: %s", err)
		return
	}
	body, err := ioutil.ReadAll(authRes.Body)
	authRes.Body.Close()
	if err != nil {
		log.Printf("Error reading response body: %s", err)
		return
	}

	err = setAuth(body)
	if err != nil {
		log.Printf("Error setting auth details: %s", err)
		return
	}

	if refresh == true {
		log.Print("Refreshed access token")
	} else {
		log.Print("Successfully authorised")
	}
	return
}

// setAuth sets the auth response from the bungie api (to memory and the database)
func setAuth(data []byte) (err error) {
	err = json.Unmarshal(data, &storage)
	if err != nil {
		return
	}
	// Subtracting five to make sure tokens are refreshed on-time, and not a few milliseconds late (sometimes causing 401's)
	storage.RefreshAt = time.Now().Unix() + storage.ExpiresIn - 5
	storage.ReAuthAt = time.Now().Unix() + storage.RefreshExpiresIn - 5

	var lp *linkedProfiles
	err = requestComponents(fmt.Sprintf("/Destiny2/254/Profile/%s/LinkedProfiles/", storage.BungieMSID), &lp)
	if err != nil {
		return
	}

	for _, profile := range lp.Response.Profiles {
		for _, membershipType := range profile.MembershipTypes {
			if membershipType == 3 {
				storage.DisplayName = profile.DisplayName
				storage.ActualMSID = profile.MembershipID
				storage.MSType = profile.MembershipType

				var cta *credentialsTargetAccount
				err = requestComponents(fmt.Sprintf("/User/GetCredentialTypesForTargetAccount/%s/", storage.ActualMSID), &cta)
				if err != nil {
					return
				}
				for _, cred := range cta.Response {
					if cred.CredentialType == 12 {
						storage.SteamID64 = cred.CredentialAsString
					}
				}

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
			storage.AutoUpdate = true
		} else {
			log.Printf("Error trying to query the database: %s", err)
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
	} else if time.Now().Unix() >= storage.ReAuthAt {
		log.Print("Your authentication details have expired. Please go to https://rich-destiny.app/cp to Reauthenticate again.")
		return
	} else if time.Now().Unix() >= storage.RefreshAt {
		requestAccessToken(storage.RefreshToken, true)
	}

	return storage, nil
}

// openOauthTab tries to open the browser with the bungie oauth authorisation page.
// Even though this does not work from within a service, it is used when first launching.
func openOauthTab() {
	err := exec.Command("rundll32", "url.dll,FileProtocolHandler", "http://localhost:35893/login").Start()
	if err != nil {
		log.Printf("Error executing browser open command: %s", err)
	}
}

// getManifestData gets the manifest info url
func getManifestData() (d manifestData, err error) {
	res, err := http.Get("https://www.bungie.net/Platform/Destiny2/Manifest/")
	if err != nil {
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &d)
	return
}

// requestComponents is a helper function to request an endpoint/component from the bungie api.
// You MUST make sure that auth is populated.
// url MUST start with a '/'!
func requestComponents(url string, responseStruct interface{}) (err error) {
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

	url = "https://www.bungie.net/Platform" + url
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("X-API-Key", config.APIKey)
	req.Header.Set("Authorization", "Bearer "+storage.AccessToken)

	res, err := bungieHTTPClient.Do(req)
	if err != nil {
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &responseStruct)

	return
}
