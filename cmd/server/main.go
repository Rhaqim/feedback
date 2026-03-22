package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rhaqim/worldgame/internal/database"
	"github.com/rhaqim/worldgame/internal/external"
	"github.com/rhaqim/worldgame/internal/game"
	"github.com/rhaqim/worldgame/internal/handlers"
	"github.com/rhaqim/worldgame/internal/models"
	"github.com/rhaqim/worldgame/internal/store"
)

func main() {
	port := flag.Int("port", 8080, "server port")
	flag.Parse()

	log.Println("=== WorldGame Server ===")
	log.Printf("Starting on port %d", *port)

	ctx := context.Background()

	// Database connection.
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://worldgame:worldgame@localhost:5432/worldgame?sslmode=disable"
	}

	pool, err := database.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	s := store.New(pool)

	if err := database.Seed(ctx, s); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	// Initialize components.
	challengeGen := game.NewChallengeGenerator(s)
	evaluator := game.NewEvaluator()

	// Hub will be set after we have the game manager.
	var hub *handlers.Hub

	// The broadcast function bridges the engine to the WebSocket hub.
	broadcastFn := func(gameID string, msg models.WSMessage) {
		if hub != nil {
			hub.BroadcastToGame(gameID, msg)
		}
	}

	engine := game.NewEngine(challengeGen, evaluator, broadcastFn, s)
	gameManager := game.NewGameManager(engine, s)

	hub = handlers.NewHub(gameManager, s)
	go hub.Run()

	// Feed service: fetches real RSS feeds periodically.
	feedSources := external.DefaultFeedSources()
	feedFetcher := external.NewFeedFetcher(feedSources)
	feedInterval := 30 * time.Minute
	if envInterval := os.Getenv("FEED_INTERVAL"); envInterval != "" {
		if d, err := time.ParseDuration(envInterval); err == nil {
			feedInterval = d
		}
	}
	feedService := external.NewFeedService(feedFetcher, s, feedInterval)
	feedService.Start()
	defer feedService.Stop()

	apiHandler := handlers.NewAPIHandler(gameManager, s)
	apiHandler.SetFeedService(feedService)

	// Set up gin router.
	r := gin.Default()

	// CORS middleware.
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: false,
	}))

	// REST endpoints.
	api := r.Group("/api")
	{
		api.POST("/games", apiHandler.CreateGame)
		api.GET("/games", apiHandler.ListGames)
		api.GET("/games/:id", apiHandler.GetGame)
		api.POST("/games/:id/join", apiHandler.JoinGame)
		api.POST("/games/:id/proposals", apiHandler.SubmitProposal)
		api.POST("/games/:id/evaluate", apiHandler.Evaluate)
		api.POST("/games/:id/next-week", apiHandler.NextWeek)
		api.GET("/regions", apiHandler.ListRegions)

		// Chat history
		api.GET("/games/:id/chat", apiHandler.GetChat)

		// Admin routes
		api.POST("/regions", apiHandler.CreateRegion)
		api.POST("/challenge-templates", apiHandler.CreateChallengeTemplate)
		api.GET("/challenge-templates", apiHandler.ListChallengeTemplates)

		// Feed routes
		api.GET("/feeds", apiHandler.ListFeedItems)
		api.POST("/feeds/fetch", apiHandler.TriggerFeedFetch)
	}

	// WebSocket endpoint.
	r.GET("/ws", func(c *gin.Context) {
		hub.HandleWebSocket(c.Writer, c.Request)
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Server listening on %s", addr)
	log.Printf("REST API: http://localhost%s/api/", addr)
	log.Printf("WebSocket: ws://localhost%s/ws", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
