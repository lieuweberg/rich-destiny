package main

// Note: Only top-level structs should be used. The other's are there because of nesting.

// Semi-local structs
// here: MSID means MembershipID
type authResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	BungieMSID		 string `json:"membership_id"`

	ActualMSID		 string
	RefreshAt		 int64
	ReAuthAt		 int64
}

type guardianIcon struct {
	DisplayText	string
	Class 		string
}

type programStatusStruct struct {
	OK		bool
	Message	string
}

// API/Manifest structs
// /Destiny2/Manifest
type manifestData struct {
	Response        manifestResponse	`json:"Response"`
	ErrorCode       int64       		`json:"ErrorCode"`
	ThrottleSeconds int64       		`json:"ThrottleSeconds"`
	ErrorStatus     string      		`json:"ErrorStatus"`
	Message         string      		`json:"Message"`
}

type manifestResponse struct {
	MobileWorldContentPaths	manifestWorldContentPaths	`json:"mobileWorldContentPaths"`
}

type manifestWorldContentPaths struct {
	En    string `json:"en"`
}

// /Destiny2/254/Profile/{BungieMSID}/LinkedProfiles
type linkedProfiles struct {
	Response	lPResponse	`json:"Response"`
	ErrorCode	int64       `json:"ErrorCode"`
}

type lPResponse struct {
	Profiles	[]lPProfile	`json:"profiles"`
}

type lPProfile struct {
	MembershipType	int64	`json:"membershipType"`
	MembershipID	string	`json:"membershipId"`
	DisplayName		string	`json:"displayName"`
}

// DisplayProperties used for most Manifest structs
type allDisplay struct {
	// Description string `json:"description"`
	Name        string `json:"name"`
}

// /Destiny2/3/Profile/{ActualMSID}?components=204,200
type characterActivitiesDefinition struct {
	Response        caDefReponse  	`json:"Response"`
	ErrorStatus     string      	`json:"ErrorStatus"`
	// Message         string      	`json:"Message"`
}

type caDefReponse struct {
	CharacterActivities caDefActivities `json:"characterActivities"`
	Characters 			caDefCharacters `json:"characters"`
}

type caDefActivities struct {
	Data    map[string]caDefCharacter `json:"data"`
}

type caDefCharacter struct {
	DateActivityStarted     string              `json:"dateActivityStarted"`
	CurrentActivityHash     int64               `json:"currentActivityHash"`
	CurrentActivityModeHash int64               `json:"currentActivityModeHash"`
}

type caDefCharacters struct {
	Data    map[string]clDefDatum `json:"data"`
}

type clDefDatum struct {
	Light                    int64            `json:"light"`
	ClassType                int64            `json:"classType"`
}

// Manifest: DestinyActivityDefinition
type currentActivityDefinition struct {
	DisplayProperties         allDisplay			`json:"displayProperties"`
	// ActivityLevel             int64              	`json:"activityLevel"`
	// ActivityLightLevel        int64              	`json:"activityLightLevel"`
	// DestinationHash           int64              	`json:"destinationHash"`
	PlaceHash                 int64              	`json:"placeHash"`
	// ActivityTypeHash          int64              	`json:"activityTypeHash"`
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
	DisplayProperties     allDisplay	`json:"displayProperties"`
	// IsTeamBased           bool   		`json:"isTeamBased"`
	// Tier                  int64  		`json:"tier"`
}


// Manifest: DestinyPlaceDefinition
type placeDefinition struct {
	DisplayProperties	allDisplay	`json:"displayProperties"`
}