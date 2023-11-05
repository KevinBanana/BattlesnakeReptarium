package services

import (
	"context"
	"errors"
	"testing"

	"BattlesnakeReptarium/internal/model"
	"BattlesnakeReptarium/internal/repo"
	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
)

func TestGameEngineSvc_StartGame(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		withGameEngineSetup(t, func(b gameEngineTestBundle) {
			b.mockDB.EXPECT().CreateGame(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			err := b.gameEngineSvc.StartGame(context.TODO(), model.Game{}, model.Board{}, model.Snake{})
			assert.NoError(t, err)
		})
	})

	t.Run("Failed to save", func(t *testing.T) {
		withGameEngineSetup(t, func(b gameEngineTestBundle) {
			b.mockDB.EXPECT().CreateGame(gomock.Any(), gomock.Any()).Return(errors.New("err")).Times(1)
			err := b.gameEngineSvc.StartGame(context.TODO(), model.Game{}, model.Board{}, model.Snake{})
			assert.Error(t, err)
		})
	})
}

func TestGameEngineSvc_EndGame(t *testing.T) {

}

func withGameEngineSetup(t gomock.TestReporter, testFunc func(b gameEngineTestBundle)) {
	mockCtrl := gomock.NewController(t)
	mockDB := repo.NewMockDB(mockCtrl)
	gameEngineSvc := NewGameEngineSvc(mockDB)

	testFunc(gameEngineTestBundle{
		gameEngineSvc: gameEngineSvc,
		mockDB:        mockDB,
	})
}

type gameEngineTestBundle struct {
	gameEngineSvc *GameEngineSvc
	mockDB        *repo.MockDB
}
