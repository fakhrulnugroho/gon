package formatter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyPairFormatterSortsKeys(t *testing.T) {
	f := NewKeyPairFormatter()
	out := f.Format(map[string]string{
		"Zeta":  "last",
		"Alpha": "first",
		"Mu":    "middle",
	})

	// Strip ANSI color codes are present but keys must appear in sorted order.
	alpha := strings.Index(out, "Alpha")
	mu := strings.Index(out, "Mu")
	zeta := strings.Index(out, "Zeta")

	assert.NotEqual(t, -1, alpha)
	assert.Less(t, alpha, mu)
	assert.Less(t, mu, zeta)
	// Values are rendered alongside their keys.
	assert.Contains(t, out, "first")
	assert.Contains(t, out, "middle")
	assert.Contains(t, out, "last")
}

func TestKeyPairFormatterEmptyMap(t *testing.T) {
	f := NewKeyPairFormatter()
	assert.Equal(t, "", f.Format(map[string]string{}))
}

func TestJsonFormatterDelegatesToPrettyJSON(t *testing.T) {
	f := NewJsonFormatter()

	t.Run("empty input", func(t *testing.T) {
		assert.Equal(t, "", f.Format([]byte("")))
	})

	t.Run("valid JSON is indented", func(t *testing.T) {
		out := f.Format([]byte(`{"a":1}`))
		assert.Contains(t, out, "\n  ")
		assert.Contains(t, out, "a")
	})
}
