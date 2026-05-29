package models

type OFFProductRequest struct {
	ProductName string `json:"name"`
}

type OFFProductResponse struct {
	Barcode          string   `json:"_id"`
	ProductName      string   `json:"product_name"`
	GenericName      string   `json:"generic_name"`
	Obsolete         string   `json:"obsolete"`
	Quantity         string   `json:"quantity"`
	FoodGroupsString string   `json:"food_groups"`
	FoodGroups       []string `json:"food_groups_tags"`
	Brands           []string `json:"brands_tags"`
	Categories       []string `json:"categories_tags"`
	Labels           []string `json:"labels_tags"`
	ImageURL         string   `json:"image_url"`
	ImageThumbURL    string   `json:"image_thumb_url"`
	Allergens        []string `json:"allergens_tags"`
	Ingredients      string   `json:"ingredients_text"`
	Origin           string   `json:"origin"`
}
