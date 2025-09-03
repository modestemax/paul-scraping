package install

import (
	"fmt"
	"strings"

	"github.com/playwright-community/playwright-go"
)

// StartPlaywright starts Playwright, and optionally auto-installs the driver/browsers if missing.
func StartPlaywright(autoInstall bool) (*playwright.Playwright, error) {
	pw, err := playwright.Run()
	if err == nil {
		return pw, nil
	}
	if autoInstall && strings.Contains(err.Error(), "please install the driver") {
		if err2 := playwright.Install(); err2 != nil {
			return nil, fmt.Errorf("auto-install failed: %w", err2)
		}
		return playwright.Run()
	}
	return nil, fmt.Errorf("could not start Playwright: %w", err)
}
