package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	richgo "github.com/hugolgst/rich-go/client"
	"github.com/mitchellh/go-ps"
)

func initPresence() {
	exeCheckTicker := time.NewTicker(15 * time.Second)
	quitPresenceTicker = make(chan struct{})
	loggedIn := false
	go func() {
		for {
			select {
			case <- exeCheckTicker.C:
				exeFound := false
				pl, _ := ps.Processes()
				for _, p := range pl {
					if p.Executable() == "destiny2.exe" {
						exeFound = true
						if !loggedIn {
							err := richgo.Login("726090012877258762")
							if err != nil {
								log.Print("Couldn't connect to Discord: " + err.Error())
								break
							}
							loggedIn = true
						}

						getAuth()
						if auth == nil {
							break
						}
						updatePresence()
						break
					}
				}
				if loggedIn && !exeFound {
					richgo.Logout()
					log.Print("No longer playing, logged ipc out")
					loggedIn = false
					previousActivity = richgo.Activity{}
				}
			case <- quitPresenceTicker:
				exeCheckTicker.Stop()
			}
		}
	}()
}

func updatePresence() {
	var ca *characterActivitiesDefinition
	err := requestComponents(fmt.Sprintf("/Destiny2/3/Profile/%s?components=204,200", auth.ActualMSID), &ca)
	if err != nil || ca.ErrorStatus != "Success" {
		log.Print(err)
		return
	}

	newActivity := richgo.Activity{
		LargeImage: "destinylogo",
		Details: "Launching the game...",
	}

	var activityModeHash int32
	var dateActivityStarted time.Time

	isLaunching := true
	for id, d := range ca.Response.CharacterActivities.Data {
		if d.CurrentActivityHash != 0 {
			if t, err := time.Parse(time.RFC3339, d.DateActivityStarted); err == nil {
				if t.Unix() > dateActivityStarted.Unix() {
					dateActivityStarted = t
					newActivity = richgo.Activity{
						LargeImage: "destinylogo",
					}
				} else {
					break;
				}
			}

			if isLaunching {
				isLaunching = false
			}

			var (
				fetchedCurrentActivity *currentActivityDefinition
				fetchedCurrentActivityMode *currentActivityModeDefinition
			)
			activityHash, err := getHashFromTable("DestinyActivityDefinition", d.CurrentActivityHash, &fetchedCurrentActivity)
			activityModeHash, err = getHashFromTable("DestinyActivityModeDefinition", d.CurrentActivityModeHash, &fetchedCurrentActivityMode)
			if err != nil { // Error indicates orbit. Seems to have been working reliably.
				newActivity.Details = "In orbit"
				newActivity.LargeImage = "destinylogo"
			} else {
				var (
					fetchedPlace *placeDefinition
				)
				placeHash, err := getHashFromTable("DestinyPlaceDefinition", fetchedCurrentActivity.PlaceHash, &fetchedPlace)
				if err == nil {
					if forge, ok := forgeHashMap[activityHash]; ok { // Forges are seen as 'Story - Earth | Forge Ignition'. Fixing that in here by making them 'Forge Ignition - Earth | FORGENAME Forge'
						newActivity.Details = fmt.Sprintf("%s - %s", fetchedCurrentActivity.DisplayProperties.Name, fetchedPlace.DisplayProperties.Name)
						newActivity.State = fmt.Sprintf("%s Forge", forge)
						newActivity.LargeImage = "forge"
					} else if placeHash == 2096719558 { // ... 'Normal Strikes - The Menagerie | The Menagerie'. Still unsure why it thinks it's a strike.
						newActivity.Details = "The Menagerie - Nessus Orbit"
						newActivity.LargeImage = "menagerie"
					} else {
						newActivity.Details = fmt.Sprintf("%s - %s", fetchedCurrentActivityMode.DisplayProperties.Name, fetchedPlace.DisplayProperties.Name)
						newActivity.State = fetchedCurrentActivity.DisplayProperties.Name
					}
				}
			}

			class := classImageMap[ca.Response.Characters.Data[id].ClassType]
			newActivity.SmallImage = class
			newActivity.SmallText = strings.Title(fmt.Sprintf("%s - %d", class, ca.Response.Characters.Data[id].Light))
		}
	}
	if isLaunching {
		setActivity(newActivity, time.Now(), 0)
	} else {
		setActivity(newActivity, dateActivityStarted, activityModeHash)
	}
}

func getHashFromTable(table string, hash int64, v interface{}) (newHash int32, err error) {
	u := uint32(hash)
	newHash = int32(u)
	var d string
	err = manifest.QueryRow(fmt.Sprintf("SELECT json FROM %s WHERE id=$2", table), newHash).Scan(&d)
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(d), &v)
	return
}

// setActivity sets the rich presence status. If there is no specific st (start time), pass an empty string.
func setActivity(newActivity richgo.Activity, st time.Time, activityModeHash int32) {
	if previousActivity.Details != newActivity.Details || previousActivity.State != newActivity.State || previousActivity.SmallText != newActivity.SmallText {
		previousActivity = newActivity

		if st.IsZero() {
			st = time.Now()
		}
		newActivity.Timestamps = &richgo.Timestamps{
			Start: &st,
		}
		newActivity.LargeText = "rich-destiny"

		if activityModeHash != 0 && newActivity.LargeImage == "destinylogo" {
			for image, hashes := range largeImageMap {
				for _, h := range hashes {
					if activityModeHash == h {
						newActivity.LargeImage = image
						break
					}
				}
				if newActivity.LargeImage != "destinylogo" {
					break
				}
			}
		}

		err := richgo.SetActivity(newActivity)
		if err != nil {
			log.Print("Error setting activity: " + err.Error())
		}
		log.Printf("%s | %s | %s", newActivity.Details, newActivity.State, newActivity.SmallText)
	}
}