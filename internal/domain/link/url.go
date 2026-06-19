package link

import (
	"net/url"
	"strings"
)

// URL is a value object representing a validated, absolute destination URL.
// It is immutable: once constructed it is guaranteed to be well-formed.
type URL struct {
	value string
}

// NewURL validates and normalizes a raw URL string. It enforces the invariant
// that a destination must be an absolute http(s) URL with a host.
func NewURL(raw string) (URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return URL{}, ErrInvalidURL
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return URL{}, ErrInvalidURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return URL{}, ErrInvalidURL
	}
	if parsed.Host == "" {
		return URL{}, ErrInvalidURL
	}

	return URL{value: parsed.String()}, nil
}

// String returns the normalized URL.
func (u URL) String() string { return u.value }

// IsZero reports whether the value object is the zero value.
func (u URL) IsZero() bool { return u.value == "" }
