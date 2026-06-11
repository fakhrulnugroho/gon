package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestBodyEncode(t *testing.T) {
	t.Run("none yields no body", func(t *testing.T) {
		data, ct, err := RequestBody{Kind: BodyNone}.Encode()
		require.NoError(t, err)
		assert.Nil(t, data)
		assert.Equal(t, "", ct)
	})

	t.Run("json marshals and sets content type", func(t *testing.T) {
		data, ct, err := RequestBody{Kind: BodyJSON, JSON: map[string]any{"a": 1}}.Encode()
		require.NoError(t, err)
		assert.JSONEq(t, `{"a":1}`, string(data))
		assert.Equal(t, "application/json", ct)
	})

	t.Run("raw passes through with given content type", func(t *testing.T) {
		data, ct, err := RequestBody{Kind: BodyRaw, Raw: "hello", ContentType: "text/plain"}.Encode()
		require.NoError(t, err)
		assert.Equal(t, "hello", string(data))
		assert.Equal(t, "text/plain", ct)
	})

	t.Run("form url-encodes and sets content type", func(t *testing.T) {
		data, ct, err := RequestBody{Kind: BodyForm, Form: map[string]string{"a": "b"}}.Encode()
		require.NoError(t, err)
		assert.Equal(t, "a=b", string(data))
		assert.Equal(t, "application/x-www-form-urlencoded", ct)
	})
}

func TestRequestToInput(t *testing.T) {
	t.Run("builds input and auto-sets content type", func(t *testing.T) {
		r := Request{
			Method:  "POST",
			URL:     "/login",
			Headers: map[string][]string{"Accept": {"application/json"}},
			Query:   map[string][]string{"remember": {"true"}},
			Body:    RequestBody{Kind: BodyJSON, JSON: map[string]any{"email": "a@b.com"}},
		}

		input, err := r.ToInput()

		require.NoError(t, err)
		assert.Equal(t, "POST", input.Method)
		assert.Equal(t, "/login", input.URL)
		assert.Equal(t, []string{"application/json"}, input.Headers["Accept"])
		assert.Equal(t, []string{"true"}, input.Query["remember"])
		assert.JSONEq(t, `{"email":"a@b.com"}`, string(input.Body))
		assert.Equal(t, []string{"application/json"}, input.Headers["Content-Type"])
	})

	t.Run("does not override an explicit content type header", func(t *testing.T) {
		r := Request{
			Method:  "POST",
			Headers: map[string][]string{"Content-Type": {"application/vnd.api+json"}},
			Body:    RequestBody{Kind: BodyJSON, JSON: map[string]any{"a": 1}},
		}

		input, err := r.ToInput()

		require.NoError(t, err)
		assert.Equal(t, []string{"application/vnd.api+json"}, input.Headers["Content-Type"])
	})

	t.Run("initializes non-nil header and query maps", func(t *testing.T) {
		input, err := Request{Method: "GET", URL: "/ping"}.ToInput()
		require.NoError(t, err)
		assert.NotNil(t, input.Headers)
		assert.NotNil(t, input.Query)
	})
}
