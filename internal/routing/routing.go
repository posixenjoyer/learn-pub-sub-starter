package routing

const (
	ArmyMovesPrefix       = "army_moves"
	WC                    = ".*"
	ArmyMovesWC           = ArmyMovesPrefix + WC
	WarRecognitionsPrefix = "war"
	WarWC                 = WarRecognitionsPrefix + WC
	PauseKey              = "pause"
	GameLogKey            = "game_logs.*"

	GameLogSlug = "game_logs"
)

const (
	ExchangePerilDirect = "peril_direct"
	ExchangePerilTopic  = "peril_topic"
)
