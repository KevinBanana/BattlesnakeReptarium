# BattlesnakeReptarium

Server for Battlesnake bots

## Setup

1. Copy a config \<env>.yml in the config folder and name it local.yml
2. Run main.go

## Bots

Every bot in `newBots` ([internal/server/server.go](internal/server/server.go)) is served at once, under its own path.

`GET /` is the service health check and lists the bots being served.

Adding a bot: implement `services.Bot` in its own package under
`internal/services/`, then add one entry to the map in `newBots`.