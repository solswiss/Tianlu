package routes

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"

	controller "github.com/solswiss/Tianlu/Server/TianluServer/controllers"
	mw "github.com/solswiss/Tianlu/Server/TianluServer/middleware"
)

func SetupProtectedGroups(router *gin.Engine, client *mongo.Client) {
	protected := router.Group("/api/p")
	protected.Use(mw.AuthMW()) // inject auth gate; aborts if client is unauthorized

	protected.POST("/add_product", controller.AddProduct(client))
	protected.GET("/product", controller.SearchProducts(client))
	protected.GET("/product/:barcode", controller.SearchProductByID(client))
	protected.GET("/recommend_products", controller.GetRecommendedProducts(client))
	protected.PATCH("/modify_product/:barcode", controller.UpdateProduct(client))
}
