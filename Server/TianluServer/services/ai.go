package services

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/solswiss/Tianlu/Server/TianluServer/models"

	"google.golang.org/api/googleapi"
	"google.golang.org/genai"
)

func cleanJSONResponse(raw string) string {
	// Trim leading/trailing whitespace and newlines
	cleaned := strings.TrimSpace(raw)

	// Remove leading ```json or ```
	if strings.HasPrefix(cleaned, "```json") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
	} else if strings.HasPrefix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```")
	}

	// Remove trailing ```
	cleaned = strings.TrimSuffix(cleaned, "```")

	return strings.TrimSpace(cleaned)
}

type ServiceResponse struct {
	Categories  []string `json:"categories"`
	Flavors     []string `json:"flavor_profile"`
	Textures    []string `json:"texture_profile"`
	Description string   `json:"description"`
}

func ClassifyProduct(c *gin.Context, keywords string) (ServiceResponse, error) {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: Failed to load .env")
		return ServiceResponse{}, err
	}

	ctx, cancel := context.WithTimeout(c, 100*time.Second)
	defer cancel()

	serviceKey := os.Getenv("SERVICE_KEY")
	if serviceKey == "" {
		return ServiceResponse{}, errors.New("Could not read SERVICE_KEY")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  serviceKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Println("Failed to connect AI service", err)
		return ServiceResponse{}, err
	}

	promptTemplate := os.Getenv("PROMPT_TEMPLATE")
	prompt := strings.Replace(
		strings.Replace(
			strings.Replace(
				strings.Replace(promptTemplate, "{categories}", strings.Join(models.Categories, ","), 1),
				"{flavors}", strings.Join(models.Flavors, ","), 1),
			"{textures}", strings.Join(models.Textures, ","), 1),
		"{keywords}", keywords, 1)

	res, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text(prompt),
		nil,
	)

	if err != nil {
		var googleError *googleapi.Error
		if errors.As(err, &googleError) {
			switch googleError.Code {
			case 429:
				log.Printf("Warning: Service rate limit exceeded")
			case 503:
				log.Printf("Warning: Service unavailable")
			default:
				log.Printf("Warning: Service error | %d: %s", googleError.Code, googleError.Details)
			}
		} else {
			log.Println("Failed to call AI service\n", err)
		}
		return ServiceResponse{}, err
	}

	if len(res.Candidates) == 0 || res.Candidates[0].Content == nil || len(res.Candidates[0].Content.Parts) == 0 {
		log.Println("AI service returned an empty response")
		return ServiceResponse{}, err
	}

	// sanitize response
	var response ServiceResponse
	text := cleanJSONResponse((res.Candidates[0].Content.Parts[0]).Text)

	if err := json.Unmarshal([]byte(text), &response); err != nil {
		return ServiceResponse{}, err
	}

	log.Println("Service response text: ", text)

	// convert whitelists into string-boolean maps and valid

	return response, nil
}
