# Hero guessing game based on Dota 2 matches 
A real-time solo/multiplayer game built with GO. The main perpause of the game is to guess the correct Hero based on in-game bought items, match outcomes and hero attributes.
## Tech Stack
The main goal of this project was to practice core backend mechanics, so I deliberately utilized the standard library (net/http) over heavy web frameworks.
* **Backend:** Golang, `net/http`
* **Database:** PostgreSQL + GORM
* **Authentication:** Custome JWT + Redis-backed Ticket system for WebSocket connection upgrades
* **Frontend** HTML, CSS and basic JS 
## Multiplayer 
The game includes a solo experience built with standard RESTful GET/POST requests.
<img width="1856" height="934" alt="image" src="https://github.com/user-attachments/assets/66c3b919-f2be-4346-b864-7d3fa7cbcf2e" />
Because predicting builds is much more interesting when competing against others, I also built a real-time Duel Mode using gorilla/websocket.
<img width="1858" height="936" alt="image" src="https://github.com/user-attachments/assets/424698eb-1ccc-462f-93df-2e8feed97f37" />
* Concurrency: Handled concurrent game states by utilizing the Actor model with Go channels and goroutines (avoiding mutex locks).
* Resilience: The backend safely handles unexpected player disconnections with corresponding game state updates.
* Reconnection: Players are allowed a 30-second window to seamlessly reconnect to an active duel after a connection loss.
## Architecture Overview
The project is split into two distinct binaries to separate concerns: 
* `cmd/server`: The primary web server handling all user authentication, API routes, and real-time WebSocket game flows.
* `cmd/scraper`: a Command-line app to get real matches with OpenDota and scrap them into puzzles for the game.
## How to run
1. ** Clone the repository**
2. **Start external services:** PostgeSQL and Redis
3. **Environment Variables:** create `.env` file with your DB credentials and JWT secret :
``` 
JWT_SECRET = 
TOKEN_EXPIRATION_HOURS = 

DB_HOST = 
DB_PORT = 
DB_USER = 
DB_PASSWORD =
DB_NAME = 
DB_SSLMODE = 
DB_TIMEZONE = 

REDIS_ADDR =
REDIS_PASSWORD = 
```
4. **Seed the Database:** Fill the database with up-to-date heroes, items, and initial puzzles using the scraper tool:
```bash
go run ./cmd/scraper -getitems
go run ./cmd/scraper -getheroes
go run ./cmd/scraper -getgames=50
```
5. **Run the core server:**
```bash
go run ./cmd/server   
```

