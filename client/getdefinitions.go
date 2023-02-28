package main

import (
	"archive/zip"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
)

func getDefinitions() (err error) {
	// The following section returns on most errors, so defer this function
	defer func() {
		if _, err := os.Stat(makePath("manifest.db")); !os.IsNotExist(err) {
			var err error
			manifest, err = sql.Open("sqlite3", makePath("manifest.db"))
			if err != nil {
				logInfoIfNoErrorSpam("Error opening manifest.db: " + err.Error())
				return
			}
			logInfoIfNoErrorSpam("Using existing manifest")
		} else {
			logInfoIfNoErrorSpam("No manifest exists and could not download new one. See errors above.")
		}
	}()

	// Check if a new manifest has to be downloaded, if so do that, then open the db
	var manifestRes *manifestData
	err = requestComponents("/Destiny2/Manifest/", &manifestRes)
	if err != nil {
		return fmt.Errorf("Error getting manifest data: %s", err)
	}
	if manifestRes.ErrorStatus != "Success" {
		return fmt.Errorf("Bungie returned an error status %s when trying to get the manifest, message: %s", manifestRes.ErrorStatus, manifestRes.Message)
	}

	var lastDefinitionsURL string
	err = db.QueryRow("SELECT value FROM data WHERE key='lastManifestURL'").Scan(&lastDefinitionsURL)
	if err != nil {
		if err != sql.ErrNoRows {
			logInfoIfNoErrorSpam(fmt.Sprintf("Error querying database for lastManifestURL. Obtaining new manifest: %s", err))
		}
	}

	if manifestRes.Response.MobileWorldContentPaths.En != lastDefinitionsURL {
		logInfoIfNoErrorSpam("Updating manifest...")

		res, err := http.Get("https://www.bungie.net" + manifestRes.Response.MobileWorldContentPaths.En)
		if err != nil {
			return fmt.Errorf("Error getting manifest database: %s", err)
		}
		out, err := os.Create(makePath("manifest.zip"))
		if err != nil {
			return fmt.Errorf("Error creating manifest.zip: %s", err)
		}
		_, err = io.Copy(out, res.Body)
		res.Body.Close()
		logInfoIfNoErrorSpam("Manifest downloaded, unzipping...")

		z, err := zip.OpenReader(out.Name())
		out.Close()
		if err != nil {
			return fmt.Errorf("Error writing/unzipping manifest.zip: %s", err)
		}

		for _, f := range z.File {
			file, err := f.Open()
			if err != nil {
				return fmt.Errorf("Error opening file: %s", err)
			}
			out, err := os.Create(makePath("manifest.db"))
			if err != nil {
				return fmt.Errorf("Error creating manifest.db: %s", err)
			}
			_, err = io.Copy(out, file)
			if err != nil {
				return fmt.Errorf("Error writing manifest.db: %s", err)
			}
			file.Close()
			out.Close()
		}
		z.Close()
		logInfoIfNoErrorSpam("Manifest downloaded and unzipped!")

		err = os.Remove(makePath("manifest.zip"))
		if err != nil {
			return fmt.Errorf("Error deleting manifest.zip: %s", err)
		}
		logInfoIfNoErrorSpam("Deleted temporary file manifest.zip")

		err = storeData("lastManifestURL", manifestRes.Response.MobileWorldContentPaths.En)
		if err != nil {
			return fmt.Errorf("Error setting lastManifestURL to storage.db: %s", err)
		}
	}

	return
}
