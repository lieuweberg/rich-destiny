package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	richgo "github.com/hugolgst/rich-go/client"
	"github.com/mitchellh/go-ps"
)

var quitExeCheckTicker chan(struct{})

func initPresence() {
	exeCheckTicker := time.NewTicker(15 * time.Second)
	quitExeCheckTicker = make(chan struct{})
	// isPlaying := false
	go func() {
		for {
			select {
			case <- exeCheckTicker.C:
				pl, _ := ps.Processes()
				for _, p := range pl {
					if p.Executable() == "destiny2.exe" {
						// isPlaying = true
						err := richgo.Login("726090012877258762")
						if err != nil {
							log.Print("Couldn't connect to Discord: " + err.Error())
							return
						}

						getAuth()
						updatePresence()
					}
					// } else if isPlaying {
					// 	isPlaying = false
					// 	richgo.Logout()
					// }
				}
			case <- quitExeCheckTicker:
				exeCheckTicker.Stop()
			}
		}
	}()
}

var previousActivity richgo.Activity

func updatePresence() {
	var ca *characterActivitiesDefinition
	err := requestComponents(fmt.Sprintf("/Destiny2/3/Profile/%s?components=1000,204", auth.ActualMSID), &ca)
	if err != nil || ca.ErrorStatus != "Success" {
		log.Print(err)
		return
	}

	isLaunching := true
	for _, d := range ca.Response.CharacterActivities.Data {
		if d.CurrentActivityHash != 0 {
			isLaunching = false
			var (
				fetchedCurrentActivity *currentActivityDefinition
				fetchedCurrentActivityMode *currentActivityModeDefinition
			)
			err = getHashFromTable("DestinyActivityDefinition", d.CurrentActivityHash, &fetchedCurrentActivity)
			err = getHashFromTable("DestinyActivityModeDefinition", d.CurrentActivityModeHash, &fetchedCurrentActivityMode)
			if err != nil {
				log.Printf("Orbit")
				
				return
			}
			var (
				fetchedPlace *placeDefinition
			)
			err = getHashFromTable("DestinyPlaceDefinition", fetchedCurrentActivity.PlaceHash, &fetchedPlace)

			newActivity := richgo.Activity{
				LargeImage: "destinylogo",
				LargeText: "Destiny 2",
				Details: fmt.Sprintf("%s - %s", fetchedCurrentActivityMode.DisplayProperties.Name, fetchedPlace.DisplayProperties.Name),
			}
			// if previousActivity == richgo.Activity{} || previousActivity.Details != newActivity.Details || previousActivity.State != newActivity.State {
				// previousActivity = newActivity
				err = richgo.SetActivity(newActivity)
				if err != nil {
					log.Print("Error setting activity: " + err.Error())
				}
			// }

			log.Printf("%s - %s", fetchedCurrentActivityMode.DisplayProperties.Name, fetchedPlace.DisplayProperties.Name)
		}
	}
	if isLaunching {
		log.Printf("Launching the game")
	}
}

func getHashFromTable(table string, hash int64, v interface{}) (err error) {
	u := uint32(hash)
	var d string
	err = manifest.QueryRow(fmt.Sprintf("SELECT json FROM %s WHERE id=$2", table), int32(u)).Scan(&d)
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(d), &v)
	return
}