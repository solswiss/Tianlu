package routes

import (
	"github.com/gin-gonic/gin"

	controller "github.com/solswiss/Tianlu/Server/TianluServer/controllers"
	mw "github.com/solswiss/Tianlu/Server/TianluServer/middleware"
)

func SetupProtectedGroups(router *gin.Engine) {
	router.Use(mw.AuthMW()) // inject auth gate; aborts if client is unauthorized

	router.POST("/api/add_product", controller.AddProduct())
	router.GET("/api/product/:product_id", controller.GetProduct())
	router.GET("/api/product", controller.SearchProducts())
	router.GET("/api/product/:barcode", controller.SearchProductByID())
	router.GET("/api/OFFproduct", controller.SearchProductString())
	router.PATCH("/api/modify_product/:barcode", controller.UpdateProduct())
}
