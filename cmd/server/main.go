package main

import (
	"doproj/internal/auth"
	"doproj/internal/config"
	"doproj/internal/database"
	"doproj/internal/middleware"
	"doproj/internal/repository"
	"doproj/internal/services/game"
	game_handlers "doproj/internal/services/game/transport"
	"doproj/internal/services/multiplayer"
	multiplayer_handlers "doproj/internal/services/multiplayer/transport"
	"doproj/internal/services/user"
	user_handlers "doproj/internal/services/user/transport"
	"log"
	"net/http"
)

func main() {
	// Loading configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("TERMINAL ERROR: Could not initialize configuration: %v", err)
	}
	// Initializing DB
	db := database.InitDB(cfg)

	// Initializing Redis
	rdb := database.InitRedis(cfg)

	// Initializing Ticket Store
	ticketStore := auth.NewRedisTicketStore(rdb)

	// Injecting Database into Repository
	masterRepo := repository.NewGormRepo(db)

	// Create TokenManager
	tokenManager := auth.NewTokenManager(cfg)
	// Give it to Auth Middleware
	authBouncer := middleware.NewAuthMiddleware(tokenManager)

	// Injecting Game Service and Handlers
	gameService := game.NewGameService(masterRepo)
	gameHandler := game_handlers.NewGameHandler(gameService)

	// Injecting User Service and Handlers
	authService := user.NewUserService(masterRepo, tokenManager)
	authHandler := user_handlers.NewUserHandler(authService)

	// Injecting Multiplayer and WebSocket Handlers
	multiplayerService := multiplayer.NewMatchmaker(masterRepo)
	wsHandler := multiplayer_handlers.NewWebSocketHandler(ticketStore, multiplayerService)

	// Base Game API
	http.HandleFunc("/round", authBouncer.Authenticate(gameHandler.GetRound))
	http.HandleFunc("POST /round", authBouncer.Authenticate(gameHandler.MakeGuess))
	http.HandleFunc("GET /heroes", gameHandler.GetHeroes)

	// Multiplayer API
	http.HandleFunc("/ws", wsHandler.HandleConnections)
	http.HandleFunc("POST /ws-ticket", authBouncer.Authenticate(wsHandler.IssueTicket))

	// User API
	http.HandleFunc("POST /register", authHandler.Register)
	http.HandleFunc("POST /login", authHandler.Login)
	http.HandleFunc("GET /top-players", authHandler.GetTopPlayers)
	http.HandleFunc("/stats", authBouncer.Authenticate(authHandler.GetUserStats))

	// Serving static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Serving Pages
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/login.html")
	})
	http.HandleFunc("/game", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/game.html")
	})
	http.HandleFunc("/duel", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/duel.html")
	})

	// Starting the server
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
