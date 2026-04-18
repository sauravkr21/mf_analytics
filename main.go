package main

import (
	"github.com/gin-gonic/gin"
	"mf-analytics/internal/api"
	"mf-analytics/internal/db"
	"mf-analytics/internal/rate"
)

func main() {
	db.Init()

	rl := rate.NewLimiter()

	r := gin.Default()

	// root route (optional)
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "MF Analytics Running"})
	})

	api.Init(r, rl)

	r.Run(":8080")
}
