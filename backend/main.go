package main

import (
	"log"
	"net/http"

	"livepoll/config"
	"livepoll/database"
	"livepoll/handlers"
	"livepoll/middleware"
	"livepoll/redisdb"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load configuration
	config.Load()

	// 2. Connect to MongoDB and Redis
	database.Connect()
	redisdb.Connect()

	// 3. Set up Gin router
	r := gin.Default()

	// CORS — dynamically allow origins (Vercel, localhost, custom domains) with credentials
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 4. Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 5. Auth routes (no JWT required)
	auth := r.Group("/auth")
	{
		auth.POST("/signup", handlers.Signup)
		auth.POST("/login", handlers.Login)
	}

	// 6. Poll routes
	polls := r.Group("/polls")
	{
		// Auth-required
		polls.POST("", middleware.AuthRequired(), handlers.CreatePoll)
		polls.GET("/mine", middleware.AuthRequired(), handlers.GetMyPolls)

		// Public
		polls.GET("/:shareCode", handlers.GetPoll)
		polls.POST("/:shareCode/vote", handlers.Vote)
	}

	// 7. WebSocket endpoint (public — anyone with the share code can watch live results)
	r.GET("/ws/poll/:shareCode", handlers.WSPoll)

	// 8. Start server
	addr := ":" + config.App.Port
	log.Printf("[server] listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("[server] fatal: %v", err)
	}
}
