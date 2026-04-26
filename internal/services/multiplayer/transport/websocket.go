package multiplayer_handlers

import (
	"context"
	"doproj/internal/middleware"
	"doproj/internal/services/multiplayer"
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
)

type TicketStore interface {
	GenerateTicket(ctx context.Context, userID uint) (string, error)
	ConsumeTicket(ctx context.Context, ticket string) (uint, error)
}

type Matchmaker interface {
	AddPlayer(player *multiplayer.Player)
	GetUserNameByID(ctx context.Context, userID uint) (string, error)
	RemovePlayer(player *multiplayer.Player)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketHandler struct {
	ticketStore TicketStore
	matchmaker  Matchmaker
}

func NewWebSocketHandler(ts TicketStore, mm Matchmaker) *WebSocketHandler {
	return &WebSocketHandler{
		ticketStore: ts,
		matchmaker:  mm,
	}
}

func (h *WebSocketHandler) HandleConnections(w http.ResponseWriter, r *http.Request) {
	// Get ticket from query parameters
	ticketString := r.URL.Query().Get("ticket")
	if ticketString == "" {
		http.Error(w, "Missing ticket parameter", http.StatusBadRequest)
		return
	}
	// Validate ticket and get user ID
	userID, valid := h.ticketStore.ConsumeTicket(r.Context(), ticketString)
	if valid != nil {
		http.Error(w, "Invalid or expired ticket", http.StatusUnauthorized)
		return
	}
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}
	name, err := h.matchmaker.GetUserNameByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to retrieve user information", http.StatusInternalServerError)
		return
	}
	player := &multiplayer.Player{
		ID:   userID,
		Name: name,
		Conn: conn,
	}
	h.matchmaker.AddPlayer(player)

	/*
		defer func() {
			log.Printf("Player %d disconnected", player.ID)
			h.matchmaker.RemovePlayer(player)
			conn.Close()
		}()
	*/
}

// Endpoint for issuing ticket to authenticated users (called by frontend before connecting to WS)
func (h *WebSocketHandler) IssueTicket(w http.ResponseWriter, r *http.Request) {
	// Get ID from token
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Generate ticket for the user
	ticket, err := h.ticketStore.GenerateTicket(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to generate ticket", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"ticket": ticket,
	})
}
