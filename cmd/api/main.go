package main

import (
	"gowitcase/db"
	"gowitcase/handlers"
	"gowitcase/services"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load env vars")
	}

	db.InitDB()
	db.InitRedis()

	ticketService := services.NewTicketService(&db.DB, &db.Redis)
	ticketHandler := handlers.NewTicketHandler(ticketService)

	router := gin.Default()

	//router.Use(middleware.TimeoutMiddleware(1 * time.Second))

	router.GET("/health", func(c *gin.Context) {
		if db.DB.IsHealthy() && db.Redis.IsHealthy() {
			c.JSON(200, gin.H{"status": "up"})
			return
		}
		c.JSON(500, gin.H{"status": "down"})
	})

	v1 := router.Group("/api/v1")
	{
		v1.POST("/tickets", ticketHandler.CreateTicket)
		v1.GET("/tickets/:id", ticketHandler.GetTicket)
		v1.POST("/tickets/:id/purchases", ticketHandler.PurchaseTicket)
	}

	// Swagger
	router.GET("/swagger.json", func(c *gin.Context) {
		c.File("docs/swagger.json")
	})
	router.GET("/api-docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger.json")))

	router.Run(":8080")
}
