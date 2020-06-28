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

var quitExeCheckTicker chan(struct{})

func initPresence() {
	exeCheckTicker := time.NewTicker(15 * time.Second)
	quitExeCheckTicker = make(chan struct{})
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
					loggedIn = false
					log.Print("No longer playing, logged ipc out")
				}
			case <- quitExeCheckTicker:
				exeCheckTicker.Stop()
			}
		}
	}()
}

var previousActivity richgo.Activity
var previousGuardian guardianIcon

func updatePresence() {
	var ca *characterActivitiesDefinition
	err := requestComponents(fmt.Sprintf("/Destiny2/3/Profile/%s?components=204,200", auth.ActualMSID), &ca)
	if err != nil || ca.ErrorStatus != "Success" {
		log.Print(err)
		return
	}

	isLaunching := true
	for id, d := range ca.Response.CharacterActivities.Data {
		if d.CurrentActivityHash != 0 {
			if isLaunching {
				isLaunching = false
			}

			newActivity := richgo.Activity{
				LargeImage: "destinylogo",
				Details: "Launching the game...",
			}

			var (
				fetchedCurrentActivity *currentActivityDefinition
				fetchedCurrentActivityMode *currentActivityModeDefinition
			)
			activityHash, err := getHashFromTable("DestinyActivityDefinition", d.CurrentActivityHash, &fetchedCurrentActivity)
			activityModeHash, err := getHashFromTable("DestinyActivityModeDefinition", d.CurrentActivityModeHash, &fetchedCurrentActivityMode)
			if err != nil { // Error indicates orbit. Seems to have been working reliably.
				newActivity.Details = "In orbit"
				newActivity.LargeImage = "destinylogo"
			} else {
				var (
					fetchedPlace *placeDefinition
				)
				_, err = getHashFromTable("DestinyPlaceDefinition", fetchedCurrentActivity.PlaceHash, &fetchedPlace)
				if err == nil {
					if forge, ok := forgeHashMap[activityHash]; ok { // Forges are seen as 'Story - Earth | Forge Ignition'. Fixing that in here by making them 'Forge Ignition - Earth | FORGENAME Forge'
						newActivity.Details = fmt.Sprintf("%s - %s", fetchedCurrentActivity.DisplayProperties.Name, fetchedPlace.DisplayProperties.Name)
						newActivity.State = fmt.Sprintf("%s Forge", forge)
						newActivity.LargeImage = "forge"
					} else {
						newActivity.Details = fmt.Sprintf("%s - %s", fetchedCurrentActivityMode.DisplayProperties.Name, fetchedPlace.DisplayProperties.Name)
						newActivity.State = fetchedCurrentActivity.DisplayProperties.Name
					}
				}
			}

			class := classImageMap[ca.Response.Characters.Data[id].ClassType]
			newGuardian := guardianIcon{
				Class: class,
				DisplayText: strings.Title(fmt.Sprintf("%s - %d", class, ca.Response.Characters.Data[id].Light)),
			}

			setActivity(newActivity, newGuardian, d.DateActivityStarted, activityModeHash)
		}
	}
	if isLaunching {
		setActivity(richgo.Activity{
			LargeImage: "destinylogo",
			Details: "Launching the game",
		}, guardianIcon{}, "", 0)
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
func setActivity(newActivity richgo.Activity, newGuardian guardianIcon, st string, activityModeHash int32) {
	if previousActivity.Details != newActivity.Details || previousActivity.State != newActivity.State || previousGuardian.DisplayText != newGuardian.DisplayText {
		previousActivity = newActivity
		previousGuardian = newGuardian
		var startTime time.Time
		if t, err := time.Parse(time.RFC3339, st); err == nil {
			startTime = t
		} else {
			startTime = time.Now()
		}
		newActivity.Timestamps = &richgo.Timestamps{
			Start: &startTime,
		}
		newActivity.LargeText = "rich destiny"

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

		if newGuardian.DisplayText != "" {
			newActivity.SmallImage = newGuardian.Class
			newActivity.SmallText = newGuardian.DisplayText
		}

		err := richgo.SetActivity(newActivity)
		marshalled, _ := json.Marshal(newActivity)
		log.Print(string(marshalled))
		if err != nil {
			log.Print("Error setting activity: " + err.Error())
		}
		log.Print(newActivity.Details + " | " + newActivity.State)
	}
}