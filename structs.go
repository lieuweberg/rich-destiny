package main

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


type manifestData struct {
	Response        manifestResponse	`json:"Response"`
	ErrorCode       int64       		`json:"ErrorCode"`
	ThrottleSeconds int64       		`json:"ThrottleSeconds"`
	ErrorStatus     string      		`json:"ErrorStatus"`
	Message         string      		`json:"Message"`
}

type manifestResponse struct {
	MobileWorldContentPaths        manifestWorldContentPaths       `json:"mobileWorldContentPaths"`
}

type manifestWorldContentPaths struct {
	En    string `json:"en"`
}


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


type characterActivitiesDefinition struct {
	Response        caDefReponse  `json:"Response"`
	ErrorCode       int64       `json:"ErrorCode"`
	ThrottleSeconds int64       `json:"ThrottleSeconds"`
	ErrorStatus     string      `json:"ErrorStatus"`
	Message         string      `json:"Message"`
}

type caDefReponse struct {
	CharacterActivities caDefActivities `json:"characterActivities"`
}

type caDefActivities struct {
	Data    map[string]caDefCharacter `json:"data"`
}

type caDefCharacter struct {
	DateActivityStarted     string              `json:"dateActivityStarted"`
	CurrentActivityHash     int64               `json:"currentActivityHash"`
	CurrentActivityModeHash int64               `json:"currentActivityModeHash"`
	LastCompletedStoryHash  int64               `json:"lastCompletedStoryHash"`
}


type currentActivityDefinition struct {
	DisplayProperties         caDefDisplay       	`json:"displayProperties"`
	OriginalDisplayProperties caDefDisplay       	`json:"originalDisplayProperties"`
	ReleaseIcon               string             	`json:"releaseIcon"`
	ReleaseTime               int64              	`json:"releaseTime"`
	ActivityLevel             int64              	`json:"activityLevel"`
	CompletionUnlockHash      int64              	`json:"completionUnlockHash"`
	ActivityLightLevel        int64              	`json:"activityLightLevel"`
	DestinationHash           int64              	`json:"destinationHash"`
	PlaceHash                 int64              	`json:"placeHash"`
	ActivityTypeHash          int64              	`json:"activityTypeHash"`
	Tier                      int64              	`json:"tier"`
	Modifiers                 []interface{}      	`json:"modifiers"`
	IsPlaylist                bool               	`json:"isPlaylist"`
	Matchmaking               caDefMatchmaking   	`json:"matchmaking"`
	IsPVP                     bool               	`json:"isPvP"`
	ActivityLocationMappings  []caDefLocationMap	`json:"activityLocationMappings"`
	Blacklisted               bool               	`json:"blacklisted"`
}

type caDefLocationMap struct {
	LocationHash     int64  `json:"locationHash"`
	ActivationSource string `json:"activationSource"`
	ActivityHash     int64  `json:"activityHash"`
}

type caDefDisplay struct {
	Description string `json:"description"`
	Name        string `json:"name"`
}

type caDefMatchmaking struct {
	IsMatchmade          bool  `json:"isMatchmade"`         
	MinParty             int64 `json:"minParty"`            
	MaxParty             int64 `json:"maxParty"`            
	MaxPlayers           int64 `json:"maxPlayers"`          
	RequiresGuardianOath bool  `json:"requiresGuardianOath"`
}


type currentActivityModeDefinition struct {
	DisplayProperties     camDefDisplay	`json:"displayProperties"`    
	PgcrImage             string 		`json:"pgcrImage"`
	ModeType              int64  		`json:"modeType"`
	ActivityModeCategory  int64  		`json:"activityModeCategory"`
	IsTeamBased           bool   		`json:"isTeamBased"`
	Tier                  int64  		`json:"tier"`
	IsAggregateMode       bool   		`json:"isAggregateMode"`
	ParentHashes          []int64		`json:"parentHashes"`
	FriendlyName          string 		`json:"friendlyName"`
	SupportsFeedFiltering bool   		`json:"supportsFeedFiltering"`
	Display               bool   		`json:"display"`
	Order                 int64  		`json:"order"`
	Hash                  int64  		`json:"hash"`
	Index                 int64  		`json:"index"`
	Redacted              bool   		`json:"redacted"`
	Blacklisted           bool   		`json:"blacklisted"`
}

type camDefDisplay struct {
	Description string `json:"description"`
	Name        string `json:"name"`       
	Icon        string `json:"icon"`       
	HasIcon     bool   `json:"hasIcon"`    
}


type placeDefinition struct {
	DisplayProperties placeDefDisplay `json:"displayProperties"`
	Hash              int64             `json:"hash"`             
	Index             int64             `json:"index"`            
	Redacted          bool              `json:"redacted"`         
	Blacklisted       bool              `json:"blacklisted"`      
}

type placeDefDisplay struct {
	Description string `json:"description"`
	Name        string `json:"name"`       
	Icon        string `json:"icon"`       
	HasIcon     bool   `json:"hasIcon"`    
}
