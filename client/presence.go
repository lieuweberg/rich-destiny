package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	richgo "github.com/lieuweberg/rich-go/client"
	"github.com/mitchellh/go-ps"
)

func initPresence() {
	exeCheckTicker := time.NewTicker(15 * time.Second)
	quitPresenceTicker = make(chan struct{})
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
						if !loggedIn {
							err := richgo.Login("726090012877258762")
							if err != nil {
								log.Print("Couldn't connect to Discord: " + err.Error())
								break
							}
							loggedIn = true

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

						getStorage()
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
		Details:    "Launching the game...",
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
				activityMode *currentActivityModeDefinition
			)
			activityHash, err := getFromTableByHash("DestinyActivityDefinition", d.CurrentActivityHash, &activity)
			placeHash, err := getFromTableByHash("DestinyPlaceDefinition", activity.PlaceHash, &place)
			activityModeHash, err = getFromTableByHash("DestinyActivityModeDefinition", d.CurrentActivityModeHash, &activityMode)

			if place != nil {
				transformPlace(place, activity)
			}

			if err != nil { // Error indicates orbit. ~~Seems to have been working reliably.~~
				debugText = fmt.Sprintf("%d, %d", activityHash, activityModeHash)

				// lol, https://github.com/Bungie-net/api/issues/910, scroll down for all my comments and edits
				if activity != nil {
					switch {
					case strings.Contains(activity.DP.Name, "Expunge:"):
						newActivity.Details = "Expunge - " + place.DP.Name
						newActivity.State = strings.Split(activity.DP.Name, ": ")[1]
						newActivity.LargeImage = "storypvecoopheroic"
					case strings.Contains(activity.DP.Name, "Override:"):
						newActivity.Details = strings.Replace(activity.DP.Name, ": ", " - ", 1)
						newActivity.LargeImage = "storypvecoopheroic"
					case strings.Contains(activity.DP.Name, "Battleground:"):
						newActivity.Details = "Battleground - " + place.DP.Name
						newActivity.State = strings.Split(activity.DP.Name, ": ")[1]
						newActivity.LargeImage = "storypvecoopheroic"
					case strings.Contains(activity.DP.Name, "Vault of Glass"):
						newActivity.Details = "Raid - Venus"
						newActivity.State = activity.DP.Name
						newActivity.LargeImage = "raid"
						getActivityPhases(id, "vog", activityHash, &newActivity)
					case activity.DP.Name == "Deep Stone Crypt":
						newActivity.Details = "Raid - Europa"
						newActivity.State = activity.DP.Name
						newActivity.LargeImage = "raid"
						getActivityPhases(id, "dsc", activityHash, &newActivity)
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
						newActivity.LargeImage = "destinylogo"
						if storage.OrbitText != "" {
							newActivity.State = storage.OrbitText
						}
					}
				} else {
					newActivity.Details = "???"
					newActivity.LargeImage = "destinylogo"
				}

			} else {
				// This part specifies more specific overrides.
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
				case strings.HasPrefix(activity.DP.Name, "Shattered Realm"):
					newActivity.Details = "Shattered Realm - The Dreaming City"
					newActivity.State = strings.SplitN(activity.DP.Name, ":", 2)[1]
				case activityMode.DP.Name == "Gambit":
					newActivity.Details = activityMode.DP.Name
					newActivity.State = activity.DP.Name
				// case activityHash == 707826522  || activityHash == 1454880421 || activityHash == -420675050:
				// 	newActivity.Details = activity.DP.Name
				// 	newActivity.LargeImage = "hauntedforest"
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
					getActivityPhases(id, "lw", activityHash, &newActivity)
				case activity.ActivityTypeHash == 332181804:
					// Story - The Moon | Nightmare Hunt: name: difficulty
					newActivity.Details = "Nightmare Hunt - " + place.DP.Name
					newActivity.State = strings.SplitN(activity.DP.Name, ":", 2)[1]
					newActivity.LargeImage = "nightmarehunt"
				case activity.DP.Name == "Last City: Eliksni Quarter":
					newActivity.Details = "Eliksni Quarter - The Last City"
					newActivity.LargeImage = "storypvecoopheroic"
				default:
					// This part specifies overrides that do not use simple conditions and can't fit in a case statement. Switch/case is prettier than a giant if/else imo
					// if forge, ok := forgeHashMap[activityHash]; ok {
					// 	// Forges are seen as 'Story - Earth | Forge Ignition'. Fixing that in here by making them 'Forge Ignition - PLACE | FORGENAME Forge'
					// 	newActivity.Details = fmt.Sprintf("%s - %s", activity.DP.Name, place.DP.Name)
					// 	newActivity.State = fmt.Sprintf("%s Forge", forge)
					// 	newActivity.LargeImage = "forge"
					if activityMode.DP.Name == "Scored Nightfall Strikes" {
						// Scored lost sectors are seen as scored nightfall strikes
						var didWeBreak bool
						for _, ls := range scoredLostSectors {
							if strings.Contains(activity.DP.Name, ls) {
								newActivity.Details = "Lost Sector - " + place.DP.Name
								newActivity.State = activity.DP.Name
								newActivity.LargeImage = "lostsector"
								didWeBreak = true
								break
							}
						}
						// It was not a lost sector
						if !didWeBreak {
							newActivity.Details = "Nightfall: The Ordeal - " + place.DP.Name
							a := strings.Split(activity.DP.Name, ": ")
							newActivity.State = "Difficulty: " + a[len(a)-1]
						}
					} else {
						newActivity.Details = activityMode.DP.Name
						if err == nil {
							newActivity.Details += " - " + place.DP.Name
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

// getFromTableByHash retrieves an object from the database by hash.
func getFromTableByHash(table string, hash int64, v interface{}) (newHash int32, err error) {
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

// transformPlace changes some funny or vague place names to where you actually are.
func transformPlace(place *placeDefinition, activity *activityDefinition) {
	if place.DP.Name == "Earth" {
		if activity.DestinationHash == 2073151843 || activity.DestinationHash == 3990611421 {
			place.DP.Name = "The Cosmodrome"
		} else if activity.DestinationHash == 697502628 || activity.DestinationHash == 1199524104 {
			place.DP.Name = "EDZ"
		}
	} else if place.DP.Name == "Rathmore Chaos, Europa" {
		place.DP.Name = "Europa"
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

// setActivity sets the rich presence status.
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

		if storage.JoinGameButton {
			if !storage.JoinOnlySocial || (newActivity.LargeImage == "socialall" || newActivity.Details == "In Orbit") {
				newActivity.Buttons = []*richgo.Button{
					{
						Label: "Join Game",
						Url:   fmt.Sprintf("steam://rungame/1085660/%s", storage.SteamID64),
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
