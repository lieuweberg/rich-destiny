package main

// Empty ones need to be kept!
var largeImageActivityModes = map[string][]string{
	"anniversary": {"Dares of Eternity"},
	"beyondlight": {"Empire Hunt"},
	"control":     {"Control: Competitive", "Control: Quickplay"},
	"crucible": {"Rift", "Quickplay PvP", "Salvage", "Supremacy", "Mayhem", "Team Scorched", "Scorched", "Survival", "Competitive Co-Op",
		"Breakthrough", "Competitive PvP", "Momentum", "Zone Control", "Clash: Competitive", "Countdown", "Showdown", "Elimination",
		"Rumble", "Clash", "Lockdown", "Momentum Control", "The Crucible", "Classic Mix"},
	"destinylogo":        {},
	"doubles":            {"All Doubles"},
	"dungeon":            {},
	"explore":            {},
	"forge":              {"Forge Ignition"},
	"gambit":             {},
	"hauntedforest":      {},
	"ironbanner":         {},
	"lostsector":         {},
	"menagerie":          {"The Menagerie"},
	"nightmarehunt":      {},
	"privatecrucible":    {"Private Matches"},
	"raid":               {},
	"reckoning":          {"The Reckoning"},
	"seasonchosen":       {},
	"seasonhaunted":      {},
	"seasonlost":         {},
	"seasonrisen":        {},
	"seasonsplicer":      {},
	"shadowkeep":         {},
	"socialall":          {"Social", "All"},
	"storypvecoopheroic": {"Heroic Adventure", "Offensive", "Story", "PvE"},
	"strikes":            {"Scored Prestige Nightfall", "Scored Nightfall Strikes", "Nightfall Strikes", "Prestige Nightfall", "Strike", "Vanguard Op", "Nightfall"},
	"thenine":            {"Trials of the Nine", "Trials of the Nine Countdown", "Trials of the Nine Survival"},
	"thewitchqueen":      {},
	"trialsofosiris":     {"The Sundial", "Lighthouse Simulation"},
	"vexoffensive":       {},
	"wellspring":         {},
}

// Maps classID to class name
var classImages = map[int32]string{
	0: "titan",
	1: "hunter",
	2: "warlock",
}

var scoredLostSectors = []string{
	"K1",                            // The Moon
	"Concealed", "E15", "Perdition", // Europa
	"2A", "Veles", "Quarry", // The Cosmodrome
	"Scavenger's", "XII", // EDZ
	"Drowned", "Starlight", "Aphelion", // Dreaming City
	"Metamorphosis", "Sepulcher", "Extraction", // Savathûn's Throne World
	"Conflux",                                             // Nessus
	"Gilded Precept", "Thrilladrome", "Hydroponics Delta", // Neomuna
}
var storyMissions = map[string][]string{
	"shadowkeep":    {"A Mysterious Disturbance", "In Search of Answers", "Ghosts of Our Past", "In the Deep", "Beyond"},
	"beyondlight":   {"Darkness's Doorstep", "The New Kell", "Rising Resistance", "The Warrior", "The Technocrat", "The Kell of Darkness", "Sabotaging Salvation", "The Aftermath"},
	"thewitchqueen": {"The Arrival", "The Investigation", "The Ghosts", "The Communion", "The Mirror", "The Cunning", "The Last Chance", "The Ritual", "Memories of"},
	"lightfall":     {"First Contact", "Under Siege", "Downfall", "Breakneck", "On the Verge", "No Time Left", "Headlong", "Desperate Measures"},
}

var raidProgressionMap = map[string][]string{
	"Garden of Salvation": {"Embrace", "Undergrowth", "The Consecrated Mind", "The Sanctified Mind"},                                                   // 4
	"Last Wish":           {"Kalli, The Corrupted", "Shuro Chi, The Corrupted", "Morgeth, The Spirekeeper", "The Vault", "Riven of a Thousand Voices"}, // 5
	"Deep Stone Crypt":    {"Desolation / Crypt Security", "Atraks-1, Fallen Exo", "The Descent", "Taniks, The Abomination"},                           // 4
	"Vault of Glass":      {"Raise the Spire / Confluxes", "The Oracles", "The Templar", "The Gatekeeper", "Atheon, Time's Conflux"},                   // 5
	"Vow of the Disciple": {"Acquisition", "Collection", "Exhibition", "Dominion"},                                                                     // 4
	"King's Fall":         {"Totems", "Warpriest", "Golgoroth", "Daughters of Oryx", "Oryx, The Taken King"},                                           // 5
	"Root of Nightmares":  {"Cataclysm", "Scission", "Macrocosm", "Nezarec"},                                                                           // 4
}
