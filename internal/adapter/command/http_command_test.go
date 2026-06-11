package command

import (
	"context"
	"testing"

	"gon/internal/core/enums"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestParseHeaders(t *testing.T) {
	t.Run("valid header", func(t *testing.T) {
		got, err := parseHeaders([]string{"Accept: application/json"})
		require.NoError(t, err)
		assert.Equal(t, map[string][]string{"Accept": {"application/json"}}, got)
	})

	t.Run("canonicalizes the key", func(t *testing.T) {
		got, err := parseHeaders([]string{"content-type: text/plain"})
		require.NoError(t, err)
		assert.Equal(t, map[string][]string{"Content-Type": {"text/plain"}}, got)
	})

	t.Run("trims whitespace around value", func(t *testing.T) {
		got, err := parseHeaders([]string{"X-Token:   abc  "})
		require.NoError(t, err)
		assert.Equal(t, map[string][]string{"X-Token": {"abc"}}, got)
	})

	t.Run("repeated headers append", func(t *testing.T) {
		got, err := parseHeaders([]string{"X-Tag: a", "X-Tag: b"})
		require.NoError(t, err)
		assert.Equal(t, map[string][]string{"X-Tag": {"a", "b"}}, got)
	})

	t.Run("value containing colon is preserved", func(t *testing.T) {
		got, err := parseHeaders([]string{"X-URL: http://example.com"})
		require.NoError(t, err)
		assert.Equal(t, map[string][]string{"X-Url": {"http://example.com"}}, got)
	})

	t.Run("missing colon is an error", func(t *testing.T) {
		_, err := parseHeaders([]string{"NoColon"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Key: Value")
	})

	t.Run("empty key is an error", func(t *testing.T) {
		_, err := parseHeaders([]string{": value"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty key")
	})

	t.Run("empty slice yields empty map", func(t *testing.T) {
		got, err := parseHeaders(nil)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestParseQuery(t *testing.T) {
	t.Run("valid query", func(t *testing.T) {
		got, err := parseQuery([]string{"page=1"})
		require.NoError(t, err)
		assert.Equal(t, map[string][]string{"page": {"1"}}, got)
	})

	t.Run("repeated keys append", func(t *testing.T) {
		got, err := parseQuery([]string{"tag=a", "tag=b"})
		require.NoError(t, err)
		assert.Equal(t, map[string][]string{"tag": {"a", "b"}}, got)
	})

	t.Run("value containing equals is preserved", func(t *testing.T) {
		got, err := parseQuery([]string{"filter=a=b"})
		require.NoError(t, err)
		assert.Equal(t, map[string][]string{"filter": {"a=b"}}, got)
	})

	t.Run("missing equals is an error", func(t *testing.T) {
		_, err := parseQuery([]string{"noequals"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Key=Value")
	})

	t.Run("empty key is an error", func(t *testing.T) {
		_, err := parseQuery([]string{"=value"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty key")
	})
}

// runResolveMode builds a minimal cli.Command with the mode flags, runs it with
// the given args, and returns the DisplayMode that resolveMode computes.
func runResolveMode(t *testing.T, args ...string) enums.DisplayMode {
	t.Helper()
	var got enums.DisplayMode
	cmd := &cli.Command{
		Name: "test",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "minimal"},
			&cli.BoolFlag{Name: "normal"},
			&cli.BoolFlag{Name: "full"},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			got = resolveMode(c)
			return nil
		},
	}
	require.NoError(t, cmd.Run(context.Background(), append([]string{"test"}, args...)))
	return got
}

func TestResolveMode(t *testing.T) {
	t.Run("defaults to normal", func(t *testing.T) {
		assert.Equal(t, enums.DisplayModeNormal, runResolveMode(t))
	})

	t.Run("minimal flag", func(t *testing.T) {
		assert.Equal(t, enums.DisplayModeMinimal, runResolveMode(t, "--minimal"))
	})

	t.Run("normal flag", func(t *testing.T) {
		assert.Equal(t, enums.DisplayModeNormal, runResolveMode(t, "--normal"))
	})

	t.Run("full flag", func(t *testing.T) {
		assert.Equal(t, enums.DisplayModeFull, runResolveMode(t, "--full"))
	})

	t.Run("minimal takes precedence over full", func(t *testing.T) {
		assert.Equal(t, enums.DisplayModeMinimal, runResolveMode(t, "--minimal", "--full"))
	})
}
