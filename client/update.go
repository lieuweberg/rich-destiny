package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	update "github.com/inconshreveable/go-update"
	"golang.org/x/mod/semver"
)

var isUpdating bool
var tryPatches = true

// The version after an update. Else it will try updating itself again because the actual version variable hasn't changed.
var updatedVersion = version

func attemptApplicationUpdate() (string, error) {
	if updatedVersion == "dev" {
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
		return "", nil
	}

	if tryPatches {
		for i := len(releases) - 1; i >= 0; i-- {
			path := "rich-destiny.exe.dump"
			if i == len(releases)-1 {
				path = "rich-destiny.exe.old"
			}
			err = updateWithOldSavePath(releases[i], path)
			if err != nil {
				log.Printf("Error trying to apply update with patches: %s. Trying exe...", err)
				tryPatches = false
				// the returned function will run before the defer, so explicitly setting it to false here prevents the Not so fast error
				isUpdating = false
				return attemptApplicationUpdate()
			}
		}
	} else {
		err = updateWithOldSavePath(releases[0], "rich-destiny.exe.old")
		if err != nil {
			return "", err
		}
	}

	// We patched twice, so we should remove the .dump file since that's not in active use nor needed in the future
	if len(releases) > 1 && tryPatches {
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
	body, err := io.ReadAll(res.Body)
	res.Body.Close()

	err = json.Unmarshal(body, &releases)
	if err != nil {
		return nil, fmt.Errorf("Error trying to parse latest release response: %s", err)
	}

	return filterReleases(releases), nil
}

func filterReleases(releases releasesFromGithub) releasesFromGithub {
	// Remove all the latest prereleases up to a normal release since they shouldn't be downloaded if prereleases are disabled
	if !storage.Prereleases {
		for i, r := range releases {
			if !r.Prerelease {
				releases = releases[i:]
				break
			}
		}
	}
	// Get all the new releases (except the ones filtered away above)
	for i, r := range releases {
		if !r.Draft {
			// Should return where the version was the same, no prerelease checks since they were already filtered away before
			if semver.Compare(r.Name, updatedVersion) != 1 {
				return releases[:i]
			}
		}
	}

	return []releaseElement{}
}

func updateWithOldSavePath(release releaseElement, oldPath string) error {
	assetType := "patch"
	if !tryPatches {
		assetType = "exe"
	}

	log.Printf("Attempting to apply update for version %s", release.Name)

	for _, asset := range release.Assets {
		if asset.Name == "rich-destiny."+assetType {

			res, err := http.Get(asset.BrowserDownloadURL)
			if err != nil {
				return fmt.Errorf("Could not request download url of new update: %s", err)
			}

			checksum, err := hex.DecodeString(getChecksumFromBody(release.Body))
			if err != nil {
				return fmt.Errorf("Error decoding checksum: %s", err)
			}

			opts := update.Options{
				Checksum:    checksum,
				OldSavePath: filepath.Join(currentDirectory, oldPath),
			}
			if tryPatches {
				opts.Patcher = update.NewBSDiffPatcher()
			}

			err = update.Apply(res.Body, opts)
			if err != nil {
				if rerr := update.RollbackError(err); rerr != nil {
					return fmt.Errorf("Failed to roll back from bad update: %s", rerr)
				}
				return fmt.Errorf("Error while applying update: %s", err)
			}
			log.Printf("Successfully applied update for version %s with %s", release.Name, assetType)
			updatedVersion = release.Name
			return nil
		}
	}
	return fmt.Errorf("Release does not seem to include a rich-destiny.%s file, no update happened", assetType)
}

func getChecksumFromBody(s string) string {
	s = strings.SplitN(s, "\r\n", 2)[0] // SHA256: FAFSS51251
	return strings.SplitN(s, " ", 2)[1] // FAFSS51251
}
