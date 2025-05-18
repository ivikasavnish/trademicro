package utils

import (
	"context"
	"errors"
)

// Key type for context values
type contextKey string

// Constant for user ID context key
const userIDKey contextKey = "userID"

// GetUserIDFromContext extracts the user ID from the context
func GetUserIDFromContext(ctx context.Context) (uint, error) {
	userID, ok := ctx.Value(userIDKey).(uint)
	if !ok {
		return 0, errors.New("user ID not found in context")
	}
	return userID, nil
}

// SetUserIDToContext adds the user ID to the context
func SetUserIDToContext(ctx context.Context, userID uint) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
