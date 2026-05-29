package main

import (
	"fmt"

	"github.com/gin-gonic/gin"

	db "github.com/solswiss/Tianlu/Server/TianluServer/database"
	routes "github.com/solswiss/Tianlu/Server/TianluServer/routes"
)

func main() {
	fmt.Println("Hello Tianlu server")

	db.SetupIndexes()

	router := gin.Default()

	// setup routes
	routes.SetupUnprotectedRoutes(router)
	routes.SetupProtectedGroups(router)

	if err := router.Run(":8080"); err != nil {
		fmt.Println("Failed to start server", err)
	}
}
