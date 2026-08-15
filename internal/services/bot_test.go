package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"BattlesnakeReptarium/internal/model"
)

func TestPlaysGamemode(t *testing.T) {
	tests := map[string]struct {
		gamemodes []string
		entered   string
		want      bool
	}{
		"Declared gamemode matches":        {[]string{model.GamemodeConstrictor}, model.GamemodeConstrictor, true},
		"Declared gamemode does not match": {[]string{model.GamemodeConstrictor}, model.GamemodeStandard, false},
		"Matches one of several":           {[]string{model.GamemodeRoyale, model.GamemodeDuel}, model.GamemodeDuel, true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bot := NewMockBot(gomock.NewController(t))
			bot.EXPECT().Gamemodes().Return(test.gamemodes)
			assert.Equal(t, test.want, PlaysGamemode(bot, test.entered))
		})
	}
}
