package multiplayer

import (
	"context"
	"doproj/internal/models"
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type MultiplayerRepository interface {
	CreateMultiplayerMatch(ctx context.Context, match *models.MultiplayerMatch) error
	UpdateMultiplayerMatch(ctx context.Context, match *models.MultiplayerMatch) error
	FindUnplayedMutualPuzzle(ctx context.Context, p1ID, p2ID uint) (*models.Puzzle, error)
	GetItemsByIDs(ctx context.Context, itemIDs []int64) ([]models.Item, error)
	GetUserByID(ctx context.Context, userID uint) (*models.User, error)
	GetHeroByID(ctx context.Context, heroID uint) (*models.Hero, error)
	IncrementWins(ctx context.Context, userID uint) error
	IncrementLosses(ctx context.Context, userID uint) error
}

type Matchmaker struct {
	repo          MultiplayerRepository
	Waiting       *Player
	ActiveMatches map[uint]*Match
	mu            sync.Mutex // Mutex because queue is shared between goroutines
}

func NewMatchmaker(repo MultiplayerRepository) *Matchmaker {
	return &Matchmaker{
		repo:          repo,
		ActiveMatches: make(map[uint]*Match),
	}
}

func (m *Matchmaker) GetUserNameByID(ctx context.Context, userID uint) (string, error) {
	user, err := m.repo.GetUserByID(ctx, userID)
	if err != nil {
		return "", err
	}
	return user.Username, nil
}

func (m *Matchmaker) RemovePlayer(player *Player) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Waiting != nil && m.Waiting.ID == player.ID && m.Waiting.Conn == player.Conn {
		log.Printf("Player %d left the matchmaking queue", player.ID)
		m.Waiting = nil
	}

}

func (m *Matchmaker) AddPlayer(player *Player) {
	opponent := m.popQueueOrWait(player)

	if opponent == nil {
		return
	}

	go m.startMAtch(opponent, player)
}

func (m *Matchmaker) popQueueOrWait(player *Player) *Player {
	m.mu.Lock()
	defer m.mu.Unlock()

	if match, exists := m.ActiveMatches[player.ID]; exists && !match.isOver {
		log.Printf("Player %d reconnecting to active match", player.ID)
		match.Reconnect(player) // Send them straight to their match!
		return nil              // Return nil so Matchmaker knows NOT to queue them
	}

	if m.Waiting != nil && m.Waiting.ID == player.ID {
		log.Printf("Player %d reconnected, updating queue.", player.ID)
		m.Waiting.Conn.Close()
		m.Waiting = player
		m.notifyPlayer(player, MsgTypeWaitingForMatch, "Waiting for an opponent.")
		return nil
	} else if m.Waiting == nil {
		log.Printf("Player %d is waiting for an opponent", player.ID)
		m.Waiting = player
		m.notifyPlayer(player, MsgTypeWaitingForMatch, "Waiting for an opponent.")
		return nil
	}

	opponent := m.Waiting
	m.Waiting = nil
	return opponent
}

func (m *Matchmaker) startMAtch(player1, player2 *Player) {
	p1Alive := m.isAlive(player1)
	p2Alive := m.isAlive(player2)

	if !p1Alive && !p2Alive {
		log.Printf("Both players %d and %d disconnected before match start.", player1.ID, player2.ID)
		return
	}
	if !p1Alive {
		log.Printf("Player %d was a ghost. Recycling Player %d.", player1.ID, player2.ID)
		m.notifyPlayer(player2, MsgTypeWaitingForMatch, "Opponent disconnected. Re-entering queue...")
	}
	if !p2Alive {
		log.Printf("Player %d was a ghost. Recycling Player %d.", player2.ID, player1.ID)
		m.notifyPlayer(player1, MsgTypeWaitingForMatch, "Opponent disconnected. Re-entering queue...")
	}

	fmt.Printf("Player %d successfully matched with Player %d\n", player1.ID, player2.ID)
	ctx := context.Background()

	puzzle, err := m.repo.FindUnplayedMutualPuzzle(ctx, player1.ID, player2.ID)
	if err != nil {
		log.Printf("Error finding mutual puzzle for players %d and %d: %v", player1.ID, player2.ID, err)
		m.notifyPlayer(player1, MsgTypeError, "Failed to find a puzzle for the match. Please try again later.")
		m.notifyPlayer(player2, MsgTypeError, "Failed to find a puzzle for the match. Please try again later.")
		player1.Conn.Close()
		player2.Conn.Close()
		return
	}

	dbMatch := &models.MultiplayerMatch{
		Player1ID: player1.ID,
		Player2ID: player2.ID,
		PuzzleID:  puzzle.ID,
		Status:    models.StatusPlaying,
	}

	if err := m.repo.CreateMultiplayerMatch(ctx, dbMatch); err != nil {
		log.Printf("Error creating multiplayer match: %v", err)
		return
	}
	mainItems, err := m.repo.GetItemsByIDs(ctx, puzzle.ItemIDs)

	if err != nil {
		log.Printf("Error getting main items: %v", err)
		return
	}
	backpackItems, err := m.repo.GetItemsByIDs(ctx, puzzle.BackPackIDs)
	if err != nil {
		log.Printf("Error getting backpack items: %v", err)
		return
	}
	hero, _ := m.repo.GetHeroByID(ctx, puzzle.HeroID)

	match := NewMatch(dbMatch, []*Player{player1, player2}, puzzle, mainItems, backpackItems, hero.Type, m.handleMatchEnd)

	payload1 := map[string]interface{}{
		"match_id":      dbMatch.ID,
		"my_name":       player1.Name,
		"opponent_name": player2.Name,
	}
	payload2 := map[string]interface{}{
		"match_id":      dbMatch.ID,
		"my_name":       player2.Name,
		"opponent_name": player1.Name,
	}

	m.notifyPlayer(player1, MsgTypeMatchFound, payload1)
	m.notifyPlayer(player2, MsgTypeMatchFound, payload2)

	m.mu.Lock()
	m.ActiveMatches[player1.ID] = match
	m.ActiveMatches[player2.ID] = match
	m.mu.Unlock()

	match.StartGame()
}

func (m *Matchmaker) isAlive(p *Player) bool {
	err := p.Conn.WriteMessage(websocket.PingMessage, []byte{})
	return err == nil
}

func (m *Matchmaker) notifyPlayer(p *Player, msgType MessageType, payload interface{}) {
	p.Conn.WriteJSON(ServerMessage{
		Type:    msgType,
		Payload: payload,
	})
}

func (m *Matchmaker) handleMatchEnd(updatedMatch *models.MultiplayerMatch) {
	ctx := context.Background()

	err := m.repo.UpdateMultiplayerMatch(ctx, updatedMatch)
	if err != nil {
		log.Printf("Error updating multiplayer match: %v", err)
		return
	}

	for _, playerID := range []uint{updatedMatch.Player1ID, updatedMatch.Player2ID} {
		if playerID == updatedMatch.WinnerId {
			if err := m.repo.IncrementWins(ctx, playerID); err != nil {
				log.Printf("Error updating user wins: %v", err)
			}
		} else {
			if err := m.repo.IncrementLosses(ctx, playerID); err != nil {
				log.Printf("Error updating user losses: %v", err)
			}
		}

	}
}
