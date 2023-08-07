package repo

import (
	"context"
	"testing"

	"BattlesnakeReptarium/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	sampleGame = model.Game{
		ID: "sampleGame",
	}
)

func TestDatabase_CreateGame(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		WithDBSetup(func(db *Database) {
			err := db.CreateGame(context.Background(), sampleGame)
			assert.NoError(t, err)
		})

	})

	t.Run("Game Already Exists", func(t *testing.T) {
		WithDBSetup(func(db *Database) {
			err := db.CreateGame(context.Background(), sampleGame)
			require.NoError(t, err)
			err = db.CreateGame(context.Background(), sampleGame)
			assert.Error(t, err)
		})
	})

	t.Run("Game ID Cannot Be Empty", func(t *testing.T) {
		WithDBSetup(func(db *Database) {
			err := db.CreateGame(context.Background(), model.Game{})
			assert.Error(t, err)
		})
	})
}

func TestDatabase_GetGame(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		WithDBSetup(func(db *Database) {
			err := db.CreateGame(context.Background(), model.Game{ID: "gameId"})
			require.NoError(t, err)

			returnedGame, err := db.GetGame(context.Background(), "gameId")
			assert.NoError(t, err)
			assert.NotNil(t, returnedGame)
		})
	})

	t.Run("Game Not Found", func(t *testing.T) {
		WithDBSetup(func(db *Database) {
			returnedGame, err := db.GetGame(context.Background(), "nonexistingGame")
			assert.Error(t, err)
			assert.Nil(t, returnedGame)
		})
	})
}

func WithDBSetup(testFunc func(db *Database)) {
	db := Database{
		games: make(map[string]*model.Game),
	}

	testFunc(&db)
}
