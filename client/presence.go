package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"time"

	richgo "github.com/lieuweberg/rich-go/client"
	"github.com/mitchellh/go-ps"
)

func initPresence() {
	exeCheckTicker := time.NewTicker(100 * time.Millisecond)
	quitPresenceTicker = make(chan bool)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Println("PANIC!\n", err)
				log.Print(string(debug.Stack()))
			}
			quitPresenceTicker = nil
		}()

		firstTime := true
		loggedIn := false
		definitionsExist := false

		for {
			select {
			case <-exeCheckTicker.C:
				exeFound := false
				pl, _ := ps.Processes()
				for _, p := range pl {
					if p.Executable() == "destiny2.exe" {
						exeFound = true

						// We require the login every time we are logged out
						if !loggedIn {
							err := richgo.Login("726090012877258762")
							if err != nil {
								logErrorIfNoErrorSpam(errorOriginDiscord, fmt.Sprintf("Couldn't connect to Discord: %s", err.Error()))
								break
							}
							loggedIn = true
							resolveErrorSpam(errorOriginDiscord)

							if storage != nil {
								if storage.AutoUpdate {
									go func() {
										_, err := attemptApplicationUpdate()
										if err != nil {
											log.Printf("Error trying to update: %s", err)
										}
									}()
								}

								if storage.ReAuthAt != 0 && time.Now().Unix() >= storage.ReAuthAt {
									openTab("https://richdestiny.app/auth-expired")
								}
							}
						}

						// We require definitions once per login
						if !definitionsExist {
							err := getDefinitions()
							if err != nil {
								setMaintenance()
								logErrorIfNoErrorSpam(errorOriginDefinitions, fmt.Sprintf("Failed to get manifest: %s", err))
								break
							}
							definitionsExist = true
							resolveErrorSpam(errorOriginDefinitions)
						}

						// We require storage every iteration
						_, err := getStorage()
						if err != nil {
							setMaintenance()
							logErrorIfNoErrorSpam(errorOriginAuth, fmt.Sprintf("Error getting storage: %s", err))
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

					err := steamAPIShutdown()
					if err != nil {
						log.Println(err)
					} else {
						log.Print("Steam API logged out")
					}

					previousActivity = richgo.Activity{}
				}
			case <-quitPresenceTicker:
				exeCheckTicker.Stop()
			}

			if firstTime {
				firstTime = false
				exeCheckTicker.Stop()
				exeCheckTicker.Reset(15 * time.Second)
			}
		}
	}()
}

func updatePresence() {
	if veryImportantStatusActive {
		forcePresenceUpdate = true
		setActivity(previousActivity, *previousActivity.Timestamps.Start, nil)
		return
	}

	var profile *profileDef
	err := requestComponents(fmt.Sprintf("/Destiny2/%d/Profile/%s/?components=204,200", storage.MSType, storage.ActualMSID), &profile)
	if err != nil {
		logErrorIfNoErrorSpam(errorOriginProfileRequest, fmt.Sprintf("Error requesting profile: %s", err))
		return
	}
	if profile.ErrorStatus != "Success" {
		if profile.ErrorStatus == "SystemDisabled" || profile.ErrorStatus == "DestinyThrottledByGameServer" {
			setMaintenance()
		} else {
			logErrorIfNoErrorSpam(errorOriginProfileRequest, fmt.Sprintf("Bungie returned an error status %s when trying to get profile, message: %s", profile.ErrorStatus, profile.Message))
		}
		return
	}

	resolveErrorSpam(errorOriginAuth)
	resolveErrorSpam(errorOriginProfileRequest)

	newActivity := richgo.Activity{
		LargeImage: "destinylogo",
	}

	var dateActivityStarted time.Time
	var characterID string
	var c profileDefCharacter

	for id, d := range profile.Response.CharacterActivities.Data {
		if d.CurrentActivityHash != 0 {
			if t, err := time.Parse(time.RFC3339, d.DateActivityStarted); err == nil {
				if t.Unix() > dateActivityStarted.Unix() {
					dateActivityStarted = t
					characterID = id
					c = d
					newActivity = richgo.Activity{
						LargeImage: "destinylogo",
					}
				}
			}
		}
	}

	if characterID != "" {
		var (
			activity         *activityDefinition
			place            *placeDefinition
			activityMode     *activityModeDefinition
			activityModeHash int32
		)

		activityHash, err := getFromTableByHash("DestinyActivityDefinition", c.CurrentActivityHash, &activity)
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

			activityModeHash, err = getFromTableByHash("DestinyActivityModeDefinition", c.CurrentActivityModeHash, &activityMode)

			debugText = fmt.Sprintf("A%d, M%d, P%d", activityHash, activityModeHash, placeHash)

			if activityMode == nil {
				if err != nil {
					log.Printf("Error getting activityMode %d from definitions: %s", activityModeHash, err)
				}

				activityTypeHash, err := getFromTableByHash("DestinyActivityTypeDefinition", activity.ActivityTypeHash, &activityMode)
				if activityMode == nil {
					if err != nil {
						log.Printf("Error getting activityType %d from definitions: %s", activityTypeHash, err)
					}
				}

				debugText += fmt.Sprintf(", T%d", activityTypeHash)
			}

			transformActivity(characterID, activityHash, activityModeHash, activity, activityMode, place, &newActivity)
		}

		characterInfo := profile.Response.Characters.Data[characterID]
		class := classImages[characterInfo.ClassType]
		newActivity.SmallImage = class
		newActivity.SmallText = strings.Title(fmt.Sprintf("%s - %d", class, characterInfo.Light))

		setActivity(newActivity, dateActivityStarted, activityMode)
		return
	}

	newActivity.Details = "Launching the game..."
	setActivity(newActivity, time.Now(), nil)
}

// getFromTableByHash retrieves an object from the database by hash. ErrNoRows is not returned.
func getFromTableByHash(table string, hash uint32, v interface{}) (newHash int32, err error) {
	newHash = int32(hash)
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
	case "Neptune":
		place.DP.Name = "Neomuna"
	case "Titan, Moon of Saturn":
		place.DP.Name = "Titan"
	}
}

// transformActivity changes the activity based on a set of prewritten overrides in case activities are badly represented
func transformActivity(charID string, activityHash, activityModeHash int32, activity *activityDefinition, activityMode *activityModeDefinition, place *placeDefinition, newActivity *richgo.Activity) {
	// We're gonna have to rely only on activity. https://github.com/Bungie-net/api/issues/910, scroll down for all my comments and edits
	if activityMode == nil || activityMode.DP.Name == "" {
		switch {
		default:
			newActivity.Details = "In Orbit"
			if storage.OrbitText != "" {
				newActivity.State = storage.OrbitText
			}
		}
	} else {
		// This part is for things that are incorrectly/unpleasantly formatted.
		switch {
		case activity.DP.Name == "H.E.L.M." || activity.DP.Name == "The Farm":
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
		case strings.HasPrefix(activity.DP.Name, "The Coil"):
			newActivity.Details = "The Coil" + " - The Dreaming City"
			newActivity.LargeImage = "seasonwish"
		case strings.HasPrefix(activity.DP.Name, "Savathûn's Spire"):
			newActivity.Details = "Savathûn's Spire - " + place.DP.Name
			if strings.Contains(activity.DP.Name, "Legend") {
				newActivity.State = "Difficulty: Legend"
			}
			newActivity.LargeImage = "seasonwitch"
		case activity.DP.Name == "Altars of Summoning":
			newActivity.Details = activity.DP.Name + " - " + place.DP.Name
			newActivity.LargeImage = "seasonwitch"
		case strings.Contains(activity.DP.Name, "Salvage"):
			newActivity.Details = activity.DP.Name + " - Titan"
			s := strings.SplitN(activity.DP.Name, ": ", 2)
			if len(s) == 2 {
				newActivity.State = "Difficulty: " + s[1]
			}
			newActivity.LargeImage = "seasondeep"
		case strings.HasPrefix(activity.DP.Name, "Deep Dives"):
			newActivity.Details = "Deep Dive - " + place.DP.Name
			if strings.Contains(activity.DP.Name, "Private") {
				newActivity.State = "Private"
			}
			newActivity.LargeImage = "seasondeep"
		case strings.Contains(activity.DP.Name, "Defiant Battleground"):
			if strings.Contains(activity.DP.Name, "Orbital Prison") {
				newActivity.Details = "Defiant Battleground - Orbital Prison"
			} else {
				newActivity.Details = "Defiant Battleground - " + place.DP.Name
			}
			if strings.Contains(activity.DP.Name, "Legend") {
				newActivity.State = "Difficulty: Legend"
			}
			newActivity.LargeImage = "seasondefiance"
		// case strings.HasPrefix(activity.DP.Name, "Ketchcrash"):
		// 	newActivity.Details = "Ketchcrash - " + place.DP.Name
		// 	s := strings.SplitN(activity.DP.Name, ": ", 2)
		// 	if len(s) == 2 {
		// 		newActivity.State = "Difficulty: " + s[1]
		// 	}
		// 	newActivity.LargeImage = "seasonplunder"
		// case strings.HasPrefix(activity.DP.Name, "Expedition"):
		// 	newActivity.Details = "Expedition - " + place.DP.Name
		// 	newActivity.LargeImage = "seasonplunder"
		// case strings.HasPrefix(activity.DP.Name, "Sever - "):
		// 	newActivity.Details = "Sever - " + place.DP.Name
		// 	newActivity.State = strings.SplitN(activity.DP.Name, " - ", 2)[1]
		// 	newActivity.LargeImage = "seasonhaunted"
		case strings.HasPrefix(activity.DP.Name, "The Wellspring:"):
			newActivity.Details = "The Wellspring - " + place.DP.Name
			newActivity.State = strings.SplitN(activity.DP.Name, ": ", 2)[1]
			newActivity.LargeImage = "wellspring"
		case strings.Contains(activity.DP.Name, "PsiOps Battleground"):
			s := strings.Split(activity.DP.Name, ": ")
			newActivity.Details = s[0]
			newActivity.State = s[1]
			newActivity.LargeImage = "seasonrisen"
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
		case activity.DP.Name == "Prophecy":
			// Change from Earth to IX Realms
			newActivity.Details = "Dungeon - IX Realms"
			newActivity.State = activity.DP.Name
		case strings.HasPrefix(activity.DP.Name, "Grasp of Avarice"):
			// Change from Earth to The Cosmodrome
			newActivity.Details = "Dungeon - The Cosmodrome"
			newActivity.State = activity.DP.Name
		case strings.HasPrefix(activity.DP.Name, "Last Wish"):
			// Remove Level: XX from the state
			newActivity.Details = "Raid - The Dreaming City"
			newActivity.State = "Last Wish"
		case strings.HasPrefix(activity.DP.Name, "Garden of Salvation"):
			// Change from Moon to Black Garden
			newActivity.Details = "Raid - Black Garden"
			newActivity.State = activity.DP.Name
		case activity.ActivityTypeHash == 332181804:
			// Story - The Moon | Nightmare Hunt: name: difficulty
			newActivity.Details = "Nightmare Hunt - " + place.DP.Name
			newActivity.State = strings.SplitN(activity.DP.Name, ":", 2)[1]
			newActivity.LargeImage = "shadowkeep"
		// case activity.DP.Name == "Last City: Eliksni Quarter":
		// 	newActivity.Details = "Eliksni Quarter - The Last City"
		// 	newActivity.LargeImage = "storypvecoopheroic"
		// Keep this case at the very bottom
		case activityMode.DP.Name == "Story":
			newActivity.Details = "Story - " + place.DP.Name
			newActivity.State = activity.DP.Name
			for campaign, missions := range storyMissions {
				for _, m := range missions {
					if strings.HasPrefix(activity.DP.Name, m) {
						if campaign == "seasonwish" {
							newActivity.Details = "Riven's Lair - The Dreaming City"
						}
						newActivity.LargeImage = campaign
						return
					}
				}
			}
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
				if strings.Contains(activity.DP.Name, "Nightfall Grandmaster") {
					newActivity.Details = "Grandmaster Nightfall - " + place.DP.Name
					newActivity.State = strings.SplitN(activity.DP.Name, ": ", 2)[1]
				} else {
					newActivity.Details = "Nightfall - " + place.DP.Name
					a := strings.Split(activity.DP.Name, ": ")
					newActivity.State = "Difficulty: " + a[len(a)-1]
				}
			} else {
				newActivity.Details = activityMode.DP.Name + " - " + place.DP.Name
				newActivity.State = activity.DP.Name
			}
		}

		if activityMode.DP.Name == "Raid" || activityMode.DP.Name == "Dungeon" {
			split := strings.SplitN(activity.DP.Name, ": ", 2)
			if len(split) > 1 && split[1] == "Normal" {
				newActivity.State = split[0]
			}

			if activityMode.DP.Name == "Raid" {
				getActivityPhases(charID, split[0], activityHash, newActivity)
			}
		}
	}
}

func getActivityPhases(charID, phasesMapKey string, activityHash int32, newActivity *richgo.Activity) {
	if _, ok := raidProgressionMap[phasesMapKey]; !ok {
		return
	}

	var p progressions
	err := requestComponents(fmt.Sprintf("/Destiny2/%d/Profile/%s/Character/%s?components=202", storage.MSType, storage.ActualMSID, charID), &p)
	if err != nil {
		logErrorIfNoErrorSpam(errorOriginActivityPhases, fmt.Sprintf("Error requesting activity phases: %s", err))
		return
	}
	if p.ErrorStatus != "Success" {
		logErrorIfNoErrorSpam(errorOriginActivityPhases, fmt.Sprintf("Bungie returned an error status %s when trying to get activity phases, message: %s", p.ErrorStatus, p.Message))
		return
	}

	for _, m := range p.Response.Progressions.Data.Milestones {
		for _, a := range m.Activities {
			if a.ActivityHash == uint32(activityHash) {
				if a.Phases != nil {
					for i, phase := range *a.Phases {
						if !phase.Complete {
							newActivity.Details = fmt.Sprintf("%s - %s", newActivity.State, strings.SplitN(newActivity.Details, " - ", 2)[1])
							newActivity.State = fmt.Sprintf("%s (%d/%d)", raidProgressionMap[phasesMapKey][i], i+1, len(raidProgressionMap[phasesMapKey]))
							return
						}
					}
				}
			}
		}
	}
}

// setActivity sets the rich presence status
func setActivity(newActivity richgo.Activity, st time.Time, activityMode *activityModeDefinition) {
	if st.IsZero() {
		st = time.Now()
	}
	newActivity.Timestamps = &richgo.Timestamps{
		Start: &st,
	}
	newActivity.LargeText = "richdestiny.app " + version

	if activityMode != nil && newActivity.LargeImage == "destinylogo" {
		newActivity.LargeImage = getLargeImage(activityMode.DP.Name)
	}

	if !veryImportantStatusActive && storage != nil && storage.JoinGameButton {
		if !storage.JoinOnlySocial || (newActivity.LargeImage == "socialall" || newActivity.Details == "In Orbit") {
			joinLink, err := getJoinLink()
			if err != nil {
				if !errors.Is(err, errNoConnectString) {
					logErrorIfNoErrorSpam(errorOriginSteam, "Unknown error trying to get connection string")
				}

				newActivity.Buttons = []*richgo.Button{
					{
						Label: "Launch Game",
						Url:   "steam://run/1085660/",
					},
				}
			} else {
				newActivity.Buttons = []*richgo.Button{
					{
						Label: "Join Game",
						Url:   joinLink,
					},
				}

				if len(previousActivity.Buttons) > 0 && previousActivity.Buttons[0].Url != joinLink {
					log.Printf("New join link: %s", joinLink)
					forcePresenceUpdate = true
				}
			}
		}
	}

	// Condition that decides whether to update the presence or not
	if previousActivity.Details != newActivity.Details ||
		previousActivity.State != newActivity.State ||
		previousActivity.SmallText != newActivity.SmallText ||
		forcePresenceUpdate {
		if forcePresenceUpdate {
			forcePresenceUpdate = false
		}

		previousActivity = newActivity
		err := richgo.SetActivity(newActivity)
		if err != nil {
			log.Print("Error setting activity: " + err.Error())
		}

		if !veryImportantStatusActive {
			log.Printf("%s | %s | %s", newActivity.Details, newActivity.State, newActivity.SmallText)
		}
	}

}

func getLargeImage(name string) string {
	if strings.HasPrefix(name, "Private Matches") {
		return "privatecrucible"
	}

	if strings.HasPrefix(name, "Iron Banner") {
		return "ironbanner"
	}

	condensedName := strings.ToLower(strings.ReplaceAll(name, " ", ""))
	var isOneInherently bool
	for image, modes := range largeImageActivityModes {
		if condensedName == image {
			isOneInherently = true
		}

		for _, m := range modes {
			if m == name {
				return image
			}
		}
	}

	if isOneInherently {
		return condensedName
	}

	return "destinylogo"
}

func setMaintenance() {
	setActivity(richgo.Activity{
		LargeImage: "destinylogo",
		Details:    "Waiting for maintenance to end",
	}, time.Now(), nil)
}
