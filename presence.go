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
						if storage == nil {
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
	err := requestComponents(fmt.Sprintf("/Destiny2/%d/Profile/%s?components=204,200", storage.MSType, storage.ActualMSID), &ca)
	if err != nil || ca.ErrorStatus != "Success" {
		if err == nil {
			log.Print(ca.ErrorStatus, ca.Message)
		} else {
			log.Print(err)
		}
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
					continue;
				}
			}

			if isLaunching {
				isLaunching = false
			}

			var (
				activity *currentActivityDefinition
				activityMode *currentActivityModeDefinition
			)
			activityHash, err := getHashFromTable("DestinyActivityDefinition", d.CurrentActivityHash, &activity)
			activityModeHash, err = getHashFromTable("DestinyActivityModeDefinition", d.CurrentActivityModeHash, &activityMode)
			if err != nil { // Error indicates orbit. ~~Seems to have been working reliably.~~
				debugText = fmt.Sprintf("%d, %d", activityHash, activityModeHash)

				// Flaw in bungie api, activity mode is the "undefined" hash and thus it can't find certain modes in the manifest.
				if activityHash == -146779922 {
					newActivity.Details = "Dungeon"
					newActivity.State = "Prophecy"
					newActivity.LargeImage = "dungeon"
				} else if activityHash == -1635244228 {
					newActivity.Details = "Raid - Black Garden"
					newActivity.State = "Garden of Salvation"
					newActivity.LargeImage = "raid"
				} else {
					newActivity.Details = "In orbit"
					newActivity.LargeImage = "destinylogo"
					if storage.OrbitText != "" {
						newActivity.State = storage.OrbitText
					}
				}
				
			} else {
				var place *placeDefinition
				placeHash, err := getHashFromTable("DestinyPlaceDefinition", activity.PlaceHash, &place)

				// Here are any overrides due to strange API shenanigans.
				// This first if part should not be long, and should be used for everything that should be changed all the time if it appears (e.g. the name of a destination).
				if placeHash == 1729879943 {
					place.DP.Name = "Europa"
				}
				if placeHash == -547261341 {
					place.DP.Name = "The Cosmodrome"
				}

				// This second part specifies more specific overrides.
				switch {
				case activityModeHash == -797199657:
					// Remove double place
					newActivity.Details = "Explore - " + place.DP.Name
					if strings.Contains(strings.ToLower(activity.DP.Name), "mission") {
						newActivity.State = activity.DP.Name
					}
				case activityHash == 707826522  || activityHash == 1454880421 || activityHash == -420675050:
					newActivity.Details = activity.DP.Name
					newActivity.LargeImage = "hauntedforest"
				case activity.ActivityTypeHash == 400075666:
					if activityHash == -1785427429 || activityHash == -1785427432 || activityHash == -1785427431 {
						// 'The Menagerie - The Menagerie | The Menagerie: The Menagerie (Heroic)' Instead of thinking of strikes, it overly formats
						newActivity.Details = "The Menagerie (Heroic)"
					} else {
						// 'Normal Strikes - The Menagerie | The Menagerie'. Still unsure why it thinks it's a strike. There are about 20 activities for
						// The Menagerie, so if it's not one of the heroic ones, assume it's regular.
						newActivity.Details = "The Menagerie"
					}
					newActivity.LargeImage = "menagerie"
				case activityHash == 2032534090:
					// Story - The Dreaming City | The Shattered Throne
					newActivity.Details = "Dungeon - The Dreaming City"
					newActivity.State = "The Shattered Throne"
					newActivity.LargeImage = "dungeon"
				case activityModeHash == 2043403989 && placeHash == -1417085778:
					// Remove Level: XX from the state
					newActivity.Details = "Raid - The Dreaming City"
					newActivity.State = "Last Wish"
				default:
					// This third part specifies overrides that do not use simple conditions and can't fit in a  case  statement
					if forge, ok := forgeHashMap[activityHash]; ok {
						// Forges are seen as 'Story - Earth | Forge Ignition'. Fixing that in here by making them 'Forge Ignition - PLACE | FORGENAME Forge'
						newActivity.Details = fmt.Sprintf("%s - %s", activity.DP.Name, place.DP.Name)
						newActivity.State = fmt.Sprintf("%s Forge", forge)
						newActivity.LargeImage = "forge"
					} else {
						newActivity.Details = activityMode.DP.Name
						if err == nil {
							newActivity.Details += fmt.Sprintf(" - %s", place.DP.Name)
						}
						newActivity.State = activity.DP.Name
					}
				}

				debugText = fmt.Sprintf("%d, %d, %d", activityHash, activityModeHash, placeHash)
			}

			class := classImageMap[ca.Response.Characters.Data[id].ClassType]
			newActivity.SmallImage = class
			newActivity.SmallText = strings.Title(fmt.Sprintf("%s - %d", class, ca.Response.Characters.Data[id].Light))
		}
	}

	// This is outside of the loop. If no characters have a current activity other than 0, it indicates the game is launching
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
	err = manifest.QueryRow(fmt.Sprintf("SELECT json FROM %s WHERE id=$1", table), newHash).Scan(&d)
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