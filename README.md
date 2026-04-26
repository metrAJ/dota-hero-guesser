# Hero guessing game based on Dota 2 matches 
A real-time solo/multiplayer game built with GO. The main perpause of the game is to guess the correct Hero based on igame bought items, match outcome and hero attribute.
## Tech Stack
Main goal was to practise my skills, so at basis I used standard `net/http`.
* **Backend:** Golang, `net/http`
* **Database:** PostgreSQL + GORM
* **Authentication:** Custome JWT + Redis-backed Ticket system for WebSocket connection upgrades
* **Frontend** HTML, CSS and basic JS 
## Multiplayer 
Game incloodes solo expiriance made with basic GET/POST requests.
<img width="1856" height="934" alt="image" src="https://github.com/user-attachments/assets/66c3b919-f2be-4346-b864-7d3fa7cbcf2e" />
As for many it will be much more einteresting to compit with other players in real time, so I also made the duel mode with `gorilla/websocket`. 
<img width="1858" height="936" alt="image" src="https://github.com/user-attachments/assets/424698eb-1ccc-462f-93df-2e8feed97f37" />
* To handle concurrent game states I utilized an Actor model with gorotines and channels.
* Game handles disconnections, with corresponding game responses.
* Players are allowed to reconnect in 30 sec after connection loss.
## Architecture Overview
The project is splited in two parts: 
* `cmd/server`: The primary webserver with all user and game flows.
* `cmd/scraper`: Command line app to get real matches with OpenDota and scrap them into puzzles for the game.
## How to run
1. ** Clone the repository**
2. **Start external services:** PostgeSQL and Redis
3. **Environment Variables:** create `.env` file with DB credentials and JWT secret :
``` 
SERCRET_KEY = 
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
4. **Fill the DB with uptodate heroes/items and puzzles**
```bash
go run ./cmd/scraper -getitems
go run ./cmd/scraper -getheroes
go run ./cmd/scraper -getgames=50
```
5. **Run the core server:**
```bash
go run ./cmd/server   
```

