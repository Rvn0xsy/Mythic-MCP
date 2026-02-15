package server

import (
	"errors"
	"fmt"

	"github.com/nbaertsch/mythic-sdk-go/pkg/mythic"
)

// translateError converts Mythic SDK errors to user-friendly error messages.
// The SDK's WrapError already provides rich context (operation, resource ID,
// specific message), so we preserve that detail and only add hints for
// authentication errors where extra guidance is useful.
func translateError(err error) error {
	if err == nil {
		return nil
	}

	// Check for specific Mythic SDK errors
	switch {
	case errors.Is(err, mythic.ErrNotAuthenticated):
		return fmt.Errorf("%v — check credentials and ensure the Mythic server is running", err)

	case errors.Is(err, mythic.ErrAuthenticationFailed):
		return fmt.Errorf("%v — invalid credentials or server unreachable", err)

	case errors.Is(err, mythic.ErrNotFound):
		// SDK already says e.g. "GetCallbackByID: callback with display_id 7 not found: not found"
		return err

	case errors.Is(err, mythic.ErrInvalidInput):
		// SDK already says e.g. "IssueTask: command is required: invalid input"
		return err

	case errors.Is(err, mythic.ErrTimeout):
		// SDK already says e.g. "WaitForTaskComplete: task 5 did not complete within 60s: request timeout"
		return err

	case errors.Is(err, mythic.ErrTaskFailed):
		// SDK already says e.g. "WaitForTaskComplete: task failed with stderr: ..."
		return err

	default:
		// Return original error with context
		return fmt.Errorf("Mythic operation failed: %w", err)
	}
}
