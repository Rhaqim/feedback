package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rhaqim/worldgame/internal/game"
	"github.com/rhaqim/worldgame/internal/handlers"
	"github.com/rhaqim/worldgame/internal/models"
)

func main() {
	port := flag.Int("port", 8080, "server port")
	flag.Parse()

	log.Println("=== WorldGame Server ===")
	log.Printf("Starting on port %d", *port)

	// Initialize components.
	challengeGen := game.NewChallengeGenerator()
	evaluator := game.NewEvaluator()

	// Hub will be set after we have the game manager.
	var hub *handlers.Hub

	// The broadcast function bridges the engine to the WebSocket hub.
	broadcastFn := func(gameID string, msg models.WSMessage) {
		if hub != nil {
			hub.BroadcastToGame(gameID, msg)
		}
	}

	engine := game.NewEngine(challengeGen, evaluator, broadcastFn)
	gameManager := game.NewGameManager(engine)

	hub = handlers.NewHub(gameManager)
	go hub.Run()

	apiHandler := handlers.NewAPIHandler(gameManager)

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
