package web

import (
	"time"

	"github.com/gin-gonic/gin"
)

func configRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.String(200, "pong")
		})
		api.POST("/task", TaskAdd)
		api.GET("/task", TaskGets)

	}
}

func GetNowTs(c *gin.Context) {
	c.String(200, time.Now().Format("2006-01-02 15:04:05"))
}
