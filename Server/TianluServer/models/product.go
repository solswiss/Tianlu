package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"

	"time"
)

type Category string
type Flavor string
type Texture string

type Product struct {
	ID             bson.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	ProductID      string        `bson:"product_id" json:"product_id" validate:"required"`
	Name           string        `bson:"name" json:"name" validate:"required,min=2,max=500"`
	GenericName    string        `bson:"generic_name" json:"generic_name" validate:"min=2,max=500"`
	Brand          string        `bson:"brand" json:"brand" validate:"max=200"`
	Origin         string        `bson:"origin" json:"origin" validate:"max=200"`
	Categories     []string      `bson:"categories" json:"categories" validate:"dive"`
	FlavorProfile  []string      `bson:"flavor_profile" json:"flavor_profile" validate:"dive"`
	TextureProfile []string      `bson:"texture_profile" json:"texture_profile" validate:"dive"`
	ImageURL       string        `bson:"image_url" json:"image_url" validate:"url"`
	Images         []string      `bson:"images_url" json:"images_url" validate:"dive,url"`
	ImageMiniURL   string        `bson:"image_mini_url" json:"image_mini_url" validate:"url"`
	Description    string        `bson:"description" json:"description" validate:"required"`
	Embedding      []float64     `bson:"embedding" json:"embedding"`
	CreatedAt      time.Time     `bson:"created_at" json:"created_at"`
}
