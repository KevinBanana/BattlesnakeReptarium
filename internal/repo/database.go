package repo

import (
	"context"
	"sync"

	"BattlesnakeReptarium/internal/model"

	"github.com/pkg/errors"
)

type DB interface {
	CreateGame(ctx context.Context, game model.Game) error
	GetGame(ctx context.Context, id string) (*model.Game, error)
}

type Database struct {
	sync.RWMutex
	games map[string]*model.Game
}

func NewDatabase() *Database {
	return &Database{
		games: make(map[string]*model.Game),
	}
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
		return nil, errors.New("game not found")
	}
	return retrievedGame, nil
}
