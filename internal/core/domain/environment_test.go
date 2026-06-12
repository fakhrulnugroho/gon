package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvironmentSubstitute(t *testing.T) {
	env := &Environment{
		Name:      "dev",
		Variables: map[string]string{"token": "abc123", "user_id": "42"},
	}

	t.Run("replaces a known placeholder", func(t *testing.T) {
		assert.Equal(t, "Bearer abc123", env.Substitute("Bearer {{token}}"))
	})

	t.Run("tolerates inner whitespace", func(t *testing.T) {
		assert.Equal(t, "abc123", env.Substitute("{{ token }}"))
	})

	t.Run("replaces multiple placeholders", func(t *testing.T) {
		assert.Equal(t, "abc123/42", env.Substitute("{{token}}/{{user_id}}"))
	})

	t.Run("leaves unknown placeholders intact", func(t *testing.T) {
		assert.Equal(t, "{{missing}}", env.Substitute("{{missing}}"))
	})

	t.Run("nil environment returns input unchanged", func(t *testing.T) {
		var nilEnv *Environment
		assert.Equal(t, "{{token}}", nilEnv.Substitute("{{token}}"))
	})
}

func TestFindPlaceholders(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, FindPlaceholders("{{a}}/{{ b }}"))
	assert.Empty(t, FindPlaceholders("no placeholders here"))
}
