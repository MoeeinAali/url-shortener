package command_test

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"

	"url-shortener/internal/application/command"
	"url-shortener/internal/domain/link"
	"url-shortener/internal/domain/shared"
)

// --- in-memory fakes -------------------------------------------------------

type fakeRepo struct {
	byCode      map[string]*link.Link
	savedEvents []shared.DomainEvent
	failSaveN   int // fail the first N saves to exercise retry
}

func newFakeRepo() *fakeRepo { return &fakeRepo{byCode: map[string]*link.Link{}} }

func (r *fakeRepo) Save(_ context.Context, l *link.Link) error {
	if r.failSaveN > 0 {
		r.failSaveN--
		return errors.New("simulated save failure")
	}
	r.byCode[l.ShortCode().String()] = l
	r.savedEvents = append(r.savedEvents, l.PullEvents()...)
	return nil
}

func (r *fakeRepo) FindByShortCode(_ context.Context, code link.ShortCode) (*link.Link, error) {
	l, ok := r.byCode[code.String()]
	if !ok {
		return nil, link.ErrLinkNotFound
	}
	return l, nil
}

func (r *fakeRepo) ExistsByShortCode(_ context.Context, code link.ShortCode) (bool, error) {
	_, ok := r.byCode[code.String()]
	return ok, nil
}

// seqGen yields a fixed sequence of short codes.
type seqGen struct {
	codes []string
	i     int
}

func (g *seqGen) Generate() (link.ShortCode, error) {
	c := g.codes[g.i%len(g.codes)]
	g.i++
	return link.NewShortCode(c)
}

// --- tests -----------------------------------------------------------------

func TestCreateLink_PersistsAggregateAndEvent(t *testing.T) {
	repo := newFakeRepo()
	gen := &seqGen{codes: []string{"AAAAAAA"}}
	h := command.NewCreateLinkHandler(repo, gen, zap.NewNop())

	res, err := h.Handle(context.Background(), command.CreateLink{URL: "https://example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ShortCode != "AAAAAAA" {
		t.Errorf("expected short AAAAAAA, got %s", res.ShortCode)
	}
	if len(repo.savedEvents) != 1 {
		t.Fatalf("expected 1 outboxed event, got %d", len(repo.savedEvents))
	}
	if _, ok := repo.savedEvents[0].(link.LinkCreated); !ok {
		t.Errorf("expected LinkCreated event, got %T", repo.savedEvents[0])
	}
}

func TestCreateLink_RejectsInvalidURL(t *testing.T) {
	repo := newFakeRepo()
	gen := &seqGen{codes: []string{"AAAAAAA"}}
	h := command.NewCreateLinkHandler(repo, gen, zap.NewNop())

	_, err := h.Handle(context.Background(), command.CreateLink{URL: "not-a-url"})
	if !errors.Is(err, link.ErrInvalidURL) {
		t.Fatalf("expected ErrInvalidURL, got %v", err)
	}
}

func TestCreateLink_SkipsCollidingCode(t *testing.T) {
	repo := newFakeRepo()
	// Pre-seed AAAAAAA so the generator's first code collides and it must retry.
	pre, _ := link.NewLink(mustCode2(t, "AAAAAAA"), mustURL2(t, "https://seed.com"))
	_ = repo.Save(context.Background(), pre)

	gen := &seqGen{codes: []string{"AAAAAAA", "BBBBBBB"}}
	h := command.NewCreateLinkHandler(repo, gen, zap.NewNop())

	res, err := h.Handle(context.Background(), command.CreateLink{URL: "https://example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ShortCode != "BBBBBBB" {
		t.Errorf("expected collision to be skipped, got %s", res.ShortCode)
	}
}

func TestDisableLink_Flow(t *testing.T) {
	repo := newFakeRepo()
	pre, _ := link.NewLink(mustCode2(t, "CCCCCCC"), mustURL2(t, "https://example.com"))
	_ = repo.Save(context.Background(), pre)
	repo.savedEvents = nil // ignore creation event

	h := command.NewDisableLinkHandler(repo, zap.NewNop())
	if err := h.Handle(context.Background(), command.DisableLink{ShortCode: "CCCCCCC"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.savedEvents) != 1 {
		t.Fatalf("expected 1 disabled event, got %d", len(repo.savedEvents))
	}
	if _, ok := repo.savedEvents[0].(link.LinkDisabled); !ok {
		t.Errorf("expected LinkDisabled, got %T", repo.savedEvents[0])
	}

	// Disabling again must surface the domain invariant.
	err := h.Handle(context.Background(), command.DisableLink{ShortCode: "CCCCCCC"})
	if !errors.Is(err, link.ErrLinkAlreadyDisabled) {
		t.Errorf("expected ErrLinkAlreadyDisabled, got %v", err)
	}
}

func TestDisableLink_NotFound(t *testing.T) {
	repo := newFakeRepo()
	h := command.NewDisableLinkHandler(repo, zap.NewNop())
	err := h.Handle(context.Background(), command.DisableLink{ShortCode: "ZZZZZZZ"})
	if !errors.Is(err, link.ErrLinkNotFound) {
		t.Errorf("expected ErrLinkNotFound, got %v", err)
	}
}

// helpers
func mustCode2(t *testing.T, s string) link.ShortCode {
	t.Helper()
	c, err := link.NewShortCode(s)
	if err != nil {
		t.Fatalf("bad code %q: %v", s, err)
	}
	return c
}

func mustURL2(t *testing.T, s string) link.URL {
	t.Helper()
	u, err := link.NewURL(s)
	if err != nil {
		t.Fatalf("bad url %q: %v", s, err)
	}
	return u
}
