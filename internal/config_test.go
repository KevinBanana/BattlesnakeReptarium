package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetConfig(t *testing.T) {
	t.Run("GetConfig_Environment not set", func(t *testing.T) {
		c := Config{}
		err := c.GetConfig()
		assert.Error(t, err)
	})

	t.Run("GetConfig_Happy path", func(t *testing.T) {
		c := Config{Environment: "local"}
		err := c.GetConfig()
		assert.NoError(t, err)
	})
}
