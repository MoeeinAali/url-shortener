package handlers

import (
	"net/http"

	"url-shortener/internal/service"

	"github.com/gin-gonic/gin"
)

type QueryHandler struct {
	service *service.QueryService
}

func NewQueryHandler(s *service.QueryService) *QueryHandler {
	return &QueryHandler{s}
}

func (h *QueryHandler) Redirect(c *gin.Context) {
	short := c.Param("short")
	url, err := h.service.Redirect(c.Request.Context(), short)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "link not found",
		})
		return
	}
	c.Redirect(http.StatusFound, url)
}

func (h *QueryHandler) Stats(c *gin.Context) {
	short := c.Param("short")
	clicks, err := h.service.Stats(c.Request.Context(), short)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"short":  short,
		"clicks": clicks,
	})
}
