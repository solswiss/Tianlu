package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func GetUserIDFromContext(c *gin.Context) (string, error) {
	userID, exists := c.Get("userID")
	if !exists {
		return "", errors.New("userID does not exist in this context")
	}

	id, ok := userID.(string)
	if !ok {
		return "", errors.New("Unable to retrieve userID")
	}

	return id, nil
}

func GetUserRoleFromContext(c *gin.Context) (string, error) {
	r, exists := c.Get("role")
	if !exists {
		return "", errors.New("role does not exist in this context")
	}

	role, ok := r.(string)
	if !ok {
		return "", errors.New("Unable to retrieve role")
	}

	return role, nil
}

// password
func HashPassword(pw string) (string, error) {
	password, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return "", err // panic
	}
	return string(password), nil
}
