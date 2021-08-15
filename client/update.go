package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/inconshreveable/go-update"
	"golang.org/x/mod/semver"
)

func attemptApplicationUpdate() (string, error) {
	if version == "dev" {
		return "", fmt.Errorf("version 'dev' does not allow updates")
	}

	res, err := http.Get("https://api.github.com/repos/lieuweberg/rich-destiny/releases")
	if err != nil {
		return "", fmt.Errorf("Error trying to get latest release: %s", err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	var releases releasesFromGithub
	err = json.Unmarshal(body, &releases)
	if err != nil {
		return "", fmt.Errorf("Error trying to parse latest release response: %s", err)
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
							return release.Name, fmt.Errorf("Could not get download url of new update: %s", err)
						}
						err = update.Apply(res.Body, update.Options{})
						if err != nil {
							if rerr := update.RollbackError(err); rerr != nil {
								return release.Name, fmt.Errorf("Failed to roll back from bad update: %s", rerr)
							}
							return release.Name, fmt.Errorf("Error while applying update: %s", err)
						}
						log.Printf("Update installed successfully; will be applied next startup. New version: %s", release.Name)
						return release.Name, nil
					}
				}
				if foundAsset == false {
					return release.Name, fmt.Errorf("Latest release does not seem to include a rich-destiny.exe file, no update happened")
				}
			}
		}
	}
	return "", nil
}