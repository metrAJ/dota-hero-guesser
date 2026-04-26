package user_handlers

import (
	"context"
	"doproj/internal/middleware"
	"doproj/internal/models"
	"encoding/json"
	"net/http"
)

type UserService interface {
	CreateUser(ctx context.Context, user *models.User) error
	LoginUser(ctx context.Context, username, password string) (string, error)
	GetTopPlayers(ctx context.Context) ([]models.User, error)
	GetUserByID(ctx context.Context, userID uint) (*models.User, error)
}

type UserHandler struct {
	service UserService
}

func NewUserHandler(service UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

type userRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req userRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	user := &models.User{
		Username: req.Username,
		Password: req.Password,
	}

	err := h.service.CreateUser(r.Context(), user)
	if err != nil {
		// Should be different Status if DB is down, for now just 400
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "user created successfully"})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req userRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	token, err := h.service.LoginUser(r.Context(), req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}

func (h *UserHandler) GetTopPlayers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.GetTopPlayers(r.Context())
	if err != nil {
		http.Error(w, "Could not fetch top players", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	// Get ID from token
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Fetch user from DB
	user, err := h.service.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "Could not fetch user stats", http.StatusInternalServerError)
		return
	}
	stats := models.UserStatsResponse{
		Name:   user.Username,
		Wins:   user.Wins,
		Losses: user.Losses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
