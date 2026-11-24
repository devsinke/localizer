package localizer

import (
	"strings"
)

// Interpolate substitutes parameters in template strings.
// Supports both {key} and {{.key}} syntax cleanly.
func Interpolate(template string, params map[string]string) string {
	if len(params) == 0 || !strings.Contains(template, "{") {
		return template
	}

	result := template
	for k, v := range params {
		cleanKey := strings.TrimSpace(k)
		cleanKey = strings.TrimPrefix(cleanKey, ".")

		// Replace double curly brackets: {{.key}} and {{key}}
		result = strings.ReplaceAll(result, "{{."+cleanKey+"}}", v)
		result = strings.ReplaceAll(result, "{{"+cleanKey+"}}", v)

		// Replace single curly brackets: {.key} and {key}
		result = strings.ReplaceAll(result, "{."+cleanKey+"}", v)
		result = strings.ReplaceAll(result, "{"+cleanKey+"}", v)
	}

	return result
}
