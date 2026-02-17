package platform

import (
	"fmt"

	"github.com/pkg/browser"
)

func OpenURL(url string) error {
	if err := browser.OpenURL(url); err != nil {
		return fmt.Errorf("open browser: %w", err)
	}
	return nil
}
