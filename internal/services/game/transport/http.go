package game_handlers

import (
	"context"
	"doproj/internal/middleware"
	"doproj/internal/models"
	"encoding/json"
	"net/http"
)

// Defining interface
type GameService interface {
	GetAllHeroes(ctx context.Context) ([]models.Hero, error)
	GetGameView(ctx context.Context, userID uint) (*models.GameStateResponse, error)
	ProcessGuess(ctx context.Context, userID uint, roundID uint, guessID uint) (*models.GameStateResponse, error)
}

// Putting deffinitions in struct
type GameHandler struct {
	service GameService
}

// Injecting real functions into memory
func NewGameHandler(service GameService) *GameHandler {
	return &GameHandler{
		service: service,
	}
}

// Get game round for the user
func (h *GameHandler) GetRound(w http.ResponseWriter, r *http.Request) {
	// Authorise + get ID from JWT
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	response, err := h.service.GetGameView(ctx, userID)
	if err != nil {
		http.Error(w, "Could not load round", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *GameHandler) MakeGuess(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value(middleware.UserIDContextKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()

	var input struct {
		RoundID uint `json:"round_id"`
		GuessID uint `json:"guess_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	response, err := h.service.ProcessGuess(ctx, userID, input.RoundID, input.GuessID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *GameHandler) GetHeroes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	heroes, err := h.service.GetAllHeroes(ctx)
	if err != nil {
		http.Error(w, "Could not fetch heroes", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(heroes)
}
