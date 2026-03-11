package email

import (
	"context"
	"errors"
	"mime/multipart"
	"net/smtp"
	"os"
	"strings"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
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
	testkit.AssertErrorIs(t, err, ErrNoRecipients)
}

func TestSend_EmptyBody(t *testing.T) {
	m, _ := newMailerWithCapture(defaultCfg(), nil)
	err := m.Send(context.Background(), Message{To: []string{"a@b.com"}, Subject: "Hi"})
	testkit.AssertErrorIs(t, err, ErrEmptyBody)
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
	testkit.AssertEqual(t, cap.from, "no-reply@example.com")
	testkit.AssertEqual(t, cap.to, []string{"recv@example.com"})
	testkit.AssertContains(t, string(cap.msg), "Subject:")
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
	testkit.AssertContains(t, body, "text/html")
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
	testkit.AssertContains(t, body, "multipart/alternative")
	testkit.AssertContains(t, body, "text/plain")
	testkit.AssertContains(t, body, "text/html")
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
	testkit.AssertTrue(t, found)
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
	testkit.AssertTrue(t, found)
	// BCC should not appear in headers.
	testkit.AssertNotContains(t, string(cap.msg), "bcc@b.com")
}

func TestSend_SMTPError(t *testing.T) {
	m, _ := newMailerWithCapture(defaultCfg(), errors.New("connection refused"))
	err := m.Send(context.Background(), Message{
		To:      []string{"a@b.com"},
		Subject: "Hi",
		Text:    "body",
	})
	testkit.AssertError(t, err)
	testkit.AssertErrorContains(t, err, "connection refused")
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
	testkit.AssertContains(t, string(cap.msg), "X-Custom: value")
}

func TestSend_Addr(t *testing.T) {
	m, cap := newMailerWithCapture(defaultCfg(), nil)
	m.Send(context.Background(), Message{
		To: []string{"a@b.com"}, Subject: "s", Text: "t",
	})
	testkit.AssertEqual(t, cap.addr, "smtp.example.com:587")
}

func TestConfig_DefaultTimeout(t *testing.T) {
	m := NewMailer(Config{Host: "h", Port: 25})
	testkit.AssertTrue(t, m.cfg.Timeout > 0)
}

// ---- Template helpers ----

func TestParseHTMLString(t *testing.T) {
	tmpl, err := ParseHTMLString("welcome", `<h1>Hello {{.Name}}</h1>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testkit.AssertNotNil(t, tmpl)
}

func TestParseHTMLString_Invalid(t *testing.T) {
	_, err := ParseHTMLString("bad", `{{.Unclosed`)
	testkit.AssertError(t, err)
}

func TestRenderHTML(t *testing.T) {
	tmpl, _ := ParseHTMLString("welcome", `<h1>Hello {{.Name}}</h1>`)
	out, err := RenderHTML(tmpl, "welcome", map[string]string{"Name": "World"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testkit.AssertEqual(t, out, "<h1>Hello World</h1>")
}

func TestRenderHTML_MissingTemplate(t *testing.T) {
	tmpl, _ := ParseHTMLString("a", `hello`)
	_, err := RenderHTML(tmpl, "nonexistent", nil)
	testkit.AssertError(t, err)
}

func TestParseTextString(t *testing.T) {
	tmpl, err := ParseTextString("plain", `Hello {{.Name}}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testkit.AssertNotNil(t, tmpl)
}

func TestParseTextString_Invalid(t *testing.T) {
	_, err := ParseTextString("bad", `{{.Unclosed`)
	testkit.AssertError(t, err)
}

func TestRenderText(t *testing.T) {
	tmpl, _ := ParseTextString("plain", `Hello {{.Name}}`)
	out, err := RenderText(tmpl, "plain", map[string]string{"Name": "Jordan"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testkit.AssertEqual(t, out, "Hello Jordan")
}

func TestRenderText_MissingTemplate(t *testing.T) {
	tmpl, _ := ParseTextString("a", `hello`)
	_, err := RenderText(tmpl, "nonexistent", nil)
	testkit.AssertError(t, err)
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
	testkit.AssertContains(t, string(raw), "Subject:")
}

func TestParseHTMLTemplates_NoFiles(t *testing.T) {
	_, err := ParseHTMLTemplates("/nonexistent/file.html")
	testkit.AssertError(t, err)
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
	testkit.AssertContains(t, out, "Hello World")
}

func TestWriteQP_FailingWriter(t *testing.T) {
	err := writeQP(&failWriter{}, "hello")
	testkit.AssertError(t, err)
}

func TestEncodeBase64(t *testing.T) {
	data := []byte("hello world")
	got := encodeBase64(data)
	testkit.AssertNotEqual(t, got, "")
	// Verify it's valid base64 by checking for expected chars.
	for _, c := range got {
		isValid := (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') || c == '+' || c == '/' || c == '='
		if !isValid {
			t.Errorf("unexpected char %q in base64 output", c)
		}
	}
}

func TestEncodeBase64_Empty(t *testing.T) {
	testkit.AssertEqual(t, encodeBase64([]byte{}), "")
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
	testkit.AssertContains(t, string(raw), "quoted-printable")
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
	testkit.AssertContains(t, string(raw), "text/plain")
}

func TestBuildMessage_DateHeader(t *testing.T) {
	raw, err := buildMessage("from@x.com", Message{
		To: []string{"a@b.com"}, Subject: "d", Text: "body",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testkit.AssertContains(t, string(raw), "Date:")
}

func TestBuildMessage_MIMEVersion(t *testing.T) {
	raw, err := buildMessage("from@x.com", Message{
		To: []string{"a@b.com"}, Subject: "m", Text: "body",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testkit.AssertContains(t, string(raw), "MIME-Version: 1.0")
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
	testkit.AssertLen(t, cap.to, 2)
}

// failWriter always returns an error from Write.
type failWriter struct{}

func (f *failWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write: broken pipe")
}

func TestWriteQP_LongStringFailingWriter(t *testing.T) {
	// A string longer than 76 chars forces quotedprintable to flush during Write.
	longStr := strings.Repeat("a", 100)
	err := writeQP(&failWriter{}, longStr)
	testkit.AssertError(t, err)
}

func TestWritePart_CreatePartError(t *testing.T) {
	// A multipart.Writer on top of a failWriter will fail when creating a part.
	mw := multipart.NewWriter(&failWriter{})
	err := writePart(mw, "text/plain; charset=utf-8", "body text")
	testkit.AssertError(t, err)
	testkit.AssertErrorContains(t, err, "create part")
}

func TestSend_WithAllRecipientTypes(t *testing.T) {
	m, cap := newMailerWithCapture(defaultCfg(), nil)
	err := m.Send(context.Background(), Message{
		To:      []string{"to@example.com"},
		CC:      []string{"cc@example.com"},
		BCC:     []string{"bcc@example.com"},
		Subject: "All Recipients",
		HTML:    "<p>Hello</p>",
		Text:    "Hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// All three should appear in the envelope recipients.
	recipientSet := make(map[string]bool)
	for _, r := range cap.to {
		recipientSet[r] = true
	}
	testkit.AssertTrue(t, recipientSet["to@example.com"])
	testkit.AssertTrue(t, recipientSet["cc@example.com"])
	testkit.AssertTrue(t, recipientSet["bcc@example.com"])
}

func TestSend_CancelledContext(t *testing.T) {
	m, _ := newMailerWithCapture(defaultCfg(), nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := m.Send(ctx, Message{
		To:      []string{"a@b.com"},
		Subject: "Hi",
		Text:    "body",
	})
	testkit.AssertError(t, err)
	testkit.AssertErrorContains(t, err, "context canceled")
}

func TestSend_DeadlineExceeded(t *testing.T) {
	m, _ := newMailerWithCapture(defaultCfg(), nil)
	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()

	err := m.Send(ctx, Message{
		To:      []string{"a@b.com"},
		Subject: "Hi",
		Text:    "body",
	})
	testkit.AssertError(t, err)
	testkit.AssertErrorContains(t, err, "deadline exceeded")
}
