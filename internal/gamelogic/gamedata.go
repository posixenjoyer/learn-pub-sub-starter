package gamelogic

type Player struct {
	Username string       `json:"Username"`
	Units    map[int]Unit `json:"Units"`
}

type UnitRank string

const (
	RankInfantry  = "infantry"
	RankCavalry   = "cavalry"
	RankArtillery = "artillery"
)

type Unit struct {
	ID       int      `json:"ID"`
	Rank     UnitRank `json:"Rank"`
	Location Location `json:"Location"`
}

type ArmyMove struct {
	Player     Player   `json:"player"`
	Units      []Unit   `json:"units"`
	ToLocation Location `json:"to_location"`
}

type RecognitionOfWar struct {
	Attacker Player
	Defender Player
}

type Location string

func getAllRanks() map[UnitRank]struct{} {
	return map[UnitRank]struct{}{
		RankInfantry:  {},
		RankCavalry:   {},
		RankArtillery: {},
	}
}

func getAllLocations() map[Location]struct{} {
	return map[Location]struct{}{
		"americas":   {},
		"europe":     {},
		"africa":     {},
		"asia":       {},
		"australia":  {},
		"antarctica": {},
	}
}
