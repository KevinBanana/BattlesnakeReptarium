package repo

import (
	"context"
	"errors"
	"sync"

	"BattlesnakeReptarium/internal/model"
)

// ErrGameNotFound is a sentinel so callers can react to a missing game without
// matching on error strings. It is routine after a restart: games live only in
// memory, and a deploy mid-game loses them.
var ErrGameNotFound = errors.New("game not found")

type DB interface {
	CreateGame(ctx context.Context, game model.Game) error
	GetGame(ctx context.Context, id string) (*model.Game, error)
	UpdateGame(ctx context.Context, game model.Game) error
}

// Database is just an in-memory store. Games are lost when the process exits
type Database struct {
	sync.RWMutex
	games map[string]*model.Game
}

func NewDatabase() *Database {
	return &Database{games: make(map[string]*model.Game)}
}

func (db *Database) CreateGame(ctx context.Context, game model.Game) error {
	db.Lock()
	defer db.Unlock()

	if game.ID == "" {
		return errors.New("game id cannot be empty")
	}

	if _, ok := db.games[game.ID]; ok {
		return errors.New("game already exists")
	}

	db.games[game.ID] = &game
	return nil
}

func (db *Database) GetGame(ctx context.Context, id string) (*model.Game, error) {
	db.RLock()
	defer db.RUnlock()
	var retrievedGame *model.Game
	var ok bool

	if retrievedGame, ok = db.games[id]; !ok {
		return nil, ErrGameNotFound
	}
	return retrievedGame, nil
}

func (db *Database) UpdateGame(ctx context.Context, game model.Game) error {
	db.Lock()
	defer db.Unlock()

	if _, ok := db.games[game.ID]; !ok {
		return ErrGameNotFound
	}

	db.games[game.ID] = &game
	return nil
}
