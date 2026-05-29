package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func Connect() *mongo.Client {
	err := godotenv.Load(".env")
	// associated with dev/deployment (cloud) mixup
	if err != nil {
		log.Println("Warning: unable to find .env file", err)
		os.Exit(1)
	}

	MongoDB := os.Getenv("MONGODB_URI")
	if MongoDB == "" {
		log.Fatal("MONGODB_URI not set")
	}

	fmt.Println("MongoDB URI:", MongoDB)

	clientOptions := options.Client().ApplyURI(MongoDB)

	client, err := mongo.Connect(clientOptions)
	if err != nil {
	}

	return client
}

var Client *mongo.Client = Connect()

func SetupIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	index := mongo.IndexModel{
		Keys:    bson.M{"product_id": 1},
		Options: options.Index().SetUnique(true),
	}

	var coll = OpenCollection("products")
	if _, err := coll.Indexes().CreateOne(ctx, index); err != nil {
		log.Fatalf("Failed to set product unique index: %s", err)
	}

	index = mongo.IndexModel{
		Keys: bson.D{
			{Key: "name", Value: "text"},
			{Key: "generic_name", Value: "text"},
			{Key: "brand", Value: "text"},
		},
		Options: options.Index().SetName("ProductSearchIndex").SetWeights(bson.D{
			{Key: "name", Value: 10},
			{Key: "generic_name", Value: 5},
			{Key: "brand", Value: 4},
		}),
	}

	if _, err := coll.Indexes().CreateOne(ctx, index); err != nil {
		log.Fatalf("Failed to create product search index: %s", err)
	}

	log.Println("MongoDB indexes set up successfully")
}

func OpenCollection(collectionName string) *mongo.Collection {
	err := godotenv.Load(".env")
	// associated with dev/deployment (cloud) mixup
	if err != nil {
		log.Println("Warning: unable to find .env file")
	}

	databaseName := os.Getenv("DATABASE_NAME")
	fmt.Println("DATABASE_NAME:", databaseName)

	collection := Client.Database(databaseName).Collection(collectionName)
	if collection == nil {
		return nil
	}

	return collection
}
