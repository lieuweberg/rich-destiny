package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	richgo "github.com/lieuweberg/rich-go/client"
	"github.com/mitchellh/go-ps"
)

func initPresence() {
	exeCheckTicker := time.NewTicker(15 * time.Second)
	quitPresenceTicker = make(chan bool)
	loggedIn := false
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Print("PANIC!\n", err)
			}
		}()

		for {
			select {
			case <-exeCheckTicker.C:
				exeFound := false
				pl, _ := ps.Processes()
				for _, p := range pl {
					if p.Executable() == "destiny2.exe" {
						exeFound = true

						getStorage()
						if storage == nil {
							break
						}

						if !loggedIn {
							err := richgo.Login("726090012877258762")
							if err != nil {
								log.Print("Couldn't connect to Discord: " + err.Error())
								break
							}
							loggedIn = true

							getDefinitions()

							if storage.AutoUpdate {
								go func() {
									// This only runs once when the game has been started, I don't really care whether it displays an error then
									// even though it's not really an error at all.
									_, err := attemptApplicationUpdate()
									if err != nil {
										log.Printf("Error trying to update: %s", err)
									}
								}()
							}
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
			case <-quitPresenceTicker:
				exeCheckTicker.Stop()
			}
		}
	}()
}

func updatePresence() {
	var ca *profileDef
	err := requestComponents(fmt.Sprintf("/Destiny2/%d/Profile/%s/?components=204,200", storage.MSType, storage.ActualMSID), &ca)
	if err != nil || ca.ErrorStatus != "Success" {
		if err == nil {
			if ca.ErrorStatus == "SystemDisabled" {
				setActivity(richgo.Activity{
					LargeImage: "destinylogo",
					Details:    "Waiting for maintenance to end",
				}, time.Now(), 0)
				return
			}
			log.Println(ca.ErrorStatus, ca.Message)
		} else {
			log.Print(err)
		}
		return
	}

	newActivity := richgo.Activity{
		LargeImage: "destinylogo",
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
					continue
				}
			}

			if isLaunching {
				isLaunching = false
			}

			var (
				activity     *activityDefinition
				place        *placeDefinition
				activityMode *activityModeDefinition
			)

			activityHash, err := getFromTableByHash("DestinyActivityDefinition", d.CurrentActivityHash, &activity)
			// Something is terribly wrong :(
			if activity == nil {
				if err != nil {
					log.Printf("Error getting activity %d from definitions: %s", activityHash, err)
				}

				newActivity.Details = "???"
			} else {
				placeHash, err := getFromTableByHash("DestinyPlaceDefinition", activity.PlaceHash, &place)
				if place == nil {
					if err != nil {
						log.Printf("Error getting place %d from definitions: %s", placeHash, err)
					}

					place = &placeDefinition{
						DP: globalDisplayProperties{
							Name: "???",
						},
					}
				} else {
					transformPlace(place, activity)
				}

				activityModeHash, err = getFromTableByHash("DestinyActivityModeDefinition", d.CurrentActivityModeHash, &activityMode)

				if activityMode == nil {
					if err != nil {
						log.Printf("Error getting activityMode %d from definitions: %s", activityModeHash, err)
					}

					debugText = fmt.Sprintf("%d, %d", activityHash, activityModeHash)
				} else {
					debugText = fmt.Sprintf("%d, %d, %d", activityHash, activityModeHash, placeHash)
				}

				transformActivity(id, activityHash, activityModeHash, activity, activityMode, place, &newActivity)
			}

			class := classImages[ca.Response.Characters.Data[id].ClassType]
			newActivity.SmallImage = class
			newActivity.SmallText = strings.Title(fmt.Sprintf("%s - %d", class, ca.Response.Characters.Data[id].Light))
			break
		}
	}

	// This is outside of the loop. If no characters have a current activity other than 0, it indicates the game is launching
	if isLaunching {
		newActivity.Details = "Launching the game..."
		setActivity(newActivity, time.Now(), 0)
	} else {
		setActivity(newActivity, dateActivityStarted, activityModeHash)
	}
}

// getFromTableByHash retrieves an object from the database by hash. ErrNoRows is not returned.
func getFromTableByHash(table string, hash int64, v interface{}) (newHash int32, err error) {
	u := uint32(hash)
	newHash = int32(u)
	var d string
	err = manifest.QueryRow(fmt.Sprintf("SELECT json FROM %s WHERE id=$1", table), newHash).Scan(&d)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return newHash, nil
		}
		return
	}
	err = json.Unmarshal([]byte(d), &v)
	return
}

// transformPlace changes some funny or vague place names to where you actually are
func transformPlace(place *placeDefinition, activity *activityDefinition) {
	switch place.DP.Name {
	case "Earth":
		if activity.DestinationHash == 2073151843 || activity.DestinationHash == 3990611421 {
			place.DP.Name = "The Cosmodrome"
		} else if activity.DestinationHash == 697502628 || activity.DestinationHash == 1199524104 {
			place.DP.Name = "EDZ"
		}
	case "Rathmore Chaos, Europa":
		place.DP.Name = "Europa"
	case "Court of Savathûn, Throne World":
		place.DP.Name = "Savathûn's Throne World"
	}
}

// transformActivity changes the activity based on a set of prewritten overrides in case activities are badly represented
func transformActivity(charID string, activityHash, activityModeHash int32, activity *activityDefinition, activityMode *activityModeDefinition, place *placeDefinition, newActivity *richgo.Activity) {
	// We're gonna have to rely only on activity. https://github.com/Bungie-net/api/issues/910, scroll down for all my comments and edits
	if activityMode == nil {
		switch {
		case strings.HasPrefix(activity.DP.Name, "Grasp of Avarice"):
			newActivity.Details = "Dungeon - The Cosmodrome"
			newActivity.State = activity.DP.Name
			newActivity.LargeImage = "dungeon"
		// case strings.HasPrefix(activity.DP.Name, "Astral Alignment"):
		// 	newActivity.Details = "Astral Alignment - " + place.DP.Name
		// 	newActivity.State = "Difficulty: " + strings.Split(activity.DP.Name, ": ")[1]
		// 	newActivity.LargeImage = "seasonlost"
		// case strings.HasPrefix(activity.DP.Name, "Expunge:"):
		// 	newActivity.Details = "Expunge - " + place.DP.Name
		// 	newActivity.State = strings.Split(activity.DP.Name, ": ")[1]
		// 	newActivity.LargeImage = "seasonsplicer"
		// case strings.HasPrefix(activity.DP.Name, "Override:"):
		// 	newActivity.Details = strings.Replace(activity.DP.Name, ": ", " - ", 1)
		// 	newActivity.LargeImage = "seasonsplicer"
		case strings.HasPrefix(activity.DP.Name, "Battleground:"):
			newActivity.Details = "Battleground - " + place.DP.Name
			name := strings.Split(activity.DP.Name, ": ")[1]
			newActivity.State = name
			for _, n := range chosenBattlegrounds {
				if name == n {
					newActivity.LargeImage = "seasonchosen"
					return
				}
			}
			newActivity.LargeImage = "seasonrisen"
		case strings.HasPrefix(activity.DP.Name, "Vault of Glass"):
			newActivity.Details = "Raid - Venus"
			newActivity.State = activity.DP.Name
			newActivity.LargeImage = "raid"
			getActivityPhases(charID, "vog", activityHash, newActivity)
		case activity.DP.Name == "Deep Stone Crypt":
			newActivity.Details = "Raid - Europa"
			newActivity.State = activity.DP.Name
			newActivity.LargeImage = "raid"
			getActivityPhases(charID, "dsc", activityHash, newActivity)
		case activity.DP.Name == "Prophecy":
			newActivity.Details = "Dungeon - IX Realms"
			newActivity.State = activity.DP.Name
			newActivity.LargeImage = "dungeon"
		case activity.DP.Name == "Garden of Salvation":
			newActivity.Details = "Raid - Black Garden"
			newActivity.State = activity.DP.Name
			newActivity.LargeImage = "raid"
			// getActivityPhases(id, "gos", activityHash, &newActivity)
		default:
			newActivity.Details = "In Orbit"
			if storage.OrbitText != "" {
				newActivity.State = storage.OrbitText
			}
		}
	} else {
		// This part is for things that are incorrectly/unpleasantly formatted.
		switch {
		case activity.DP.Name == "H.E.L.M.":
			// Explore - EDZ
			newActivity.Details = "Social - Earth"
			newActivity.State = activity.DP.Name
			newActivity.LargeImage = "socialall"
		// case activity.DP.Name == "European Aerial Zone":
		// 	newActivity.Details = activity.DP.Name
		case activityMode.DP.Name == "Explore":
			// Remove double place
			newActivity.Details = "Explore - " + place.DP.Name
			if strings.Contains(strings.ToLower(activity.DP.Name), "mission") {
				newActivity.State = activity.DP.Name
			}
			// Unknown Space is Eternity
			if place.DP.Name == "Unknown Space" {
				newActivity.Details = "Traversing Eternity"
				newActivity.LargeImage = "anniversary"
			}
		case activityMode.DP.Name == "Story":
			newActivity.Details = "Story - " + place.DP.Name
			newActivity.State = activity.DP.Name
			for campaign, missions := range storyMissions {
				for _, m := range missions {
					if strings.HasPrefix(activity.DP.Name, m) {
						newActivity.LargeImage = campaign
						return
					}
				}
			}
		case activityMode.DP.Name == "Dares of Eternity":
			newActivity.Details = activityMode.DP.Name
			newActivity.State = "Difficulty: " + strings.Split(activity.DP.Name, ": ")[1]
			newActivity.LargeImage = "anniversary"
		case activity.DP.Name == "Haunted Sectors":
			newActivity.Details = "Haunted Sector"
			newActivity.State = place.DP.Name
			newActivity.LargeImage = "hauntedforest"
		// case strings.HasPrefix(activity.DP.Name, "Shattered Realm"):
		// 	newActivity.Details = "Shattered Realm - The Dreaming City"
		// 	newActivity.State = strings.SplitN(activity.DP.Name, ":", 2)[1]
		// 	newActivity.LargeImage = "seasonlost"
		case activityMode.DP.Name == "Gambit":
			newActivity.Details = activityMode.DP.Name
			newActivity.State = activity.DP.Name
		// case activity.ActivityTypeHash == 400075666:
		// 	if activityHash == -1785427429 || activityHash == -1785427432 || activityHash == -1785427431 {
		// 		// 'The Menagerie - The Menagerie | The Menagerie: The Menagerie (Heroic)' Instead of thinking of strikes, it overly formats
		// 		newActivity.Details = "The Menagerie (Heroic)"
		// 	} else {
		// 		// 'Normal Strikes - The Menagerie | The Menagerie'. Still unsure why it thinks it's a strike. There are about 20 activities for
		// 		// The Menagerie, so if it's not one of the heroic ones, assume it's regular.
		// 		newActivity.Details = "The Menagerie"
		// 	}
		// 	newActivity.LargeImage = "menagerie"
		case activity.DP.Name == "The Shattered Throne":
			// Story - The Dreaming City | The Shattered Throne
			newActivity.Details = "Dungeon - The Dreaming City"
			newActivity.State = activity.DP.Name
			newActivity.LargeImage = "dungeon"
		case activityMode.DP.Name == "Raid" && place.DP.Name == "The Dreaming City":
			// Remove Level: XX from the state
			newActivity.Details = "Raid - The Dreaming City"
			newActivity.State = "Last Wish"
			getActivityPhases(charID, "lw", activityHash, newActivity)
		case activity.ActivityTypeHash == 332181804:
			// Story - The Moon | Nightmare Hunt: name: difficulty
			newActivity.Details = "Nightmare Hunt - " + place.DP.Name
			newActivity.State = strings.SplitN(activity.DP.Name, ":", 2)[1]
			newActivity.LargeImage = "shadowkeep"
		case activity.DP.Name == "Last City: Eliksni Quarter":
			newActivity.Details = "Eliksni Quarter - The Last City"
			newActivity.LargeImage = "storypvecoopheroic"
		default:
			if activityMode.DP.Name == "Scored Nightfall Strikes" {
				// Scored lost sectors are seen as scored nightfall strikes
				for _, ls := range scoredLostSectors {
					if strings.Contains(activity.DP.Name, ls) {
						newActivity.Details = "Lost Sector - " + place.DP.Name
						newActivity.State = activity.DP.Name
						newActivity.LargeImage = "lostsector"
						return
					}
				}
				// It was not a lost sector
				newActivity.Details = "Nightfall: The Ordeal - " + place.DP.Name
				a := strings.Split(activity.DP.Name, ": ")
				newActivity.State = "Difficulty: " + a[len(a)-1]
			} else {
				newActivity.Details = activityMode.DP.Name + " - " + place.DP.Name
				newActivity.State = activity.DP.Name
			}
		}
	}
}

func getActivityPhases(charID, shortName string, activityHash int32, newActivity *richgo.Activity) {
	var p progressions
	err := requestComponents(fmt.Sprintf("/Destiny2/%d/Profile/%s/Character/%s?components=202", storage.MSType, storage.ActualMSID, charID), &p)
	if err != nil {
		log.Print(err)
	}
	if p.ErrorStatus != "Success" {
		log.Println(p.ErrorStatus, p.Message)
	}

	for _, m := range p.Response.Progressions.Data.Milestones {
		for _, a := range m.Activities {
			if a.ActivityHash == int64(activityHash) {
				if a.Phases != nil {
					for i, phase := range *a.Phases {
						if !phase.Complete {
							newActivity.Details = fmt.Sprintf("%s - %s", newActivity.State, strings.SplitN(newActivity.Details, " - ", 2)[1])
							newActivity.State = fmt.Sprintf("%s (%d/%d)", raidProgressionMap[shortName][i], i+1, len(raidProgressionMap[shortName]))
							return
						}
					}
				}
			}
		}
	}
}

// setActivity sets the rich presence status
func setActivity(newActivity richgo.Activity, st time.Time, activityModeHash int32) {
	// Condition that decides whether to update the presence or not
	if previousActivity.Details != newActivity.Details ||
		previousActivity.State != newActivity.State ||
		previousActivity.SmallText != newActivity.SmallText ||
		forcePresenceUpdate {

		if forcePresenceUpdate {
			forcePresenceUpdate = false
		}

		if st.IsZero() {
			st = time.Now()
		}
		newActivity.Timestamps = &richgo.Timestamps{
			Start: &st,
		}
		newActivity.LargeText = "richdestiny.app " + version

		if activityModeHash != 0 && newActivity.LargeImage == "destinylogo" {
			for image, hashes := range commonLargeImageActivityModes {
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

		if storage.JoinGameButton {
			if !storage.JoinOnlySocial || (newActivity.LargeImage == "socialall" || newActivity.Details == "In Orbit") {
				newActivity.Buttons = []*richgo.Button{
					{
						Label: "Launch Game",
						Url:   "steam://run/1085660/",
					},
				}
			}
		}

		previousActivity = newActivity
		err := richgo.SetActivity(newActivity)
		if err != nil {
			log.Print("Error setting activity: " + err.Error())
		}
		log.Printf("%s | %s | %s", newActivity.Details, newActivity.State, newActivity.SmallText)
	}
}
