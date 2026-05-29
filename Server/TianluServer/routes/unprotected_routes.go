package routes

import (
	"github.com/gin-gonic/gin"

	controller "github.com/solswiss/Tianlu/Server/TianluServer/controllers"
)

func SetupUnprotectedRoutes(router *gin.Engine) {
	router.GET("/", func(c *gin.Context) {
		c.String(200, "Welcome to Tianlu")
	})
	router.GET("/hello", func(c *gin.Context) {
		c.String(200, "Hello guest user :-)")
	})

	router.GET("/api/products", controller.GetProducts())

	router.POST("/register", controller.RegisterUser())
	router.POST("/login", controller.LoginUser())
}
