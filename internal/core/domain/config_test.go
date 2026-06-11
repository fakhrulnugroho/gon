package domain

import (
	"testing"

	"gon/internal/core/payload"

	"github.com/stretchr/testify/assert"
)

func TestConfigApplyDefaults(t *testing.T) {
	t.Run("injects header and query when absent", func(t *testing.T) {
		c := Config{
			Headers: map[string]string{"Authorization": "Bearer token"},
			Query:   map[string]string{"debug": "1"},
		}
		input := &payload.HttpExecuteInput{}

		c.ApplyDefaults(input)

		assert.Equal(t, []string{"Bearer token"}, input.Headers["Authorization"])
		assert.Equal(t, []string{"1"}, input.Query["debug"])
	})

	t.Run("existing key wins and is not duplicated", func(t *testing.T) {
		c := Config{Headers: map[string]string{"authorization": "Bearer default"}}
		input := &payload.HttpExecuteInput{
			Headers: map[string][]string{"Authorization": {"Bearer override"}},
		}

		c.ApplyDefaults(input)

		assert.Equal(t, []string{"Bearer override"}, input.Headers["Authorization"])
	})

	t.Run("empty config is a no-op", func(t *testing.T) {
		c := Config{}
		input := &payload.HttpExecuteInput{}

		c.ApplyDefaults(input)

		assert.Empty(t, input.Headers)
		assert.Empty(t, input.Query)
	})
}
