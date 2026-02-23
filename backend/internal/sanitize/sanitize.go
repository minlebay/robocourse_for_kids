package sanitize

import (
	"strings"
	"unicode"

	"github.com/microcosm-cc/bluemonday"
)

var (
	// strictPolicy strips all HTML tags; safe for markdown content that may contain raw HTML.
	// Markdown syntax (**, ##, etc.) is preserved; only <...> tags are removed.
	strictPolicy = bluemonday.StrictPolicy()
)

// LessonContent removes HTML/script from markdown lesson step content to prevent XSS.
// Use before storing in DB. Keeps plain text and markdown; strips all HTML tags.
func LessonContent(s string) string {
	return strictPolicy.Sanitize(s)
}

// Description removes HTML from short text (module/lesson description). Use before storing in DB.
func Description(s string) string {
	return strictPolicy.Sanitize(s)
}

// ChatMessage strips null bytes and ASCII control characters from a user chat message,
// preserving newlines and tabs which are legitimate in conversational text.
// Use before storing and before sending to external AI APIs.
func ChatMessage(s string) string {
	s = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' {
			return r
		}
		if unicode.IsControl(r) {
			return -1 // drop
		}
		return r
	}, s)
	return strings.TrimSpace(s)
}
