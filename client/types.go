package main

import richgo "github.com/lieuweberg/rich-go/client"

// Internal structs
// here: MSID means MembershipID
type storageStruct struct {
	AccessToken         string  `json:"access_token"`
	TokenType           string  `json:"token_type"`
	ExpiresIn           int64   `json:"expires_in"`
	RefreshToken        string  `json:"refresh_token"`
	RefreshExpiresIn    int64   `json:"refresh_expires_in"`
	BungieMSID          string  `json:"membership_id"`

	ActualMSID  string
	MSType      int64
	DisplayName string
	SteamID64	string
	RefreshAt   int64
	ReAuthAt    int64

	// Custom settings
	OrbitText       string  `json:"orbitText"`
	AutoUpdate      bool    `json:"autoUpdate"`
	JoinGameButton	bool	`json:"joinGameButton"`
	JoinOnlySocial	bool	`json:"joinOnlySocial"`
}

type currentProgramStatus struct {
	Status     	    string `json:"status"`
	Debug      	    string `json:"debug"`
	Version    	    string `json:"version"`
	Name            string `json:"name"`

	OrbitText       string `json:"orbitText"`
	AutoUpdate      bool   `json:"autoUpdate"`
	JoinGameButton  bool 	`json:"joinGameButton"`
	JoinOnlySocial	bool	`json:"joinOnlySocial"`
	
	Presence    richgo.Activity `json:"presence"`
}

type releasesFromGithub []releaseElement

type releaseElement struct {
	Name        string          `json:"name"`
	Draft       bool            `json:"draft"`
	Prerelease  bool            `json:"prerelease"`
	Assets      []releaseAsset  `json:"assets"`
}

type releaseAsset struct {
	BrowserDownloadURL  string  `json:"browser_download_url"`
	Name                string  `json:"name"`
}

// API/Manifest structs
// /Destiny2/Manifest
type manifestData struct {
	Response        manifestResponse    `json:"Response"`
	ErrorCode       int64               `json:"ErrorCode"`
	ThrottleSeconds int64               `json:"ThrottleSeconds"`
	ErrorStatus     string              `json:"ErrorStatus"`
	Message         string              `json:"Message"`
}

type manifestResponse struct {
	MobileWorldContentPaths manifestWorldContentPaths   `json:"mobileWorldContentPaths"`
}

type manifestWorldContentPaths struct {
	En  string  `json:"en"`
}

// /Destiny2/254/Profile/{BungieMSID}/LinkedProfiles
type linkedProfiles struct {
	Response    lPResponse  `json:"Response"`
	ErrorCode   int64       `json:"ErrorCode"`
}

type lPResponse struct {
	Profiles    []lPProfile `json:"profiles"`
}

type lPProfile struct {
	MembershipTypes []int64 `json:"applicableMembershipTypes"`
	MembershipType  int64   `json:"membershipType"`
	MembershipID    string  `json:"membershipId"`
	DisplayName     string  `json:"displayName"`
}

// DisplayProperties used for most Manifest structs
type globalDisplayProperties struct {
	// Description string `json:"description"`
	Name    string  `json:"name"`
}

// /Destiny2/{MSType}/Profile/{ActualMSID}?components=204,200
type profileDef struct {
	Response    profileDefReponse	`json:"Response"`
	ErrorStatus string          	`json:"ErrorStatus"`
	Message     string          	`json:"Message"`
}

type profileDefReponse struct {
	CharacterActivities profileDefActivities `json:"characterActivities"`
	Characters          profileDefCharacters 	`json:"characters"`
}

type profileDefActivities struct {
	Data    map[string]profileDefCharacter   `json:"data"`
}

type profileDefCharacter struct {
	DateActivityStarted     string  `json:"dateActivityStarted"`
	CurrentActivityHash     int64   `json:"currentActivityHash"`
	CurrentActivityModeHash int64   `json:"currentActivityModeHash"`
}

type profileDefCharacters struct {
	Data    map[string]profileDefCharInfo   `json:"data"`
}

type profileDefCharInfo struct {
	Light       int64   `json:"light"`
	ClassType   int64   `json:"classType"`
}

// /User/GetCredentialTypesForTargetAccount/{ActualMSID}/
type credentialsTargetAccount struct {
	Response        []ctaResponse  `json:"Response"`       
	ErrorCode       int64          `json:"ErrorCode"`
	ErrorStatus     string         `json:"ErrorStatus"`
}

type ctaResponse struct {
	CredentialType        int64   `json:"credentialType"`
	CredentialAsString    string  `json:"credentialAsString"`
}


// Manifest: DestinyActivityDefinition
type activityDefinition struct {
	DP                          globalDisplayProperties			`json:"displayProperties"`
	// ActivityLevel             int64              	`json:"activityLevel"`
	// ActivityLightLevel        int64              	`json:"activityLightLevel"`
	DestinationHash             int64              	`json:"destinationHash"`
	PlaceHash                   int64              	`json:"placeHash"`
	ActivityTypeHash            int64              	`json:"activityTypeHash"`
	// Tier                      int64              	`json:"tier"`
	// IsPlaylist                bool               	`json:"isPlaylist"`
	// Matchmaking               caDefMatchmaking   	`json:"matchmaking"`
	// IsPVP                     bool               	`json:"isPvP"`
	// ActivityLocationMappings  []caDefLocationMap	`json:"activityLocationMappings"`
}

// type caDefLocationMap struct {
// 	LocationHash     int64  `json:"locationHash"`
// 	ActivationSource string `json:"activationSource"`
// 	ActivityHash     int64  `json:"activityHash"`
// }

// type caDefMatchmaking struct {
// 	IsMatchmade          bool  `json:"isMatchmade"`
// 	MinParty             int64 `json:"minParty"`
// 	MaxParty             int64 `json:"maxParty"`
// 	MaxPlayers           int64 `json:"maxPlayers"`
// 	RequiresGuardianOath bool  `json:"requiresGuardianOath"`
// }

// Manifest: DestinyActivityModeDefinition
type currentActivityModeDefinition struct {
	DP globalDisplayProperties `json:"displayProperties"`
	// IsTeamBased           bool   		`json:"isTeamBased"`
	// Tier                  int64  		`json:"tier"`
}

// Manifest: DestinyPlaceDefinition
type placeDefinition struct {
	DP globalDisplayProperties `json:"displayProperties"`
}

// /Destiny2/{MSType}/Profile/{ActualMSID}/Character/{charID}?components=202
type progressions struct {
	Response        progressionsResponse    `json:"Response"`
	ErrorStatus     string      `json:"ErrorStatus"`
	Message     string          	`json:"Message"`
}

type progressionsResponse struct {
	Progressions progressionsClass `json:"progressions"`
}

type progressionsClass struct {
	Data    progressionsData  `json:"data"`
}

type progressionsData struct {
	Milestones	map[string]progressionsMilestone	`json:"milestones"`
}

type progressionsMilestone struct {
	Activities      []progressionsActivity       `json:"activities"`
}

type progressionsActivity struct {
	ActivityHash           int64           `json:"activityHash"`
	Phases                 *[]progressionsPhase         `json:"phases"`                          
}

type progressionsPhase struct {
	Complete  bool  `json:"complete"`
}