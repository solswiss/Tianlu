package routes

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"

	controller "github.com/solswiss/Tianlu/Server/TianluServer/controllers"
)

func SetupUnprotectedRoutes(router *gin.Engine, client *mongo.Client) {
	router.GET("/", func(c *gin.Context) {
		c.String(200, "Welcome to Tianlu")
	})
	router.GET("/hello", func(c *gin.Context) {
		c.String(200, "Hello guest user :-)")
	})

	router.GET("/api/products", controller.GetProducts(client))

	router.POST("/api/register", controller.RegisterUser(client))
	router.POST("/api/login", controller.LoginUser(client))
}
