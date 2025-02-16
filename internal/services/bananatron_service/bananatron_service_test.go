package bananatron_service

import (
	"reflect"
	"testing"

	"BattlesnakeReptarium/internal/model"
)

func TestDetermineSnakeAction(t *testing.T) {
	tests := []struct {
		name            string
		weightedOptions map[model.Direction]float64
		want            model.Direction
	}{
		{"One positive weight", map[model.Direction]float64{model.UP: 1, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0}, model.UP},
		{"Three negative weights", map[model.Direction]float64{model.UP: -1, model.LEFT: -1, model.DOWN: -1, model.RIGHT: 0}, model.RIGHT},
		{"Two positive weights", map[model.Direction]float64{model.UP: 1, model.LEFT: 0, model.DOWN: 2, model.RIGHT: 0}, model.DOWN},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := determineSnakeAction(tt.weightedOptions).Move; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("determineSnakeAction() = %v, want %v", got, tt.want)
			}
		})
	}
}
