package sanitize

import (
	"testing"
)

func TestLessonContent_StripsHTML(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"script", `<script>alert(1)</script>`, ""},
		{"img onerror", `Hello <img src=x onerror="alert(1)"> world`, "Hello  world"},
		{"javascript link", `<a href="javascript:alert(1)">click</a>`, "click"},
		{"markdown preserved", "**bold** and *italic* and [link](https://a.com)", "**bold** and *italic* and [link](https://a.com)"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LessonContent(tt.in)
			if got != tt.want {
				t.Errorf("LessonContent(%q) = %q; want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestDescription_StripsHTML(t *testing.T) {
	in := `<b>Safe</b> and <script>evil</script>`
	got := Description(in)
	if got != "Safe and " {
		t.Errorf("Description(%q) = %q; want %q", in, got, "Safe and ")
	}
}
