package client

import (
	"net/url"
	"strings"
)

// replace all occurrences of $ placeholder in a string, url-encoded if desired
func replacePlaceholder(s string, replacement string) string {
	replacement = url.QueryEscape(replacement)
	return strings.Replace(s, "$", replacement, -1)
}
