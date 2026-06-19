package link

// LinkStatus is a value object capturing the lifecycle state of a link.
type LinkStatus string

const (
	// StatusActive means the link resolves and records clicks.
	StatusActive LinkStatus = "active"
	// StatusDisabled means the link no longer resolves.
	StatusDisabled LinkStatus = "disabled"
)

// IsActive reports whether the link is currently resolvable.
func (s LinkStatus) IsActive() bool { return s == StatusActive }

// IsDisabled reports whether the link has been disabled.
func (s LinkStatus) IsDisabled() bool { return s == StatusDisabled }
