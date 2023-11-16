package util

import (
	"testing"

	"BattlesnakeReptarium/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	t.Run("Int slice contains target", func(t *testing.T) {
		slice := []int{1, 2, 3}
		target := 2
		assert.True(t, Contains(slice, target))
	})

	t.Run("Int slice does not contain target", func(t *testing.T) {
		slice := []int{1, 2, 3}
		target := 4
		assert.False(t, Contains(slice, target))
	})

	t.Run("Coord slice contains target", func(t *testing.T) {
		slice := []model.Coord{{X: 1, Y: 1}, {X: 2, Y: 2}}
		target := model.Coord{X: 1, Y: 1}
		assert.True(t, Contains(slice, target))
	})

	t.Run("Coord slice does not contain target", func(t *testing.T) {
		slice := []model.Coord{{X: 1, Y: 1}, {X: 2, Y: 2}}
		target := model.Coord{X: 3, Y: 3}
		assert.False(t, Contains(slice, target))
	})
}
