package link_test

import (
	"errors"
	"testing"
	"time"

	"url-shortener/internal/domain/link"
)

func l0time() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }

func mustURL(t *testing.T, raw string) link.URL {
	t.Helper()
	u, err := link.NewURL(raw)
	if err != nil {
		t.Fatalf("NewURL(%q) unexpected error: %v", raw, err)
	}
	return u
}

func mustCode(t *testing.T) link.ShortCode {
	t.Helper()
	c, err := link.GenerateShortCode()
	if err != nil {
		t.Fatalf("GenerateShortCode unexpected error: %v", err)
	}
	return c
}

func TestNewURL_Validation(t *testing.T) {
	valid := []string{
		"https://example.com",
		"http://example.com/path?q=1",
		"https://sub.domain.io/a/b/c",
	}
	for _, v := range valid {
		if _, err := link.NewURL(v); err != nil {
			t.Errorf("expected %q to be valid, got %v", v, err)
		}
	}

	invalid := []string{
		"",
		"   ",
		"ftp://example.com",
		"example.com",         // no scheme
		"https://",            // no host
		"javascript:alert(1)", // wrong scheme
	}
	for _, v := range invalid {
		if _, err := link.NewURL(v); !errors.Is(err, link.ErrInvalidURL) {
			t.Errorf("expected %q to be invalid with ErrInvalidURL, got %v", v, err)
		}
	}
}

func TestNewShortCode_Validation(t *testing.T) {
	if _, err := link.NewShortCode("short"); !errors.Is(err, link.ErrInvalidShortCode) {
		t.Errorf("expected too-short code to be invalid, got %v", err)
	}
	if _, err := link.NewShortCode("with-dash"); !errors.Is(err, link.ErrInvalidShortCode) {
		t.Errorf("expected non-base62 code to be invalid, got %v", err)
	}
	// A generated code must round-trip through validation.
	gen := mustCode(t)
	if _, err := link.NewShortCode(gen.String()); err != nil {
		t.Errorf("generated code %q failed validation: %v", gen.String(), err)
	}
}

func TestGenerateShortCode_Uniqueness(t *testing.T) {
	seen := make(map[string]struct{}, 1000)
	for i := 0; i < 1000; i++ {
		c := mustCode(t)
		if len(c.String()) != 7 {
			t.Fatalf("expected length 7, got %d", len(c.String()))
		}
		if _, dup := seen[c.String()]; dup {
			t.Fatalf("unexpected duplicate short code: %s", c.String())
		}
		seen[c.String()] = struct{}{}
	}
}

func TestNewLink_RaisesCreatedEvent(t *testing.T) {
	code := mustCode(t)
	l, err := link.NewLink(code, mustURL(t, "https://example.com"))
	if err != nil {
		t.Fatalf("NewLink unexpected error: %v", err)
	}
	if !l.Status().IsActive() {
		t.Errorf("expected new link to be active, got %s", l.Status())
	}

	events := l.PullEvents()
	if len(events) != 1 {
		t.Fatalf("expected exactly 1 event, got %d", len(events))
	}
	created, ok := events[0].(link.LinkCreated)
	if !ok {
		t.Fatalf("expected LinkCreated, got %T", events[0])
	}
	if created.ShortCode != code.String() {
		t.Errorf("event short code mismatch: %s != %s", created.ShortCode, code.String())
	}
	if created.EventType() != link.EventTypeLinkCreated {
		t.Errorf("unexpected event type %s", created.EventType())
	}

	// Events must be drained after a pull.
	if rest := l.PullEvents(); len(rest) != 0 {
		t.Errorf("expected events to be cleared after pull, got %d", len(rest))
	}
}

func TestDisable_RaisesEventAndEnforcesInvariant(t *testing.T) {
	l, err := link.NewLink(mustCode(t), mustURL(t, "https://example.com"))
	if err != nil {
		t.Fatalf("NewLink unexpected error: %v", err)
	}
	_ = l.PullEvents() // drop the creation event

	if err := l.Disable(); err != nil {
		t.Fatalf("first Disable should succeed, got %v", err)
	}
	if !l.Status().IsDisabled() {
		t.Errorf("expected link to be disabled")
	}
	events := l.PullEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 disabled event, got %d", len(events))
	}
	if _, ok := events[0].(link.LinkDisabled); !ok {
		t.Fatalf("expected LinkDisabled, got %T", events[0])
	}

	// Invariant: disabling again is rejected and raises no event.
	if err := l.Disable(); !errors.Is(err, link.ErrLinkAlreadyDisabled) {
		t.Errorf("expected ErrLinkAlreadyDisabled, got %v", err)
	}
	if rest := l.PullEvents(); len(rest) != 0 {
		t.Errorf("expected no events after rejected disable, got %d", len(rest))
	}
}

func TestReconstitute_RaisesNoEvents(t *testing.T) {
	id := link.NewLinkID()
	l := link.Reconstitute(id, mustCode(t), mustURL(t, "https://example.com"), link.StatusActive, l0time(), 3)
	if l.Version() != 3 {
		t.Errorf("expected version 3, got %d", l.Version())
	}
	if len(l.PullEvents()) != 0 {
		t.Errorf("reconstituted aggregate must not raise events")
	}
}
