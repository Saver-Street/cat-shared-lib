package email

import (
	"testing"
)

func BenchmarkBuildMessage_TextOnly(b *testing.B) {
	msg := Message{
		To:      []string{"user@example.com"},
		Subject: "Hello",
		Text:    "Plain text body content for benchmarking.",
	}
	for b.Loop() {
		buildMessage("sender@example.com", msg)
	}
}

func BenchmarkBuildMessage_HTMLOnly(b *testing.B) {
	msg := Message{
		To:      []string{"user@example.com"},
		Subject: "Hello",
		HTML:    "<h1>Hello</h1><p>HTML body content for benchmarking.</p>",
	}
	for b.Loop() {
		buildMessage("sender@example.com", msg)
	}
}

func BenchmarkBuildMessage_Multipart(b *testing.B) {
	msg := Message{
		To:      []string{"user@example.com", "other@example.com"},
		CC:      []string{"cc@example.com"},
		Subject: "Multipart message",
		Text:    "Plain text fallback.",
		HTML:    "<h1>Hello</h1><p>HTML version with <b>formatting</b>.</p>",
		Headers: map[string]string{"X-Priority": "1"},
	}
	for b.Loop() {
		buildMessage("sender@example.com", msg)
	}
}

func BenchmarkParseHTMLString(b *testing.B) {
	src := `<html><body>Hello {{.Name}}, welcome!</body></html>`
	for b.Loop() {
		ParseHTMLString("welcome", src)
	}
}

func BenchmarkParseTextString(b *testing.B) {
	src := `Hello {{.Name}}, welcome to {{.Service}}!`
	for b.Loop() {
		ParseTextString("welcome", src)
	}
}

func BenchmarkRenderHTML(b *testing.B) {
	tmpl, _ := ParseHTMLString("welcome", `<html><body>Hello {{.Name}}</body></html>`)
	data := map[string]string{"Name": "Alice"}
	b.ResetTimer()
	for b.Loop() {
		RenderHTML(tmpl, "welcome", data)
	}
}

func BenchmarkRenderText(b *testing.B) {
	tmpl, _ := ParseTextString("welcome", `Hello {{.Name}}, welcome!`)
	data := map[string]string{"Name": "Alice"}
	b.ResetTimer()
	for b.Loop() {
		RenderText(tmpl, "welcome", data)
	}
}
