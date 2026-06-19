package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"url-shortener/internal/application/command"
)

// CommandHandler exposes the write-side (POST) endpoints.
type CommandHandler struct {
	createLink  *command.CreateLinkHandler
	disableLink *command.DisableLinkHandler
	baseURL     string
}

// NewCommandHandler wires the handler.
func NewCommandHandler(create *command.CreateLinkHandler, disable *command.DisableLinkHandler, baseURL string) *CommandHandler {
	return &CommandHandler{createLink: create, disableLink: disable, baseURL: baseURL}
}

type createLinkRequest struct {
	URL string `json:"url" binding:"required"`
}

// CreateLink handles POST /links.
func (h *CommandHandler) CreateLink(c *gin.Context) {
	var req createLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request body must contain a non-empty 'url'"})
		return
	}

	res, err := h.createLink.Handle(c.Request.Context(), command.CreateLink{URL: req.URL})
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"short_code": res.ShortCode,
		"long_url":   res.LongURL,
		"short_url":  h.baseURL + "/" + res.ShortCode,
	})
}

// DisableLink handles POST /links/:short/disable.
func (h *CommandHandler) DisableLink(c *gin.Context) {
	short := c.Param("short")
	if err := h.disableLink.Handle(c.Request.Context(), command.DisableLink{ShortCode: short}); err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"short_code": short, "status": "disabled"})
}
