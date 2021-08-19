package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	update "github.com/inconshreveable/go-update"
	"golang.org/x/mod/semver"
)

var isUpdating bool

func attemptApplicationUpdate() (string, error) {
	if version == "dev" {
		return "", fmt.Errorf("Version 'dev' does not allow updates")
	}

	if isUpdating {
		return "", fmt.Errorf("Not so fast! Already trying to update")
	}
	isUpdating = true
	defer func() {
		isUpdating = false
	}()

	releases, err := getNewReleases()
	if err != nil {
		log.Print(err)
	}

	if len(releases) == 0 {
		return "", fmt.Errorf("No newer version found.")
	}

	for i := len(releases) - 1; i >= 0; i-- {
		path := "rich-destiny.exe.dump"
		if i == len(releases)-1 {
			path = "rich-destiny.exe.old"
		}
		err = updateWithOldSavePath(releases[i], path)
		if err != nil {
			return "", err
		}
	}

	// We patched twice, so we should remove the .dump file since that's not in active use nor needed in the future
	if len(releases) > 1 {
		err = os.Remove("rich-destiny.exe.dump")
		if err != nil {
			return "", fmt.Errorf("Error removing the .dump file, %s", err)
		}
	}

	return releases[0].Name, nil
}

func getNewReleases() (releases releasesFromGithub, err error) {
	res, err := http.Get("https://api.github.com/repos/lieuweberg/rich-destiny/releases")
	if err != nil {
		return nil, fmt.Errorf("Error trying to get latest release: %s", err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	err = json.Unmarshal(body, &releases)
	if err != nil {
		return nil, fmt.Errorf("Error trying to parse latest release response: %s", err)
	}

	return filterReleases(releases), nil
}

func filterReleases(releases releasesFromGithub) releasesFromGithub {
	for i, r := range releases {
		if !r.Draft && !r.Prerelease {
			if semver.Compare(r.Name, version) == 1 {
				releases = releases[:i+1]
			}
		}
	}

	return releases
}

func updateWithOldSavePath(release releaseElement, path string) error {
	for _, asset := range release.Assets {
		if asset.Name == "rich-destiny.patch" {
			log.Printf("Attempting to apply patch for version %s", release.Name)

			res, err := http.Get(asset.BrowserDownloadURL)
			if err != nil {
				return fmt.Errorf("Could not request download url of new update: %s", err)
			}

			checksum, err := hex.DecodeString(getChecksumFromBody(release.Body))
			if err != nil {
				return fmt.Errorf("Error decoding checksum: %s", err)
			}

			err = update.Apply(res.Body, update.Options{
				Checksum:    checksum,
				Patcher:     update.NewBSDiffPatcher(),
				OldSavePath: path,
			})
			if err != nil {
				if rerr := update.RollbackError(err); rerr != nil {
					return fmt.Errorf("Failed to roll back from bad update: %s", rerr)
				}
				return fmt.Errorf("Error while applying update: %s", err)
			}
			log.Printf("Successfully applied update for version %s", release.Name)
			return nil
		}
	}
	return fmt.Errorf("Release does not seem to include a rich-destiny.patch file, no update happened")
}

func getChecksumFromBody(s string) string {
	s = strings.SplitN(s, "\r\n", 2)[0] // SHA256: FAFSS51251
	return strings.SplitN(s, " ", 2)[1] // FAFSS51251
}
