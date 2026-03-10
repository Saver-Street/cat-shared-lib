package email

import (
	"context"
	"errors"
	"net/smtp"
	"os"
	"strings"
	"testing"
)

// capturedSend records the last call to the send function.
type capturedSend struct {
	addr string
	from string
	to   []string
	msg  []byte
	err  error
}

func newCapture(returnErr error) (*capturedSend, sendFunc) {
	c := &capturedSend{err: returnErr}
	fn := func(addr string, _ smtp.Auth, from string, to []string, msg []byte) error {
		c.addr = addr
		c.from = from
		c.to = to
		c.msg = msg
		return returnErr
	}
	return c, fn
}

func newMailerWithCapture(cfg Config, returnErr error) (*Mailer, *capturedSend) {
	m := NewMailer(cfg)
	cap, fn := newCapture(returnErr)
	m.send = fn
	return m, cap
}

func defaultCfg() Config {
	return Config{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
		From:     "no-reply@example.com",
	}
}

// ---- Send validation ----

func TestSend_NoRecipients(t *testing.T) {
	m, _ := newMailerWithCapture(defaultCfg(), nil)
	err := m.Send(context.Background(), Message{Subject: "Hi", Text: "body"})
	if !errors.Is(err, ErrNoRecipients) {
		t.Fatalf("expected ErrNoRecipients, got %v", err)
	}
}

func TestSend_EmptyBody(t *testing.T) {
	m, _ := newMailerWithCapture(defaultCfg(), nil)
	err := m.Send(context.Background(), Message{To: []string{"a@b.com"}, Subject: "Hi"})
	if !errors.Is(err, ErrEmptyBody) {
		t.Fatalf("expected ErrEmptyBody, got %v", err)
	}
}

func TestSend_PlainText(t *testing.T) {
	m, cap := newMailerWithCapture(defaultCfg(), nil)
	err := m.Send(context.Background(), Message{
		To:      []string{"recv@example.com"},
		Subject: "Hello",
		Text:    "Plain body",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.from != "no-reply@example.com" {
		t.Errorf("from = %q", cap.from)
	}
	if len(cap.to) != 1 || cap.to[0] != "recv@example.com" {
		t.Errorf("to = %v", cap.to)
	}
	if !strings.Contains(string(cap.msg), "Subject:") {
		t.Error("expected Subject header in message")
	}
}

func TestSend_HTMLOnly(t *testing.T) {
	m, cap := newMailerWithCapture(defaultCfg(), nil)
	err := m.Send(context.Background(), Message{
		To:      []string{"a@b.com"},
		Subject: "HTML mail",
		HTML:    "<h1>Hi</h1>",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body := string(cap.msg)
	if !strings.Contains(body, "text/html") {
		t.Error("expected text/html content-type")
	}
}

func TestSend_Multipart(t *testing.T) {
	m, cap := newMailerWithCapture(defaultCfg(), nil)
	err := m.Send(context.Background(), Message{
		To:      []string{"a@b.com"},
		Subject: "Multi",
		HTML:    "<p>hello</p>",
		Text:    "hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body := string(cap.msg)
	if !strings.Contains(body, "multipart/alternative") {
		t.Error("expected multipart/alternative")
	}
	if !strings.Contains(body, "text/plain") {
		t.Error("expected text/plain part")
	}
	if !strings.Contains(body, "text/html") {
		t.Error("expected text/html part")
	}
}

func TestSend_WithCC(t *testing.T) {
	m, cap := newMailerWithCapture(defaultCfg(), nil)
	err := m.Send(context.Background(), Message{
		To:      []string{"a@b.com"},
		CC:      []string{"cc@b.com"},
		Subject: "Test",
		Text:    "body",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, r := range cap.to {
		if r == "cc@b.com" {
			found = true
		}
	}
	if !found {
		t.Error("expected CC in recipients list")
	}
}

func TestSend_WithBCC(t *testing.T) {
	m, cap := newMailerWithCapture(defaultCfg(), nil)
	err := m.Send(context.Background(), Message{
		To:      []string{"a@b.com"},
		BCC:     []string{"bcc@b.com"},
		Subject: "Test",
		Text:    "body",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, r := range cap.to {
		if r == "bcc@b.com" {
			found = true
		}
	}
	if !found {
		t.Error("expected BCC in envelope recipients")
	}
	// BCC should not appear in headers.
	if strings.Contains(string(cap.msg), "bcc@b.com") {
		t.Error("BCC address should not appear in message headers")
	}
}

func TestSend_SMTPError(t *testing.T) {
	m, _ := newMailerWithCapture(defaultCfg(), errors.New("connection refused"))
	err := m.Send(context.Background(), Message{
		To:      []string{"a@b.com"},
		Subject: "Hi",
		Text:    "body",
	})
	if err == nil {
		t.Fatal("expected error from SMTP")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSend_ExtraHeaders(t *testing.T) {
	m, cap := newMailerWithCapture(defaultCfg(), nil)
	err := m.Send(context.Background(), Message{
		To:      []string{"a@b.com"},
		Subject: "Hi",
		Text:    "body",
		Headers: map[string]string{"X-Custom": "value"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(cap.msg), "X-Custom: value") {
		t.Error("expected custom header in message")
	}
}

func TestSend_Addr(t *testing.T) {
	m, cap := newMailerWithCapture(defaultCfg(), nil)
	m.Send(context.Background(), Message{
		To: []string{"a@b.com"}, Subject: "s", Text: "t",
	})
	if cap.addr != "smtp.example.com:587" {
		t.Errorf("unexpected addr %q", cap.addr)
	}
}

func TestConfig_DefaultTimeout(t *testing.T) {
	m := NewMailer(Config{Host: "h", Port: 25})
	if m.cfg.Timeout == 0 {
		t.Error("expected non-zero default timeout")
	}
}

// ---- Template helpers ----

func TestParseHTMLString(t *testing.T) {
	tmpl, err := ParseHTMLString("welcome", `<h1>Hello {{.Name}}</h1>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tmpl == nil {
		t.Fatal("expected non-nil template")
	}
}

func TestParseHTMLString_Invalid(t *testing.T) {
	_, err := ParseHTMLString("bad", `{{.Unclosed`)
	if err == nil {
		t.Fatal("expected error for invalid template")
	}
}

func TestRenderHTML(t *testing.T) {
	tmpl, _ := ParseHTMLString("welcome", `<h1>Hello {{.Name}}</h1>`)
	out, err := RenderHTML(tmpl, "welcome", map[string]string{"Name": "World"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "<h1>Hello World</h1>" {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestRenderHTML_MissingTemplate(t *testing.T) {
	tmpl, _ := ParseHTMLString("a", `hello`)
	_, err := RenderHTML(tmpl, "nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for missing template name")
	}
}

func TestParseTextString(t *testing.T) {
	tmpl, err := ParseTextString("plain", `Hello {{.Name}}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tmpl == nil {
		t.Fatal("expected non-nil template")
	}
}

func TestParseTextString_Invalid(t *testing.T) {
	_, err := ParseTextString("bad", `{{.Unclosed`)
	if err == nil {
		t.Fatal("expected error for invalid template")
	}
}

func TestRenderText(t *testing.T) {
	tmpl, _ := ParseTextString("plain", `Hello {{.Name}}`)
	out, err := RenderText(tmpl, "plain", map[string]string{"Name": "Jordan"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "Hello Jordan" {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestRenderText_MissingTemplate(t *testing.T) {
	tmpl, _ := ParseTextString("a", `hello`)
	_, err := RenderText(tmpl, "nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for missing template name")
	}
}

func TestBuildMessage_SubjectEncoding(t *testing.T) {
	msg := Message{
		To:      []string{"a@b.com"},
		Subject: "Héllo Wörld",
		Text:    "body",
	}
	raw, err := buildMessage("from@example.com", msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(raw), "Subject:") {
		t.Error("expected Subject header")
	}
}

func TestParseHTMLTemplates_NoFiles(t *testing.T) {
	_, err := ParseHTMLTemplates("/nonexistent/file.html")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestParseHTMLTemplates_ValidFile(t *testing.T) {
	f, err := os.CreateTemp("", "*.html")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(`{{define "welcome"}}<h1>Hello {{.}}</h1>{{end}}`)
	f.Close()

	tmpl, err := ParseHTMLTemplates(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, err := RenderHTML(tmpl, "welcome", "World")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if !strings.Contains(out, "Hello World") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestWriteQP_FailingWriter(t *testing.T) {
	err := writeQP(&failWriter{}, "hello")
	if err == nil {
		t.Fatal("expected error from failing writer")
	}
}

func TestEncodeBase64(t *testing.T) {
	data := []byte("hello world")
	got := encodeBase64(data)
	if got == "" {
		t.Fatal("expected non-empty base64 output")
	}
	// Verify it's valid base64 by checking for expected chars.
	for _, c := range got {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') || c == '+' || c == '/' || c == '=') {
			t.Errorf("unexpected char %q in base64 output", c)
		}
	}
}

func TestEncodeBase64_Empty(t *testing.T) {
	if encodeBase64([]byte{}) != "" {
		t.Error("expected empty string for empty input")
	}
}

func TestSend_NilAuth(t *testing.T) {
	// No username → nil auth path.
	cfg := Config{Host: "h", Port: 25, From: "from@example.com"}
	m, cap := newMailerWithCapture(cfg, nil)
	err := m.Send(context.Background(), Message{
		To: []string{"a@b.com"}, Subject: "s", Text: "t",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = cap
}

func TestBuildMessage_HTMLOnly_Content(t *testing.T) {
	raw, err := buildMessage("from@x.com", Message{
		To:      []string{"a@b.com"},
		Subject: "hi",
		HTML:    "<b>hello</b>",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(raw), "quoted-printable") {
		t.Error("expected quoted-printable encoding")
	}
}

func TestBuildMessage_TextOnly_Content(t *testing.T) {
	raw, err := buildMessage("from@x.com", Message{
		To:      []string{"a@b.com"},
		Subject: "hi",
		Text:    "plain body",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(raw), "text/plain") {
		t.Error("expected text/plain content-type")
	}
}

func TestBuildMessage_DateHeader(t *testing.T) {
	raw, err := buildMessage("from@x.com", Message{
		To: []string{"a@b.com"}, Subject: "d", Text: "body",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(raw), "Date:") {
		t.Error("expected Date header")
	}
}

func TestBuildMessage_MIMEVersion(t *testing.T) {
	raw, err := buildMessage("from@x.com", Message{
		To: []string{"a@b.com"}, Subject: "m", Text: "body",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(raw), "MIME-Version: 1.0") {
		t.Error("expected MIME-Version header")
	}
}

func TestSend_MultipleRecipients(t *testing.T) {
	m, cap := newMailerWithCapture(defaultCfg(), nil)
	err := m.Send(context.Background(), Message{
		To:      []string{"a@b.com", "c@d.com"},
		Subject: "Hi",
		Text:    "body",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cap.to) != 2 {
		t.Errorf("expected 2 recipients, got %d", len(cap.to))
	}
}

// failWriter always returns an error from Write.
type failWriter struct{}

func (f *failWriter) Write(_ []byte) (int, error) {
return 0, errors.New("write: broken pipe")
}
