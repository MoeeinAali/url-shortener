package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"url-shortener/internal/application/command"
	"url-shortener/internal/application/query"
	"url-shortener/internal/domain/link"
	"url-shortener/internal/domain/shared"
	transport "url-shortener/internal/interfaces/http"
)

// fakeRead is an in-memory read model.
type fakeRead struct {
	urls   map[string]string
	clicks map[string]int64
}

func (f *fakeRead) LongURL(_ context.Context, s string) (string, error) {
	u, ok := f.urls[s]
	if !ok {
		return "", link.ErrLinkNotFound
	}
	return u, nil
}

func (f *fakeRead) Clicks(_ context.Context, s string) (int64, error) {
	if _, ok := f.urls[s]; !ok {
		return 0, link.ErrLinkNotFound
	}
	return f.clicks[s], nil
}

type fakeOutbox struct{ appended int }

func (f *fakeOutbox) Append(_ context.Context, events ...shared.DomainEvent) error {
	f.appended += len(events)
	return nil
}

func buildRouter(read *fakeRead, outbox *fakeOutbox) *gin.Engine {
	gin.SetMode(gin.TestMode)
	log := zap.NewNop()

	recordClick := command.NewRecordClickHandler(outbox)
	redirect := query.NewRedirectHandler(read, recordClick, log)
	stats := query.NewGetStatsHandler(read)
	qry := transport.NewQueryHandler(redirect, stats)

	// Command handlers are registered but not exercised here (nil repo is fine
	// because the create/disable routes are not called).
	cmd := transport.NewCommandHandler(
		command.NewCreateLinkHandler(nil, nil, log),
		command.NewDisableLinkHandler(nil, log),
		"http://localhost:8080",
	)
	return transport.NewRouter(cmd, qry)
}

func TestRouter_RegistersWithoutConflict_AndHealthz(t *testing.T) {
	r := buildRouter(&fakeRead{urls: map[string]string{}, clicks: map[string]int64{}}, &fakeOutbox{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("healthz: expected 200, got %d", w.Code)
	}
}

func TestRouter_Redirect(t *testing.T) {
	read := &fakeRead{
		urls:   map[string]string{"AAAAAAA": "https://example.com"},
		clicks: map[string]int64{},
	}
	outbox := &fakeOutbox{}
	r := buildRouter(read, outbox)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/AAAAAAA", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("redirect: expected 302, got %d", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "https://example.com" {
		t.Errorf("redirect: unexpected Location %q", loc)
	}
	if outbox.appended != 1 {
		t.Errorf("redirect: expected 1 click recorded, got %d", outbox.appended)
	}
}

func TestRouter_RedirectNotFound(t *testing.T) {
	r := buildRouter(&fakeRead{urls: map[string]string{}, clicks: map[string]int64{}}, &fakeOutbox{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/UNKNOWN", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("redirect: expected 404, got %d", w.Code)
	}
}

func TestRouter_Stats(t *testing.T) {
	read := &fakeRead{
		urls:   map[string]string{"BBBBBBB": "https://example.com"},
		clicks: map[string]int64{"BBBBBBB": 42},
	}
	r := buildRouter(read, &fakeOutbox{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/links/BBBBBBB/stats", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("stats: expected 200, got %d", w.Code)
	}
}
