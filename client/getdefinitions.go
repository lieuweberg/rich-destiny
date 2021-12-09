package main

import (
	"archive/zip"
	"database/sql"
	"io"
	"log"
	"net/http"
	"os"
)

var definitionsExist bool

func getDefinitions() {
	// The following section returns on most errors, so defer this function
	defer func() {
		if definitionsExist {
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
	var manifestRes *manifestData
	err := requestComponents("/Destiny2/Manifest/", &manifestRes)
	if err != nil {
		log.Printf("Error getting manifest data: %s", err)
		return
	}

	var lastDefinitionsURL string
	err = db.QueryRow("SELECT value FROM data WHERE key='lastManifestURL'").Scan(&lastDefinitionsURL)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Error querying database for lastManifestURL. Obtaining new manifest: %s", err)
		}
	}

	if _, err := os.Stat(makePath("manifest.db")); os.IsNotExist(err) || manifestRes.Response.MobileWorldContentPaths.En != lastDefinitionsURL {
		if os.IsNotExist(err) {
			log.Print("Manifest doesn't exist, downloading one...")
		} else {
			definitionsExist = true
			log.Print("Manifest is outdated, downloading a new one...")
		}
		
		if quitPresenceTicker != nil {
			close(quitPresenceTicker)
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
		definitionsExist = true

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
		definitionsExist = true
	}
}