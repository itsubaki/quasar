package handler

import (
	"fmt"
	"strings"
)

// JSON may encode each source byte as a six-byte Unicode escape.
const maxRequestBytes = 6*maxSize + 1024

func checkCode(code string) error {
	if len(code) > maxSize {
		return fmt.Errorf("code size exceeds %d bytes", maxSize)
	}

	if len(strings.TrimSpace(code)) == 0 {
		return ErrCodeNotFound
	}

	return nil
}
