package util

import (
	"testing"

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
}

func TestGetKeysFromMap(t *testing.T) {
	t.Run("Get keys from map", func(t *testing.T) {
		m := map[string]int{"1": 0, "2": 0, "3": 0, "4": 0}
		want := []string{"1", "2", "3", "4"}
		assert.ElementsMatch(t, want, GetKeysFromMap(m))
	})
}
