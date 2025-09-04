package utils

import (
	"context"
	"errors"
)

// GetUserFromContext extracts the user ID from the context
func GetUserFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value("userID").(string)
	if !ok || userID == "" {
		return "", errors.New("user not authenticated")
	}
	return userID, nil
}
