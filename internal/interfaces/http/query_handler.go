package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"url-shortener/internal/application/query"
)

// QueryHandler exposes the read-side (GET) endpoints.
type QueryHandler struct {
	redirect *query.RedirectHandler
	getStats *query.GetStatsHandler
}

// NewQueryHandler wires the handler.
func NewQueryHandler(redirect *query.RedirectHandler, getStats *query.GetStatsHandler) *QueryHandler {
	return &QueryHandler{redirect: redirect, getStats: getStats}
}

// Redirect handles GET /:short — resolves and 302-redirects, recording a click.
func (h *QueryHandler) Redirect(c *gin.Context) {
	short := c.Param("short")
	longURL, err := h.redirect.Handle(c.Request.Context(), query.Redirect{ShortCode: short})
	if err != nil {
		writeError(c, err)
		return
	}
	c.Redirect(http.StatusFound, longURL)
}

// Stats handles GET /links/:short/stats.
func (h *QueryHandler) Stats(c *gin.Context) {
	short := c.Param("short")
	res, err := h.getStats.Handle(c.Request.Context(), query.GetStats{ShortCode: short})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"short_code": res.ShortCode, "clicks": res.Clicks})
}
