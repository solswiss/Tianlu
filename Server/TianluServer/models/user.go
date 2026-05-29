package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID            bson.ObjectID      `bson:"_id,omitempty" json:"_id,omitempty"`
	UserID        string             `bson:"user_id" json:"user_id"`
	Username      string             `bson:"username" json:"username" validate:"required,min=1,max=20"`
	Email         string             `bson:"email" json:"email" validate:"omitempty,email"`
	Password      string             `bson:"password" json:"password" validate:"omitempty,min=3"`
	Role          string             `bson:"role" json:"role" validate:"required,oneof=admin user guest"`
	PalateProfile map[string]float64 `bson:"palate_profile" json:"palate_profile" validate:"dive"`
	TasteSummary  string             `bson:"taste_summary" json:"taste_summary"`
	Status        string             `bson:"status" json:"status" validate:"required,oneof=active banned"`
	Token         string             `bson:"token" json:"token"`
	RefreshToken  string             `bson:"refresh_token" json:"refresh_token"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

type UserLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=3"`
}

type UserResponse struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}
