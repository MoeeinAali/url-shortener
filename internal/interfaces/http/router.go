package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewRouter assembles the REST API. Command (POST) and query (GET) routes are
// registered side by side but delegate to the two separate CQRS handler sets.
func NewRouter(cmd *CommandHandler, qry *QueryHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Command side (writes).
	r.POST("/links", cmd.CreateLink)
	r.POST("/links/:short/disable", cmd.DisableLink)

	// Query side (reads).
	r.GET("/links/:short/stats", qry.Stats)
	r.GET("/:short", qry.Redirect)

	return r
}
