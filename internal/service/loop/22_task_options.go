package loop

import (
	"fmt"
	"strings"
)

func parseAutoFetchSiteEnabled(raw string, defaultValue bool) (bool, error) {
	current := strings.TrimSpace(strings.ToLower(raw))
	if current == "" {
		return defaultValue, nil
	}
	switch current {
	case "1", "true", "on", "yes", "y":
		return true, nil
	case "0", "false", "off", "no", "n":
		return false, nil
	default:
		return false, fmt.Errorf("auto_fetch_site 参数错误")
	}
}
