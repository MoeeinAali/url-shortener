package link

import "errors"

// Domain errors. These are part of the ubiquitous language and are translated
// into transport-level responses (e.g. HTTP status codes) at the edges.
var (
	// ErrInvalidURL means the provided long URL is not a valid absolute http(s) URL.
	ErrInvalidURL = errors.New("invalid url: must be an absolute http(s) URL")
	// ErrInvalidShortCode means a short code does not satisfy the format invariant.
	ErrInvalidShortCode = errors.New("invalid short code")
	// ErrInvalidLinkID means a link id could not be parsed as a UUID.
	ErrInvalidLinkID = errors.New("invalid link id")
	// ErrLinkNotFound means no link exists for the given identity.
	ErrLinkNotFound = errors.New("link not found")
	// ErrLinkAlreadyDisabled is the invariant violated when disabling a disabled link.
	ErrLinkAlreadyDisabled = errors.New("link is already disabled")
)
