package controllers

import (
	"context"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/solswiss/Tianlu/Server/TianluServer/database"
	"github.com/solswiss/Tianlu/Server/TianluServer/models"
	"github.com/solswiss/Tianlu/Server/TianluServer/services"
	"github.com/solswiss/Tianlu/Server/TianluServer/utils"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// LOCAL GET / FETCH / RETRIEVE
func GetProducts(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c, 100*time.Second)
		defer cancel()

		var products []models.Product

		productCollection := database.OpenCollection(client, "products")

		cursor, err := productCollection.Find(ctx, bson.M{})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		}
		defer cursor.Close(ctx)

		if err = cursor.All(ctx, &products); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode products"})
		}

		c.JSON(http.StatusOK, products)
	}
}

func GetProduct(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c, 100*time.Second)
		defer cancel()

		productID := c.Param("product_id")

		if productID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
			return
		}

		var product models.Product

		productCollection := database.OpenCollection(client, "products")

		if err := productCollection.FindOne(ctx, bson.M{"product_id": productID}).Decode(&product); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		c.JSON(http.StatusOK, product)
	}
}

// SEARCH
// search products by string query
func SearchProducts(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("query")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Query cannot be empty"})
			return
		}

		productCollection := database.OpenCollection(client, "products")

		localProducts, err := utils.SearchLocalProducts(c, productCollection, query)

		// lim first 10
		if err == nil && len(localProducts) > 9 {
			c.JSON(http.StatusOK, localProducts[:10])
		}

		OFFProducts, err := services.SearchOFFProducts(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product data from OFF"})
			log.Println("Warning: Failed to fetch products from OFF")
			return
		}

		res := utils.MergeSearchedProducts(c, productCollection, localProducts, OFFProducts)

		if len(res) < 10 {
			c.JSON(http.StatusOK, res)
			return
		}

		c.JSON(http.StatusOK, res[:10])
	}
}

// search product by product barcode
func SearchProductByID(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		barcode := c.Param("barcode")
		if len(barcode) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID missing"})
			return
		}

		ctx, cancel := context.WithTimeout(c, 100*time.Second)
		defer cancel()

		// search local db
		var res models.Product

		productCollection := database.OpenCollection(client, "products")

		err := productCollection.FindOne(ctx, bson.M{"product_id": barcode}).Decode(&res)
		// not found -> search OFF
		if err == mongo.ErrNoDocuments {
			res, err := services.FindOFFProduct(barcode)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			// transform, add to local db, then show
			p, err := utils.FormatOFFProduct(c, res)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if _, err := productCollection.InsertOne(ctx, p); err != nil {
				log.Printf("Warning: Failed to insert searched product by ID %s", barcode)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, p)
		}
		c.JSON(http.StatusOK, res)
	}
}

// ADD
func AddProduct(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c, 100*time.Second)
		defer cancel()

		var product models.Product

		if err := c.ShouldBindJSON(&product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := validate.Struct(product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
			return
		}

		productCollection := database.OpenCollection(client, "products")

		result, err := productCollection.InsertOne(ctx, product)

		// catch duplicate products by product_id
		if mongo.IsDuplicateKeyError(err) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Product already exists"})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add product", "details": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, result)
	}
}

// UPDATE
// admin-only modifiable fields
type UpdateProductInput struct {
	Brand          string   `json:"brand"`
	Origin         string   `json:"origin"`
	Categories     []string `json:"category"`
	FlavorProfile  []string `json:"flavor_profile"`
	TextureProfile []string `json:"texture_profile"`
	ImageURL       string   `json:"image_url"`
	Images         []string `json:"images_url"`
	ImageMiniURL   string   `json:"image_mini_url"`
	Description    string   `json:"description"`
}

func UpdateProduct(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, err := utils.GetUserRoleFromContext(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "User must be an administrator"})
			return
		}

		// get barcode UID
		barcode := c.Param("barcode")

		var input UpdateProductInput

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// prepare updated fields
		updateData := bson.M{}

		if len(input.Categories) > 0 {
			updateData["categories"] = input.Categories
		}
		if input.ImageURL != "" {
			updateData["image_url"] = input.ImageURL
		}

		ctx, cancel := context.WithTimeout(c, 100*time.Second)
		defer cancel()

		productCollection := database.OpenCollection(client, "products")

		var p models.Product
		if err := productCollection.FindOne(ctx, bson.M{"product_id": barcode}).Decode(&p); err == mongo.ErrNoDocuments {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No product found with provided ID"})
			return
		}

		// add image url if new
		if !slices.Contains(p.Images, input.ImageURL) {
			updateData["images_url"] = append(p.Images, input.ImageURL)
		}

		if len(updateData) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields provided for update"})
			return
		}

		// update document
		result, err := productCollection.UpdateOne(ctx, bson.M{"product_id": barcode}, bson.M{"$set": updateData})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found in local database"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully"})
	}
}

// RECOMMENDATION
func GetRecommendedProducts(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := utils.GetUserIDFromContext(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// get recommended products using userID
		c.JSON(http.StatusOK, gin.H{"message": "Sorry " + userID + ", this feature is still in development"})
	}
}
