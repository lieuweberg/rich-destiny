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

var auth *authResponse

// requestAccessToken requests an access token from bungie.
// code is the ?code param if refresh is false.
// code is the refresh token is refresh is true.
func requestAccessToken(code string, refresh bool) (err error) {
	data := url.Values{}
	if refresh {
		log.Print("Refreshing access token with refresh token: " + code)
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
		printErr(err)
		return
	}
	authReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	authRes, err := http.DefaultClient.Do(authReq)
	if err != nil {
		printErr(err)
		return
	}
	body, err := ioutil.ReadAll(authRes.Body)
	if err != nil {
		printErr(err)
		return
	}
	authRes.Body.Close()

	err = setAuth(body)
	if err != nil {
		printErr(err)
	}

	return
}

// setAuth sets the auth response from the bungie api (to memory and the database)
func setAuth(data []byte) (err error) {
	err = json.Unmarshal(data, &auth)
	if err != nil {
		return
	}
	// Subtracting five to make sure tokens are refreshed on-time, and not a few milliseconds late (sometimes causing 401's)
	auth.RefreshAt = time.Now().Unix() + auth.ExpiresIn - 5
	auth.ReAuthAt = time.Now().Unix() + auth.RefreshExpiresIn - 5
	if auth.ActualMSID == "" {
		var lp *linkedProfiles
		err = requestComponents(fmt.Sprintf("/Destiny2/254/Profile/%s/LinkedProfiles/", auth.BungieMSID), &lp)
		if err != nil {
			return
		}

		for _, p := range lp.Response.Profiles {
			if p.MembershipType == 3 {
				auth.ActualMSID = p.MembershipID
				break
			}
		}

		log.Print(auth.ActualMSID)
	}

	d, err := json.Marshal(auth)
	if err != nil {
		return
	}
	_, err = db.Exec("INSERT OR REPLACE INTO data(key, value) VALUES('auth', $1)", string(d))
	return
}

// getAuth gets the AuthResponse from the database, or if there is none
// (or the refresh token is expired) tries to initiate an oauth tab in the browser.
// This function also refreshes the auth token.
func getAuth() (ar *authResponse, err error) {
	if auth == nil {
		var r string
		err = db.QueryRow("SELECT value FROM data WHERE key='auth'").Scan(&r)
		if err == sql.ErrNoRows {
			makeOauthTab()
			err = nil
			return
		} else if err != nil {
			return
		}
		err = json.Unmarshal([]byte(r), &auth)
		if err != nil {
			return
		}
		ar, err = getAuth()
	} else if time.Now().Unix() >= auth.ReAuthAt {
		makeOauthTab()
		return
	} else if time.Now().Unix() >= auth.RefreshAt {
		requestAccessToken(auth.RefreshToken, true)
	}

	ar = auth
	return
}

var browserOpened bool

// makeOauthTab tries to open the browser with /login (redirects to bungie).
// This uses the localhost path for convenience.
func makeOauthTab() {
	if !browserOpened {
		err := exec.Command("rundll32", "url.dll,FileProtocolHandler", "http://localhost:35893/login").Start()
		if err != nil {
			printErr(err)
		}
		browserOpened = true
		log.Printf("Opened Oauth in browser")
	}
}

// getManifestData gets the manifest info url
func getManifestData() (d *manifestData, err error) {
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

var httpClient *http.Client

// requestComponents is a helper function to request an endpoint/componenent from the bungie api.
// You MUST make sure that auth is populated.
// url MUST start with a '/'!
func requestComponents(url string, responseStruct interface{}) (err error) {
	if httpClient == nil {
		cookieJar, err := cookiejar.New(nil)
		if err == nil {
			httpClient = &http.Client{
				Jar: cookieJar,
			}
		} else {
			httpClient = &http.Client{}
			log.Print("Couldn't create cookie jar, resolving to http client without it. This can cause some \"stuttery\" presence.")
			printErr(err)
		}
	}

	url = "https://bungie.net/Platform" + url
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Add("X-API-Key", config.APIKey)
	req.Header.Add("Authorization", "Bearer " + auth.AccessToken)

	res, err := httpClient.Do(req)
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