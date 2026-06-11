package model

import (
	"fmt"
	"net/textproto"
	"strings"

	"gon/internal/core/domain"
)

type BodyModel struct {
	JSON        any               `yaml:"json,omitempty"`
	Raw         string            `yaml:"raw,omitempty"`
	ContentType string            `yaml:"contentType,omitempty"`
	Form        map[string]string `yaml:"form,omitempty"`
}

type RequestModel struct {
	Name        string            `yaml:"name,omitempty"`
	Description string            `yaml:"description,omitempty"`
	Method      string            `yaml:"method,omitempty"`
	URL         string            `yaml:"url,omitempty"`
	Headers     map[string]string `yaml:"headers,omitempty"`
	Query       map[string]string `yaml:"query,omitempty"`
	Body        *BodyModel        `yaml:"body,omitempty"`
}

func NewRequestModelFromDomain(request domain.Request) *RequestModel {
	headers := make(map[string]string, len(request.Headers))
	for key, values := range request.Headers {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	query := make(map[string]string, len(request.Query))
	for key, values := range request.Query {
		if len(values) > 0 {
			query[key] = values[0]
		}
	}

	m := &RequestModel{
		Name:        request.Name,
		Description: request.Description,
		Method:      request.Method,
		URL:         request.URL,
		Headers:     headers,
		Query:       query,
	}
	switch request.Body.Kind {
	case domain.BodyJSON:
		m.Body = &BodyModel{JSON: request.Body.JSON}
	case domain.BodyRaw:
		m.Body = &BodyModel{Raw: request.Body.Raw, ContentType: request.Body.ContentType}
	case domain.BodyForm:
		m.Body = &BodyModel{Form: request.Body.Form}
	}
	return m
}

func (m *RequestModel) ToDomain() (*domain.Request, error) {
	body, err := m.Body.toDomain()
	if err != nil {
		return nil, err
	}

	headers := make(map[string][]string, len(m.Headers))
	for key, value := range m.Headers {
		canonical := textproto.CanonicalMIMEHeaderKey(key)
		headers[canonical] = []string{value}
	}
	query := make(map[string][]string, len(m.Query))
	for key, value := range m.Query {
		query[key] = []string{value}
	}

	return &domain.Request{
		Name:        m.Name,
		Description: m.Description,
		Method:      strings.ToUpper(m.Method),
		URL:         m.URL,
		Headers:     headers,
		Query:       query,
		Body:        body,
	}, nil
}

func (b *BodyModel) toDomain() (domain.RequestBody, error) {
	if b == nil {
		return domain.RequestBody{Kind: domain.BodyNone}, nil
	}

	set := 0
	if b.JSON != nil {
		set++
	}
	if b.Raw != "" {
		set++
	}
	if len(b.Form) > 0 {
		set++
	}
	if set > 1 {
		return domain.RequestBody{}, fmt.Errorf("invalid body: only one body kind (json, raw, form) may be set")
	}

	switch {
	case b.JSON != nil:
		return domain.RequestBody{Kind: domain.BodyJSON, JSON: b.JSON}, nil
	case b.Raw != "":
		return domain.RequestBody{Kind: domain.BodyRaw, Raw: b.Raw, ContentType: b.ContentType}, nil
	case len(b.Form) > 0:
		return domain.RequestBody{Kind: domain.BodyForm, Form: b.Form}, nil
	default:
		return domain.RequestBody{Kind: domain.BodyNone}, nil
	}
}
