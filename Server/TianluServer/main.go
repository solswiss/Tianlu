package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"

	db "github.com/solswiss/Tianlu/Server/TianluServer/database"
	routes "github.com/solswiss/Tianlu/Server/TianluServer/routes"
)

func main() {
	fmt.Println("Hello Tianlu server") // greet!

	// check .env, vital
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: unable to find .env file: %v", err.Error()) // associated with dev/deployment (cloud) mixup
	}

	// allowed origins setup (for CORS config)
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	var origins []string
	if allowedOrigins != "" {
		origins = strings.Split(allowedOrigins, ",")
		log.Println("Allowed origins: ")
		for i, s := range origins {
			origins[i] = strings.TrimSpace(s)
			log.Println(s)
		}
	} else {
		origins = []string{"http://localhost:5173"}
		log.Println("Allowed origins:\nhttp://localhost:5173")
	}

	// CORS config setup
	config := cors.Config{}
	config.AllowOrigins = origins
	config.AllowMethods = []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour

	// router
	router := gin.Default()
	router.Use(cors.New(config))
	router.Use(gin.Logger())

	var Client *mongo.Client = db.Connect()
	// confirm connection to MongoDB database
	if err := Client.Ping(context.Background(), nil); err != nil {
		log.Fatalf("Failed to reach server: %v", err.Error())
	}

	// ensure clean disconnect from MongoDB database
	defer func() {
		if err := Client.Disconnect(context.Background()); err != nil {
			log.Fatalf("Failed to disconnect from server: %v", err.Error())
		}
	}()

	db.SetupIndexes(Client)

	// routes setup
	routes.SetupUnprotectedRoutes(router, Client)
	routes.SetupProtectedGroups(router, Client)

	if err := router.Run(":8080"); err != nil {
		fmt.Println("Failed to start server", err)
	}
}
