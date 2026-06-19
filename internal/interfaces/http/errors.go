// Package http is the inbound REST adapter (Gin). It translates HTTP requests
// into application commands/queries and maps domain errors back to status codes.
package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"url-shortener/internal/application/command"
	"url-shortener/internal/domain/link"
)

// writeError maps a domain/application error to an HTTP response. Keeping this in
// one place means the transport layer owns the protocol, not the core.
func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, link.ErrLinkNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
	case errors.Is(err, link.ErrLinkAlreadyDisabled):
		c.JSON(http.StatusConflict, gin.H{"error": "link is already disabled"})
	case errors.Is(err, link.ErrInvalidURL),
		errors.Is(err, link.ErrInvalidShortCode),
		errors.Is(err, link.ErrInvalidLinkID):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, command.ErrShortCodeExhausted):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "could not allocate a short code, try again"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
