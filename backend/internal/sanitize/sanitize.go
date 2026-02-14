package sanitize

import (
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
