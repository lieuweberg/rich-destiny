package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/inconshreveable/go-update"
	"golang.org/x/mod/semver"
)

func attemptApplicationUpdate() {
	res, err := http.Get("https://api.github.com/repos/lieuweberg/rich-destiny/releases")
	if err != nil {
		log.Printf("Error trying to get latest release: %s", err)
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	var releases releasesFromGithub
	err = json.Unmarshal(body, &releases)
	if err != nil {
		log.Printf("Error trying to parse latest release response: %s", err)
		return
	}

	for _, release := range releases {
		if !release.Draft && !release.Prerelease {
			if semver.Compare(release.Name, version) == 1 {
				log.Printf("Attempting to update to %s (currently %s)", release.Name, version)
				var foundAsset bool
				for _, asset := range release.Assets {
					if asset.Name == "rich-destiny.exe" {
						foundAsset = true

						res, err = http.Get(asset.BrowserDownloadURL)
						if err != nil {
							log.Printf("Could not get download url of new update: %s", err)
							break
						}
						err = update.Apply(res.Body, update.Options{})
						if err != nil {
							if rerr := update.RollbackError(err); rerr != nil {
								log.Printf("Failed to roll back from bad update: %s", rerr)
							} else {
								log.Printf("Error while applying update: %s", err)
							}
						} else {
							log.Printf("Update applied successfully.")
						}

						break
					}
				}
				if foundAsset == false {
					log.Print("Latest release does not seem to include a rich-destiny.exe file, no update happened")
				}
			}
		}
	}
}