package main

// Maps collective name (large images) to []ActivityModeHash
var largeImageMap = map[string][]int32{
	"control": {-1270265980, -1095868816, -621632665, 1340118533}, // Private Matches Control, Control, Control: Quickplay, Control: Competitive
	// Survival, Competitive PvP, Showdown, Momentum, Quickplay PvP, Salvage, Breakthrough, Clash: Competitive, Elimination, Rumble,
	"crucible": {-2055718213, -1808243978, -1748182994, -1381428397, -869856616, -390342404, -261966967, -242891787, -216527492, 157639802,
		// Clash: Quickplay, Supremaxy, Crucible, Mayhem, Team Scorched, Countdown, Clash, Scorched, Lockdown
		244324567, 910991990, 1164760504, 1264443021, 1372032677, 1505888634, 1585787840, 1668923154, 2096553452},
	// "doubles": {-473465279, 946648766, 1703282980}, // Doubles, Crimson Doubles, All Doubles
	"dungeon": {608898761}, // Dungeon
	"explore": {-797199657}, // Explore
	// "forge": {803838459}, // Forge, doesn't really work though but oh well
	"gambit": {1418469392, 1848252830}, // Gambit Prime, Gambit
	// Iron Banner Control, Iron Banner Clash, Iron Banner Supremacy, Iron Banner Salvage, Iron Banner
	"ironbanner": {-1451282428, -815750908, -82084646, 1317961215, 1826469369},
	// "menagerie": {400075666}, // The Menagerie
	"nightmarehunt": {332181804}, // Nightmare Hunt
	// Private Matches Rumble, - Survival, - Supremacy, - Mayhem, -, - Countdown, - Clash
	"privatecrucible": {-1741423138, -514207857, -432922534, 116827562, 122250361, 152599670, 575409284},
	"raid": {2043403989}, // Raid
	// "reckoning": {-400492470}, // The Reckoning
	"socialall": {748895195, 1589650888}, // All (idk what this is but sure), Social
	"storypvecoopheroic": {-1845790942, 175275639, 1164760493, 1686739444}, // Competitive Co-Op, Heroic Adventure, PvE, Story
	// Strikes, Nightfall Strikes, Normal Strikes, Scored Prestige Nightfall, Scored Nightfall Strikes, Prestige Nightfall
	"strikes": {-1900351293, -505945566, -184361721, 532484583, 547513715, 1350109474},
	// "thenine": {-2031201205, 470484296, 1370326378}, // Trials Of The Nine Countdown, - Survival, -
	"trialsofosiris": {-1975465249, 1673724806}, // The Sundial, Trials Of Osiris
	// "vexoffensive": {1963485238}, // Vex Offensive
}

// Maps ActivityHash to Forge name (because they are seen as story)
// var forgeHashMap = map[int32]string{
// 	10898844: "Bergusia",
// 	1019949956: "Volundr",
// 	1483179969: "Gofannon",
// 	1878615566: "Izanami",
// }

// Maps classID to class name
var classImageMap = map[int64]string{
	0: "titan",
	1: "hunter",
	2: "warlock",
}

var scoredLostSectors = []string{"K1", "Concealed", "E15", "Perdition", "2A", "Veles", "Quarry", "Scavenger's", "XII", "Empty Tank", "Drowned", "Starlight", "Aphelion"}

var raidProgressionMap = map[string][]string{
	// "gos": {}, // Unimplemented in milestone
	"dsc": {"Desolation / Crypt Security", "Atraks-1, Fallen Exo", "The Descent", "Taniks, The Abomination"}, // 4
	"lw":  {"Kalli, The Corrupted", "Shuro Chi, The Corrupted", "Morgeth, The Spirekeeper", "The Vault", "Riven of a Thousand Voices"}, // 5
	"vog": {"Raise the Spire / Confluxes", "The Oracles", "The Templar", "The Gatekeeper", "Atheon, Time's Conflux"}, // 5
}