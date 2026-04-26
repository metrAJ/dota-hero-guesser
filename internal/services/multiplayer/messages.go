package multiplayer

type MessageType string

const (
	MsgTypeGuess           MessageType = "guess"
	MsgTypeWaitingForMatch MessageType = "waiting_for_match"
	MsgTypeMatchFound      MessageType = "match_found"
	MsgTypeGuessResult     MessageType = "guess_result"
	MsgTypeGameOver        MessageType = "game_over"
	MsgTypeError           MessageType = "error"
)

type ClientMessage struct {
	Type   MessageType `json:"type"`
	HeroID uint        `json:"hero_id"`
}

type ServerMessage struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}
