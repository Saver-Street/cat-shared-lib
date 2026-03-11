package email

import (
	"testing"
)

// FuzzParseHTMLString exercises the HTML template parser with arbitrary
// template source strings. Malformed Go template syntax should return an
// error, never panic.
func FuzzParseHTMLString(f *testing.F) {
	f.Add("greeting", "<h1>Hello {{.Name}}</h1>")
	f.Add("empty", "")
	f.Add("nested", `{{define "inner"}}inner{{end}}{{template "inner"}}`)
	f.Add("action", `{{if .OK}}yes{{else}}no{{end}}`)
	f.Add("range", `{{range .Items}}<li>{{.}}</li>{{end}}`)
	f.Add("unclosed", `{{if .X}}<p>{{.X}`)
	f.Add("pipe", `{{"hello" | printf "%q"}}`)
	f.Add("special", "{{.}} \x00\xff\n\r\t")

	f.Fuzz(func(t *testing.T, name, src string) {
		tmpl, err := ParseHTMLString(name, src)
		if err != nil {
			return // expected for malformed templates
		}
		if tmpl == nil {
			t.Error("ParseHTMLString returned nil template without error")
		}
	})
}

// FuzzParseTextString exercises the text template parser with arbitrary
// template source strings.
func FuzzParseTextString(f *testing.F) {
	f.Add("greeting", "Hello {{.Name}}")
	f.Add("empty", "")
	f.Add("nested", `{{define "inner"}}inner{{end}}{{template "inner"}}`)
	f.Add("unclosed", `{{if .X}}{{.X}`)
	f.Add("special", "{{.}} \x00\xff\n\r\t")

	f.Fuzz(func(t *testing.T, name, src string) {
		tmpl, err := ParseTextString(name, src)
		if err != nil {
			return
		}
		if tmpl == nil {
			t.Error("ParseTextString returned nil template without error")
		}
	})
}

// FuzzBuildMessage exercises MIME email construction with arbitrary field
// values. It checks that buildMessage either returns valid bytes or an error,
// never panics.
func FuzzBuildMessage(f *testing.F) {
	f.Add(
		"sender@example.com",
		"recipient@example.com",
		"Test Subject",
		"<h1>Hello</h1>",
		"Hello plain",
	)
	f.Add("", "", "", "", "")
	f.Add(
		"noreply@cat.dev",
		"user@example.com",
		"Ünïcödé Sübjéct 🎉",
		"<p>Ünïcödé HTML 🎉</p>",
		"",
	)
	f.Add(
		"a@b.c",
		"x@y.z",
		"",
		"",
		"plain only",
	)
	f.Add(
		"from@example.com",
		"to@example.com",
		"Long subject "+string(make([]byte, 200)),
		"<html>"+string(make([]byte, 1000))+"</html>",
		"text "+string(make([]byte, 1000)),
	)

	f.Fuzz(func(t *testing.T, from, to, subject, html, text string) {
		msg := Message{
			To:      []string{to},
			Subject: subject,
			HTML:    html,
			Text:    text,
		}
		raw, err := buildMessage(from, msg)
		if err != nil {
			return
		}
		if len(raw) == 0 {
			t.Error("buildMessage returned empty bytes without error")
		}
	})
}

// FuzzMessageBuilder exercises the fluent message builder with arbitrary
// inputs to verify it never panics.
func FuzzMessageBuilder(f *testing.F) {
	f.Add("to@example.com", "cc@example.com", "Subject", "<b>Hi</b>", "Hi")
	f.Add("", "", "", "", "")
	f.Add("a@b.c", "x@y.z", "Ünïcödé", "<p>🎉</p>", "🎉")

	f.Fuzz(func(t *testing.T, to, cc, subject, html, text string) {
		msg := NewMessage().
			To(to).
			CC(cc).
			Subject(subject).
			HTML(html).
			Text(text).
			Header("X-Fuzz", "true").
			Build()

		if len(msg.To) != 1 || msg.To[0] != to {
			t.Errorf("To = %v, want [%q]", msg.To, to)
		}
		if msg.Subject != subject {
			t.Errorf("Subject = %q, want %q", msg.Subject, subject)
		}
	})
}
