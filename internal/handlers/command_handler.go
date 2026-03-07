package handlers

import (
	"net/http"

	"url-shortener/internal/service"

	"github.com/gin-gonic/gin"
)

type CommandHandler struct {
	service *service.CommandService
}

func NewCommandHandler(s *service.CommandService) *CommandHandler {
	return &CommandHandler{s}
}

type createRequest struct {
	URL string `json:"url" binding:"required"`
}

func (h *CommandHandler) CreateLink(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	link, err := h.service.CreateLink(c.Request.Context(), req.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"short": link.ShortCode,
		"url":   link.LongURL,
	})
}

func (h *CommandHandler) DisableLink(c *gin.Context) {
	short := c.Param("short")
	err := h.service.DisableLink(c.Request.Context(), short)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "disabled",
	})
}
