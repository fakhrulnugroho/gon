package domain

import (
	"encoding/json"
	"net/textproto"
	"net/url"

	"gon/internal/core/payload"
)

type BodyKind int

const (
	BodyNone BodyKind = iota
	BodyJSON
	BodyRaw
	BodyForm
)

type RequestBody struct {
	Kind        BodyKind
	JSON        any
	Raw         string
	ContentType string
	Form        map[string]string
}

// Encode returns the wire bytes for the body and the content type that should
// be used when the request does not already specify one.
func (b RequestBody) Encode() ([]byte, string, error) {
	switch b.Kind {
	case BodyJSON:
		data, err := json.Marshal(b.JSON)
		if err != nil {
			return nil, "", err
		}
		return data, "application/json", nil
	case BodyRaw:
		return []byte(b.Raw), b.ContentType, nil
	case BodyForm:
		values := url.Values{}
		for key, value := range b.Form {
			values.Set(key, value)
		}
		return []byte(values.Encode()), "application/x-www-form-urlencoded", nil
	default:
		return nil, "", nil
	}
}

type Request struct {
	Name        string
	Description string
	Method      string
	URL         string
	Headers     map[string][]string
	Query       map[string][]string
	Body        RequestBody
}

// ToInput builds an HttpExecuteInput from the request, encoding the body and
// auto-setting Content-Type when the request does not already provide one.
func (r Request) ToInput() (*payload.HttpExecuteInput, error) {
	data, contentType, err := r.Body.Encode()
	if err != nil {
		return nil, err
	}

	headers := make(map[string][]string, len(r.Headers))
	for key, values := range r.Headers {
		headers[textproto.CanonicalMIMEHeaderKey(key)] = append([]string(nil), values...)
	}
	query := make(map[string][]string, len(r.Query))
	for key, values := range r.Query {
		query[key] = append([]string(nil), values...)
	}

	if contentType != "" {
		if _, ok := headers["Content-Type"]; !ok {
			headers["Content-Type"] = []string{contentType}
		}
	}

	return &payload.HttpExecuteInput{
		Method:  r.Method,
		URL:     r.URL,
		Headers: headers,
		Query:   query,
		Body:    data,
	}, nil
}
