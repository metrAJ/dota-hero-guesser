package multiplayer

import (
	"context"
	"doproj/internal/models"
	"log"
	"maps"
	"time"

	"github.com/gorilla/websocket"
)

type EventType string

const (
	EventGuess      EventType = "guess"
	EventTimeout    EventType = "timeout"
	EventDisconnect EventType = "disconnect"
	EventNextPhase  EventType = "next_phase"
)

type GameEvent struct {
	Type     EventType
	PlayerID uint
	HeroID   uint
}

type OnGameEndCallback func(updatedMatch *models.MultiplayerMatch)

type Match struct {
	DBMatch       *models.MultiplayerMatch
	Puzzle        *models.Puzzle
	MainItems     []models.Item
	BackpackItems []models.Item
	HeroAtribute  string
	saveResult    OnGameEndCallback
	isOver        bool

	Players        map[uint]*Player
	WrongGuesses   map[string][]uint
	CurrentGuesses map[uint]Guess

	CurrentPhase uint
	phaseTimer   *time.Timer

	events chan GameEvent

	ctx    context.Context
	cancel context.CancelFunc
}

type Player struct {
	ID   uint
	Name string
	Conn *websocket.Conn
}

type Guess struct {
	HeroID    uint
	Timestamp time.Time
}

func NewMatch(dbm *models.MultiplayerMatch, players []*Player, puzzle *models.Puzzle, mainItems []models.Item, backpackItems []models.Item, atr string, saveResult OnGameEndCallback) *Match {
	ctx, cancel := context.WithCancel(context.Background())
	m := &Match{
		DBMatch:        dbm,
		Players:        make(map[uint]*Player),
		Puzzle:         puzzle,
		MainItems:      mainItems,
		BackpackItems:  backpackItems,
		HeroAtribute:   atr,
		saveResult:     saveResult,
		isOver:         false,
		WrongGuesses:   make(map[string][]uint),
		CurrentGuesses: make(map[uint]Guess),
		events:         make(chan GameEvent, 20),
		ctx:            ctx,
		cancel:         cancel,
	}

	for _, p := range players {
		m.Players[p.ID] = p
		m.WrongGuesses[p.Name] = []uint{}
	}
	return m
}

func (m *Match) StartGame() {
	for _, p := range m.Players {
		player := p
		go m.listenToPlayer(player, player.Conn)
	}
	go m.runEventLoop()
}

func (m *Match) Reconnect(newPlayer *Player) {
	player, exists := m.Players[newPlayer.ID]
	if !exists || m.isOver {
		return
	}
	log.Printf("Player %d successfully reconnected!\n", player.ID)
	player.Conn = newPlayer.Conn // Attach the new, fresh websocket!

	var opponent *Player
	for id, p := range m.Players {
		if id != player.ID {
			opponent = p
		}
	}

	player.Conn.WriteJSON(ServerMessage{
		Type: "match_found",
		Payload: map[string]interface{}{
			"match_id":      m.DBMatch.ID,
			"my_name":       player.Name,
			"opponent_name": opponent.Name,
		},
	})

	payload := m.phaseInfo()
	_, resultsPayload := m.gradeGuesses()

	maps.Copy(payload, resultsPayload)
	player.Conn.WriteJSON(ServerMessage{
		Type:    "phase_start",
		Payload: payload,
	})

	go m.listenToPlayer(player, player.Conn)

	if opponent != nil && opponent.Conn != nil {
		opponent.Conn.WriteJSON(ServerMessage{
			Type:    "waiting",
			Payload: "Суперник повернувся! Продовжуємо гру...", // Opponent returned! Resuming...
		})
	}

}

func (m *Match) listenToPlayer(player *Player, localConn *websocket.Conn) {
	defer func() {
		if localConn != nil {
			localConn.Close()
		}
		if player.Conn == localConn {
			m.events <- GameEvent{Type: EventDisconnect, PlayerID: player.ID}
		}

	}()

	for {
		var clientMsg ClientMessage
		err := localConn.ReadJSON(&clientMsg)
		if err != nil {
			return // The network dropped! Break the loop and trigger the defer.
		}
		if clientMsg.Type == MsgTypeGuess {
			m.events <- GameEvent{Type: EventGuess, PlayerID: player.ID, HeroID: clientMsg.HeroID}
		}
	}

}

func (m *Match) runEventLoop() {
	m.startPhase()

	for {
		select {
		case <-m.ctx.Done():
			// Shut down Goroutine when match is over
			return
		case event := <-m.events:
			if m.isOver {
				continue
			}
			switch event.Type {
			case EventGuess:
				m.handleGuess(event.PlayerID, event.HeroID)
			case EventTimeout:
				m.resolvePhase()
			case EventNextPhase:
				m.startPhase()
			case EventDisconnect:
				m.handleDisconnect(event.PlayerID)
			}
		}
	}
}

func (m *Match) phaseInfo() map[string]interface{} {
	hints := make(map[string]interface{})
	hints["round"] = m.CurrentPhase
	hints["time_limit"] = 30
	hints["main_items"] = m.MainItems
	if m.CurrentPhase > 3 {
		hints["backpack_items"] = m.BackpackItems
	}
	if m.CurrentPhase > 4 {
		hints["is_won"] = m.Puzzle.IsWon
	}
	if m.CurrentPhase > 5 {
		hints["hero_attribute"] = m.HeroAtribute
	}
	return hints
}

func (m *Match) startPhase() {
	m.CurrentPhase++

	m.CurrentGuesses = make(map[uint]Guess)

	hints := m.phaseInfo()
	m.broadcast("phase_start", hints)
	m.phaseTimer = time.AfterFunc(30*time.Second, func() {
		m.events <- GameEvent{Type: EventTimeout}
	})
}

func (m *Match) resolvePhase() {
	correctPlayers, resultsPayload := m.gradeGuesses()
	m.broadcast("phase_results", resultsPayload)

	winnerID := m.determinePhaseWinner(correctPlayers)

	if winnerID != nil {
		m.endGame(*winnerID, m.Puzzle.HeroID, "guessed_hero")
		return
	}
	if m.CurrentPhase >= 20 {
		m.endGameDraw()
		return
	}
	time.AfterFunc(time.Second, func() {
		m.events <- GameEvent{Type: EventNextPhase}
	})
}

func (m *Match) gradeGuesses() ([]uint, map[string]interface{}) {
	var correctPlayers []uint
	guessesPayload := make(map[uint]uint)

	for playerID, guess := range m.CurrentGuesses {
		guessesPayload[playerID] = guess.HeroID

		if guess.HeroID == m.Puzzle.HeroID {
			correctPlayers = append(correctPlayers, playerID)
		} else {
			player := m.Players[playerID]
			m.WrongGuesses[player.Name] = append(m.WrongGuesses[player.Name], guess.HeroID)
		}
	}
	payload := map[string]interface{}{
		"guesses":       guessesPayload,
		"wrong_guesses": m.WrongGuesses,
	}
	return correctPlayers, payload
}

func (m *Match) determinePhaseWinner(correctPlayers []uint) *uint {
	if len(correctPlayers) == 0 {
		return nil
	}
	if len(correctPlayers) == 1 {
		return &correctPlayers[0]
	}
	fastestPlayerID := correctPlayers[0]
	fastestTime := m.CurrentGuesses[fastestPlayerID].Timestamp

	for i := 1; i < len(correctPlayers); i++ {
		candidateID := correctPlayers[i]
		candidateTime := m.CurrentGuesses[candidateID].Timestamp

		if candidateTime.Before(fastestTime) {
			fastestPlayerID = candidateID
			fastestTime = candidateTime
		}
	}
	return &fastestPlayerID
}

func (m *Match) endGame(winnerID uint, heroID uint, reason string) {
	m.isOver = true
	m.DBMatch.WinnerId = winnerID
	player := m.Players[winnerID]
	m.DBMatch.Status = "completed"
	m.saveResult(m.DBMatch)
	m.cancel()

	m.broadcast(MsgTypeGameOver, map[string]interface{}{
		"winner_name": player.Name,
		"hero_id":     heroID,
		"reason":      reason,
	})
}

func (m *Match) handleGuess(playerID uint, heroID uint) {
	player, exists := m.Players[playerID]
	if !exists || m.isOver {
		return
	}
	if _, alreadyGuessed := m.CurrentGuesses[playerID]; alreadyGuessed {
		if player.Conn != nil {
			player.Conn.WriteJSON(ServerMessage{
				Type:    MsgTypeError,
				Payload: "Already guessed.",
			})
		}
		return
	}

	m.CurrentGuesses[playerID] = Guess{HeroID: heroID, Timestamp: time.Now()}
	if player.Conn != nil {
		player.Conn.WriteJSON(ServerMessage{
			Type:    MsgTypeGuess,
			Payload: " Guess received.",
		})
	}
	if len(m.CurrentGuesses) == len(m.Players) {
		if m.phaseTimer != nil {
			m.phaseTimer.Stop()
		}
		m.resolvePhase()
	}
}

/*
func (m *Match) handleDisconnect(playerID uint) {
	fmt.Printf("Player %d disconnected\n", playerID)
	delete(m.Players, playerID)
	if len(m.Players) == 1 {
		var lastRemainingID uint
		for id := range m.Players {
			lastRemainingID = id
		}
		m.endGame(lastRemainingID, m.Puzzle.HeroID)
		return
	} else if len(m.Players) == 0 {
		m.endGameDraw()
		return
	}
}
*/

func (m *Match) handleDisconnect(playerID uint) {
	player, exists := m.Players[playerID]
	if !exists {
		log.Printf("Player %d Does Not Exist.\n", playerID)
		return
	}
	if player.Conn == nil {
		m.handleAbandon(playerID)
		return
	}

	log.Printf("Player %d lost connection. 30s grace period started.\n", playerID)
	player.Conn = nil

	var survivor *Player
	for id, p := range m.Players {
		if id != playerID {
			survivor = p
		}
	}
	if survivor != nil && survivor.Conn != nil {
		survivor.Conn.WriteJSON(ServerMessage{
			Type:    "waiting",
			Payload: "Opponent disconnected, you will win in 30 seconds if they don't return...",
		})
	}

	go func() {
		time.Sleep(30 * time.Second)
		player, exists := m.Players[playerID]
		if !m.isOver && player.Conn == nil && exists {
			m.events <- GameEvent{Type: EventDisconnect, PlayerID: playerID}
		}
	}()
}

func (m *Match) handleAbandon(playerID uint) {
	player, exists := m.Players[playerID]
	if !exists || player.Conn != nil {
		return
	}

	log.Printf("Player %d abandoned. Ending match.\n", playerID)
	delete(m.Players, playerID)

	if len(m.Players) == 1 {
		var survivor *Player
		for _, p := range m.Players {
			survivor = p
		}
		m.endGame(survivor.ID, m.Puzzle.HeroID, "opponent_disconnected")
	}

}

func (m *Match) endGameDraw() {
	m.isOver = true
	m.DBMatch.Status = "draw"
	m.saveResult(m.DBMatch)
	m.cancel()
}

func (m *Match) broadcast(msgType MessageType, payload interface{}) {
	for _, player := range m.Players {
		if player.Conn != nil {
			err := player.Conn.WriteJSON(ServerMessage{
				Type:    msgType,
				Payload: payload,
			})
			if err != nil {
				log.Printf("Warning: failed to broadcast to player %d: %v\n", player.ID, err)
			}
		}
	}
}
