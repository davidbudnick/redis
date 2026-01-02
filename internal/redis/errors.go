package redis

import "fmt"

// errInvalidRegex wraps an error for invalid regex patterns
func errInvalidRegex(err error) error {
	return fmt.Errorf("invalid regex: %w", err)
}
