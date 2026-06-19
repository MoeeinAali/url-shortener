package link

import "github.com/google/uuid"

// LinkID is the value object for an aggregate's identity.
type LinkID struct {
	value uuid.UUID
}

// NewLinkID mints a fresh identity.
func NewLinkID() LinkID { return LinkID{value: uuid.New()} }

// LinkIDFromUUID wraps an existing UUID (e.g. when reconstituting from storage).
func LinkIDFromUUID(id uuid.UUID) LinkID { return LinkID{value: id} }

// LinkIDFromString parses an identity from its string form.
func LinkIDFromString(s string) (LinkID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return LinkID{}, ErrInvalidLinkID
	}
	return LinkID{value: id}, nil
}

// UUID exposes the underlying UUID for persistence and events.
func (id LinkID) UUID() uuid.UUID { return id.value }

// String returns the canonical string form.
func (id LinkID) String() string { return id.value.String() }
