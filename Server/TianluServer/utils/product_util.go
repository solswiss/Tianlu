package utils

import (
	"context"
	"errors"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/solswiss/Tianlu/Server/TianluServer/models"
	"github.com/solswiss/Tianlu/Server/TianluServer/services"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// SEARCH
func SearchLocalProducts(coll *mongo.Collection, query string) ([]models.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	// MongoDB text search filter
	filter := bson.M{"$text": bson.M{"$search": query}}
	// sort by text search score relevance
	opts := options.Find().
		SetProjection(bson.M{"score": bson.M{"$meta": "textScore"}}).
		SetSort(bson.M{"score": bson.M{"$meta": "textScore"}}).
		SetLimit(10) // set limit to top 10

	var localProducts []models.Product

	// search local db
	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		log.Printf("Warning: Failed Local DB product search | %s", err)
		return []models.Product{}, err
	}

	if err = cursor.All(ctx, &localProducts); err != nil {
		log.Printf("Warning: Failed Local DB product search | %s", err)
		return []models.Product{}, err
	}
	// return or cache local products if found
	if len(localProducts) > 10 {
		return localProducts[:10], nil
	}
	return localProducts, nil
}

// filters b for unique products where product_id is not shared
func FilterUniqueProducts(a []models.Product, b []models.OFFProductResponse) []models.OFFProductResponse {
	inA := make(map[string]struct{}, len(a))

	for _, p := range a {
		inA[p.ProductID] = struct{}{}
	}

	result := make([]models.OFFProductResponse, 0, len(b))

	for _, p := range b {
		if _, found := inA[p.Barcode]; !found {
			result = append(result, p)
		}
	}

	return result
}

func FormatOFFProduct(op models.OFFProductResponse) (models.Product, error) {
	brand := ""
	if len(op.Brands) > 0 {
		brand = strings.TrimPrefix(op.Brands[0], "en:")
	}

	genericName := op.GenericName
	if genericName == "" && len(op.Categories) > 0 {
		genericName = strings.TrimPrefix(op.Categories[0], "en:")
	}

	image := op.ImageURL
	image_mini := op.ImageThumbURL

	// AI stuff
	data := []string{
		"Brand:", brand,
		"Product name:", op.ProductName,
		"Identifiers:", genericName, op.FoodGroupsString, strings.Join(op.Categories, ", "), strings.Join(op.Labels, ", "),
		"Ingredients: " + op.Ingredients,
	}

	AIContextString := strings.Join(data, " ")
	AIRes, err := services.ClassifyProduct(AIContextString)
	if err != nil {
		log.Printf("Warning: Service failed to profile product %s", op.Barcode)
		return models.Product{}, errors.New("Service failed to profile product")
	}

	return models.Product{
		ProductID:      op.Barcode,
		Name:           op.ProductName,
		GenericName:    genericName,
		Brand:          brand,
		Origin:         op.Origin,
		Categories:     AIRes.Categories,
		FlavorProfile:  AIRes.Flavors,
		TextureProfile: AIRes.Textures,
		ImageURL:       image,
		ImageMiniURL:   image_mini,
		Description:    AIRes.Description,
		Embedding:      []float64{},
		CreatedAt:      time.Now(),
	}, nil
}

func InsertProducts(coll *mongo.Collection, products []models.Product) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	documents := make([]any, len(products))
	for i, p := range products {
		documents[i] = p
	}

	if _, err := coll.InsertMany(ctx, documents); err != nil {
		log.Printf("Warning: Failed to insert products")
		return
	}
}

func GenerateProductUpsertPipeline(p models.Product) mongo.Pipeline {
	data, err := bson.Marshal(p)
	if err != nil {
		return mongo.Pipeline{}
	}

	var productMap bson.M
	bson.Unmarshal(data, &productMap)

	conditional := bson.M{
		"product_id": "$product_id",
		"created_at": "$created_at",
	}

	for k, v := range productMap {
		// skip permanent keys
		if k == "product_id" || k == "created_at" {
			continue
		}

		// handle type: array or string
		var emptyMatches bson.A
		switch v.(type) {
		case bson.A, []any:
			// array/slice, target null or empty arrays []
			emptyMatches = bson.A{nil, bson.A{}}
		default:
			// text/number v, target null or empty strings ""
			emptyMatches = bson.A{nil, ""}
		}

		// Inject the dynamic MongoDB pipeline rule
		conditional[k] = bson.M{
			"$cond": bson.A{
				bson.M{"$in": bson.A{"$" + k, emptyMatches}},
				v,       // if empty, insert the incoming Go v
				"$" + k, // if populated, retain the existing DB v
			},
		}
	}

	return mongo.Pipeline{
		bson.D{{Key: "$replaceWith", Value: bson.M{
			"$mergeObjects": bson.A{
				productMap,
				conditional,
			},
		}}},
	}
}

func MergeSearchedProducts(coll *mongo.Collection, localProducts []models.Product, OFFProducts []models.OFFProductResponse) []models.Product {
	var mergedProducts []models.Product
	newProducts := FilterUniqueProducts(localProducts, OFFProducts)
	if len(newProducts) > 0 {
		var formattedProducts []models.Product
		for _, op := range newProducts {
			p, err := FormatOFFProduct(op)
			if err != nil {
				continue
			}
			formattedProducts = append(formattedProducts, p)
		}
		mergedProducts = slices.Concat(mergedProducts, formattedProducts)

		// insert missing products to database as background process
		if len(mergedProducts) > len(localProducts) {
			go InsertProducts(coll, formattedProducts)
		}
	}
	return mergedProducts
}

func UpsertProducts(coll *mongo.Collection, products []models.Product) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var documents []any
	// upsert: insert if missing, update if found
	for _, p := range products {
		// handle duplicate by product ID
		filter := bson.M{"barcode": p.ProductID}

		// only insert a field if it is missing or empty

		/* pipeline := mongo.Pipeline{bson.D{{Key: "$replaceWith", Value: bson.M{"$mergeObjects": bson.A{p, bson.M{
			"origin": bson.M{"$cond": bson.A{
				bson.M{"$in": bson.A{"$origin", bson.A{nil, ""}}}, p.Name, "$name",
			}},

			// check if arrays are null or empty []
			"categories": bson.M{"$cond": bson.A{
				bson.M{"$in": bson.A{"$categories", bson.A{nil, bson.A{}}}}, p.Categories, "$categories",
			}},
			"flavor_profile": bson.M{"$cond": bson.A{
				bson.M{"$in": bson.A{"$flavor_profile", bson.A{nil, bson.A{}}}}, p.Categories, "$categories",
			}},
			"taste_profile": bson.M{"$cond": bson.A{
				bson.M{"$in": bson.A{"$taste_profile", bson.A{nil, bson.A{}}}}, p.Categories, "$categories",
			}},

			"image_url": bson.M{"$cond": bson.A{
				bson.M{"$in": bson.A{"$image_url", bson.A{nil, ""}}}, p.GenericName, "$generic_name",
			}},
			"images_url": bson.M{"$cond": bson.A{
				bson.M{"$in": bson.A{"$images_url", bson.A{nil, ""}}}, p.Origin, "$origin",
			}},
			"image_mini_url": bson.M{"$cond": bson.A{
				bson.M{"$in": bson.A{"$image_mini_url", bson.A{nil, ""}}}, p.GenericName, "$generic_name",
			}},
			"description": bson.M{"$cond": bson.A{
				bson.M{"$in": bson.A{"$description", bson.A{nil, ""}}}, p.Origin, "$origin",
			}},
			"embedding": bson.M{"$cond": bson.A{
				bson.M{"$in": bson.A{"$embedding", bson.A{nil, bson.A{}}}}, p.Categories, "$categories",
			}},
			"product_id": "$product_id",
			"createdAt":  "$createdAt",
		}},}}},} */

		pipeline := GenerateProductUpsertPipeline(p)

		d := mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(pipeline).SetUpsert(true)
		documents = append(documents, d)
	}
	if len(documents) > 0 {
		// bulkwrite can handle upsert; insertmany can only insert new documents
		if _, err := coll.InsertMany(ctx, documents); err != nil {
			log.Println("Failed to bulk upsert:", err)
		}
	}
}

func CalculateProductRelevanceScore(p models.Product, queryWords []string) int {
	score := 0
	fullName := strings.ToLower(p.Brand + " " + p.Name)

	for _, word := range queryWords {
		if strings.Contains(fullName, word) {
			score += 10 // Heavy points for direct word matches in the title
		}
	}
	return score
}

// Jaccard-style keyword relevance sort
func RankProductResults(products []models.Product, query string) {
	queryWords := strings.Fields(strings.ToLower(query))

	// inline sorting layout
	for i := range len(products) {
		I := CalculateProductRelevanceScore(products[i], queryWords)
		for j := i + 1; j < len(products); j++ {
			post := CalculateProductRelevanceScore(products[j], queryWords)
			if post > I {
				products[i], products[j] = products[j], products[i]
			}
		}
	}
}
