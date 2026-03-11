package sanitize

import "testing"

func BenchmarkStripHTML(b *testing.B) {
	html := `<div class="content"><p>Hello <b>world</b>!</p><script>alert("xss")</script></div>`
	for b.Loop() {
		StripHTML(html)
	}
}

func BenchmarkDeref(b *testing.B) {
	s := "hello"
	p := &s
	for b.Loop() {
		Deref(p, "default")
	}
}

