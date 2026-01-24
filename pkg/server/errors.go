package server

import (
	"errors"
	"fmt"

	"github.com/nbaertsch/mythic-sdk-go/pkg/mythic"
)

// translateError converts Mythic SDK errors to user-friendly error messages
func translateError(err error) error {
	if err == nil {
		return nil
	}

	// Check for specific Mythic SDK errors
	switch {
	case errors.Is(err, mythic.ErrNotAuthenticated):
		return fmt.Errorf("not authenticated with Mythic server - please check credentials")

	case errors.Is(err, mythic.ErrAuthenticationFailed):
		return fmt.Errorf("authentication failed - invalid credentials or server unreachable")

	case errors.Is(err, mythic.ErrNotFound):
		return fmt.Errorf("requested resource not found")

	case errors.Is(err, mythic.ErrInvalidInput):
		return fmt.Errorf("invalid input parameters")

	case errors.Is(err, mythic.ErrTimeout):
		return fmt.Errorf("request timed out - Mythic server may be slow or unreachable")

	default:
		// Return original error with context
		return fmt.Errorf("Mythic operation failed: %w", err)
	}
}
