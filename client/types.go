package main

import richgo "github.com/lieuweberg/rich-go/client"

// Internal structs
// here: MSID means MembershipID
type storageStruct struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	BungieMSID       string `json:"membership_id"`

	ActualMSID string
	MSType     int32
	BungieName string
	BungieCode string
	SteamID64  string
	RefreshAt  int64
	ReAuthAt   int64

	// Settings
	OrbitText      string `json:"orbitText"`
	AutoUpdate     bool   `json:"autoUpdate"`
	Prereleases    bool   `json:"prereleases"`
	JoinGameButton bool   `json:"joinGameButton"`
	JoinOnlySocial bool   `json:"joinOnlySocial"`
}

type currentProgramStatus struct {
	Status  string `json:"status"`
	Debug   string `json:"debug"`
	Version string `json:"version"`
	Name    string `json:"name"`

	OrbitText      string `json:"orbitText"`
	AutoUpdate     bool   `json:"autoUpdate"`
	Prereleases    bool   `json:"prereleases"`
	JoinGameButton bool   `json:"joinGameButton"`
	JoinOnlySocial bool   `json:"joinOnlySocial"`

	Presence richgo.Activity `json:"presence"`
}

type releasesFromGithub []releaseElement

type releaseElement struct {
	Name       string         `json:"name"`
	Draft      bool           `json:"draft"`
	Prerelease bool           `json:"prerelease"`
	Assets     []releaseAsset `json:"assets"`
	Body       string         `json:"body"`
}

type releaseAsset struct {
	BrowserDownloadURL string `json:"browser_download_url"`
	Name               string `json:"name"`
}

type genericErrorResponse struct {
	ErrorStatus string `json:"ErrorStatus,omitempty"`
	Message     string `json:"Message,omitempty"`
}

// API/Manifest structs
// /Destiny2/Manifest
type manifestData struct {
	Response manifestResponse `json:"Response"`
	genericErrorResponse
}

type manifestResponse struct {
	MobileWorldContentPaths manifestWorldContentPaths `json:"mobileWorldContentPaths"`
}

type manifestWorldContentPaths struct {
	En string `json:"en"`
}

// /app/oauth/token
type oauthError struct {
	ErrorDescription string `json:"error_description"`
}

// /Destiny2/254/Profile/{BungieMSID}/LinkedProfiles
type linkedProfiles struct {
	Response lPResponse `json:"Response"`
	genericErrorResponse
}

type lPResponse struct {
	Profiles []lPProfile `json:"profiles"`
}

type lPProfile struct {
	MembershipTypes             []int32 `json:"applicableMembershipTypes"`
	MembershipType              int32   `json:"membershipType"`
	MembershipID                string  `json:"membershipId"`
	BungieGlobalDisplayName     string  `json:"bungieGlobalDisplayName"`
	BungieGlobalDisplayNameCode int16   `json:"bungieGlobalDisplayNameCode"`
}

// DisplayProperties used for most Manifest structs
type globalDisplayProperties struct {
	// Description string `json:"description"`
	Name string `json:"name"`
}

// /Destiny2/{MSType}/Profile/{ActualMSID}?components=204,200
type profileDef struct {
	Response profileDefReponse `json:"Response"`
	genericErrorResponse
}

type profileDefReponse struct {
	CharacterActivities profileDefActivities `json:"characterActivities"`
	Characters          profileDefCharacters `json:"characters"`
}

type profileDefActivities struct {
	Data map[string]profileDefCharacter `json:"data"`
}

type profileDefCharacter struct {
	DateActivityStarted     string `json:"dateActivityStarted"`
	CurrentActivityHash     uint32 `json:"currentActivityHash"`
	CurrentActivityModeHash uint32 `json:"currentActivityModeHash"`
}

type profileDefCharacters struct {
	Data map[string]profileDefCharInfo `json:"data"`
}

type profileDefCharInfo struct {
	Light     int32 `json:"light"`
	ClassType int32 `json:"classType"`
}

// // /User/GetCredentialTypesForTargetAccount/{ActualMSID}/
// type credentialsTargetAccount struct {
// 	Response    []ctaResponse `json:"Response"`
// 	ErrorCode   int64         `json:"ErrorCode"`
// 	ErrorStatus string        `json:"ErrorStatus"`
// }

// type ctaResponse struct {
// 	CredentialType     int64  `json:"credentialType"`
// 	CredentialAsString string `json:"credentialAsString"`
// }

// Manifest: DestinyActivityDefinition
type activityDefinition struct {
	DP globalDisplayProperties `json:"displayProperties"`
	// ActivityLightLevel        int32              	`json:"activityLightLevel"`
	DestinationHash  uint32 `json:"destinationHash"`
	PlaceHash        uint32 `json:"placeHash"`
	ActivityTypeHash uint32 `json:"activityTypeHash"`
	// Tier                      int32              	`json:"tier"`
	// IsPlaylist                bool               	`json:"isPlaylist"`
	// Matchmaking               caDefMatchmaking   	`json:"matchmaking"`
	// IsPVP                     bool               	`json:"isPvP"`
	// ActivityLocationMappings  []caDefLocationMap	`json:"activityLocationMappings"`
}

// type caDefLocationMap struct {
// 	LocationHash     uint32  `json:"locationHash"`
// 	ActivationSource string `json:"activationSource"`
// 	ActivityHash     uint32  `json:"activityHash"`
// }

// type caDefMatchmaking struct {
// 	IsMatchmade          bool  `json:"isMatchmade"`
// 	MinParty             int32 `json:"minParty"`
// 	MaxParty             int32 `json:"maxParty"`
// 	MaxPlayers           int32 `json:"maxPlayers"`
// 	RequiresGuardianOath bool  `json:"requiresGuardianOath"`
// }

// Manifest: DestinyActivityModeDefinition
type activityModeDefinition struct {
	DP globalDisplayProperties `json:"displayProperties"`
	// IsTeamBased           bool   		`json:"isTeamBased"`
}

// Manifest: DestinyPlaceDefinition
type placeDefinition struct {
	DP globalDisplayProperties `json:"displayProperties"`
}

// /Destiny2/{MSType}/Profile/{ActualMSID}/Character/{charID}?components=202
type progressions struct {
	Response progressionsResponse `json:"Response"`
	genericErrorResponse
}

type progressionsResponse struct {
	Progressions progressionsClass `json:"progressions"`
}

type progressionsClass struct {
	Data progressionsData `json:"data"`
}

type progressionsData struct {
	Milestones map[string]progressionsMilestone `json:"milestones"`
}

type progressionsMilestone struct {
	Activities []progressionsActivity `json:"activities"`
}

type progressionsActivity struct {
	ActivityHash uint32               `json:"activityHash"`
	Phases       *[]progressionsPhase `json:"phases"`
}

type progressionsPhase struct {
	Complete bool `json:"complete"`
}
