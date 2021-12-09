package main

import (
	"archive/zip"
	"database/sql"
	"io"
	"log"
	"net/http"
	"os"
)

func getManifest() {
	log.Print("hello")
	if quitPresenceTicker != nil {
		close(quitPresenceTicker)
	}
	log.Print("hello2")
	var manifestExists bool
	// The following section returns on most errors, so defer this function (long manifest downloads can cause issues for initPresence, too)
	defer func() {
		if manifestExists {
			var err error
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