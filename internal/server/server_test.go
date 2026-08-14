package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"BattlesnakeReptarium/internal/controllers"

	"github.com/stretchr/testify/assert"
)

func TestGame_Routes(t *testing.T) {
	bots := newBots()
	router := NewRouter(controllers.NewGameController(bots, nil), bots)

	cases := map[string]struct {
		method, path string
		wantStatus   int
	}{
		"service health":   {http.MethodGet, "/", http.StatusOK},
		"bot info":         {http.MethodGet, "/bananabot", http.StatusOK},
		"bot info slash":   {http.MethodGet, "/bananabot/", http.StatusOK},
		"unknown bot info": {http.MethodGet, "/not_a_bot", http.StatusNotFound},
		"bot move":         {http.MethodPost, "/bananabot/move", http.StatusOK},
		"unknown bot move": {http.MethodPost, "/not_a_bot/move", http.StatusNotFound},
		"unknown bot end":  {http.MethodPost, "/not_a_bot/end", http.StatusNotFound},
		"root move":        {http.MethodPost, "/move", http.StatusMethodNotAllowed},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, strings.NewReader("{}")))
			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}
