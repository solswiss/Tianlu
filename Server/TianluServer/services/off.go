package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/solswiss/Tianlu/Server/TianluServer/models"
)

func FindOFFProduct(barcode string) (models.OFFProductResponse, error) {
	// USING STAGING ENV | barcode search using v2 API
	url := fmt.Sprintf(
		"https://world.openfoodfacts.net/api/v2/product/%s",
		url.QueryEscape(barcode),
	)

	// HTTP request (OFF requires custom user-agent header)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Tianlu/1.0 (aeromatic@tutamail.com)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return models.OFFProductResponse{}, err
	}
	defer resp.Body.Close()

	// decode (unmarshal) response
	var container struct {
		Product models.OFFProductResponse `json:"product"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&container); err != nil {
		return models.OFFProductResponse{}, err
	}

	return container.Product, nil
}

func SearchOFFProducts(userQuery string) ([]models.OFFProductResponse, error) {
	// clean input, join with '+'
	//query := strings.Join(strings.Fields(strings.ToLower(userQuery)), "+")

	// fields of interest
	fields := "_id,product_name,generic_name,quantity,food_groups,food_groups_tags,brands_tags,categories_tags,labels_tags,image_url,image_thumb_url,allergens_tags,ingredients_text,origin"

	// USING STAGING ENV | full-text search using v1 API
	url := fmt.Sprintf(
		"https://world.openfoodfacts.net/cgi/search.pl?search_terms=%s&search_simple=1&action=process&json=1&fields=%s",
		url.QueryEscape(userQuery),
		fields,
	)

	// HTTP request (OFF requires custom user-agent header)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Tianlu/1.0 (aeromatic@tutamail.com)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// decode (unmarshal) response
	var container struct {
		Products []models.OFFProductResponse `json:"products"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&container); err != nil {
		return nil, err
	}

	return container.Products, nil
}
