package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"BattlesnakeReptarium/internal/model"
	"BattlesnakeReptarium/internal/services"

	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testBot = "test_bot"

func TestGame_Info(t *testing.T) {
	t.Run("Serves the bot's own customizations", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			b.mockBot.EXPECT().Customizations().Return(model.SnakeCustomizations{
				Head:  "shades",
				Tail:  "bolt",
				Color: "#00e5ff",
			}).Times(1)

			c := b.controller()
			c.WithBot(c.Info)(b.w, botRequest(http.MethodGet, testBot, "", nil))
			assert.Equal(t, http.StatusOK, b.w.Code)

			var body map[string]string
			require.NoError(t, json.Unmarshal(b.w.Body.Bytes(), &body))
			assert.Equal(t, "shades", body["head"])
			assert.Equal(t, "bolt", body["tail"])
			assert.Equal(t, "#00e5ff", body["color"])
			assert.Equal(t, "1", body["apiversion"])
		})
	})

	t.Run("Unknown bot", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			c := b.controller()
			c.WithBot(c.Info)(b.w, botRequest(http.MethodGet, "not_a_bot", "", nil))
			assert.Equal(t, http.StatusNotFound, b.w.Code)
		})
	})
}

func TestGame_Health(t *testing.T) {
	t.Run("Lists every bot it serves", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			controller := NewGameController(map[string]services.Bot{
				"zzz_bot": b.mockBot,
				testBot:   b.mockBot,
			}, b.mockGameEngine)

			controller.Health(b.w, httptest.NewRequest(http.MethodGet, "/", nil))
			assert.Equal(t, http.StatusOK, b.w.Code)

			var body struct {
				Status string   `json:"status"`
				Bots   []string `json:"bots"`
			}
			require.NoError(t, json.Unmarshal(b.w.Body.Bytes(), &body))
			assert.Equal(t, "ok", body.Status)
			assert.Equal(t, []string{testBot, "zzz_bot"}, body.Bots)
		})
	})
}

func TestGame_StartGame(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			b.mockGameEngine.EXPECT().StartGame(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

			c := b.controller()
			c.WithGame(c.StartGame)(b.w, botRequest(http.MethodPost, testBot, "/start", defaultRequest))
			assert.Equal(t, http.StatusOK, b.w.Code)
		})
	})

	t.Run("Bad request", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			c := b.controller()
			c.WithGame(c.StartGame)(b.w, botRequest(http.MethodPost, testBot, "/start", "bad request"))
			assert.Equal(t, http.StatusBadRequest, b.w.Code)
		})
	})

	t.Run("Unknown bot", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			b.mockGameEngine.EXPECT().StartGame(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			c := b.controller()
			c.WithGame(c.StartGame)(b.w, botRequest(http.MethodPost, "not_a_bot", "/start", defaultRequest))
			assert.Equal(t, http.StatusNotFound, b.w.Code)
		})
	})

	t.Run("Internal server error", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			b.mockGameEngine.EXPECT().StartGame(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("error")).Times(1)

			c := b.controller()
			c.WithGame(c.StartGame)(b.w, botRequest(http.MethodPost, testBot, "/start", defaultRequest))
			assert.Equal(t, http.StatusInternalServerError, b.w.Code)
		})
	})
}

func TestGame_CalculateMove(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			b.mockBot.EXPECT().CalculateMove(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				&model.SnakeAction{}, nil).Times(1)

			c := b.controller()
			c.WithGame(c.CalculateMove)(b.w, botRequest(http.MethodPost, testBot, "/move", defaultRequest))
			assert.Equal(t, http.StatusOK, b.w.Code)
		})
	})

	t.Run("Each bot gets its own moves", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			otherBot := services.NewMockBot(gomock.NewController(t))
			otherBot.EXPECT().CalculateMove(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				&model.SnakeAction{}, nil).Times(1)
			b.mockBot.EXPECT().CalculateMove(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			c := NewGameController(map[string]services.Bot{
				testBot:     b.mockBot,
				"other_bot": otherBot,
			}, b.mockGameEngine)

			c.WithGame(c.CalculateMove)(b.w, botRequest(http.MethodPost, "other_bot", "/move", defaultRequest))
			assert.Equal(t, http.StatusOK, b.w.Code)
		})
	})

	t.Run("Unknown bot", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			b.mockBot.EXPECT().CalculateMove(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			c := b.controller()
			c.WithGame(c.CalculateMove)(b.w, botRequest(http.MethodPost, "not_a_bot", "/move", defaultRequest))
			assert.Equal(t, http.StatusNotFound, b.w.Code)
		})
	})

	t.Run("Bad request", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			b.mockBot.EXPECT().CalculateMove(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			c := b.controller()
			c.WithGame(c.CalculateMove)(b.w, botRequest(http.MethodPost, testBot, "/move", "bad request"))
			assert.Equal(t, http.StatusBadRequest, b.w.Code)
		})
	})

	t.Run("Internal server error", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			b.mockBot.EXPECT().CalculateMove(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				nil, errors.New("error")).Times(1)

			c := b.controller()
			c.WithGame(c.CalculateMove)(b.w, botRequest(http.MethodPost, testBot, "/move", defaultRequest))
			assert.Equal(t, http.StatusInternalServerError, b.w.Code)
		})
	})
}

// botRequest builds a request the way the router hands it over, with {bot} already resolved.
func botRequest(method, bot, action string, body any) *http.Request {
	buf, _ := json.Marshal(body)
	r := httptest.NewRequest(method, "/"+bot+action, bytes.NewReader(buf))
	r.SetPathValue("bot", bot)
	return r
}

func withGameSetup(t gomock.TestReporter, testFunc func(testBundle gameTestBundle)) {
	mockCtrl := gomock.NewController(t)

	mockBot := services.NewMockBot(mockCtrl)
	mockBot.EXPECT().Gamemodes().Return(nil).AnyTimes()

	testFunc(gameTestBundle{
		w:              httptest.NewRecorder(),
		mockBot:        mockBot,
		mockGameEngine: services.NewMockGameEngineService(mockCtrl),
	})
}

type gameTestBundle struct {
	w              *httptest.ResponseRecorder
	mockBot        *services.MockBot
	mockGameEngine *services.MockGameEngineService
}

func (b gameTestBundle) controller() GameController {
	return NewGameController(map[string]services.Bot{testBot: b.mockBot}, b.mockGameEngine)
}
