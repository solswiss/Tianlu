package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type SubRating struct {
	TasteRating     float32 `bson:"taste" json:"taste" validate:"min=0.0,max=10.0"`
	TextureRating   float32 `bson:"texture" json:"texture" validate:"min=0.0,max=10.0"`
	PackagingRating float32 `bson:"packaging" json:"packaging" validate:"min=0.0,max=10.0"`
}

type Review struct {
	ID             bson.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	ProductID      string        `bson:"product_id" json:"product_id" validate:"required"`
	UserID         string        `bson:"user_id" json:"user_id" validate:"required"`
	Rating         float32       `bson:"rating" json:"rating" validate:"required,min=0.0,max=10.0"`
	SubRating      SubRating     `bson:"subrating" json:"subrating" validate:"omitempty,dive"`
	Text           string        `bson:"text" json:"text" validate:"omitempty,max=4096"`
	ImageID        string        `bson:"image_id" json:"image_id" validate:"omitempty"`
	SentimentScore float32       `bson:"sentiment_score" json:"sentiment_score"`
	CreatedAt      time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time     `bson:"updated_at" json:"updated_at"`
}
