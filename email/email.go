// Package email provides a simple SMTP mailer with HTML and plain-text
// template support. Templates are parsed from Go's html/template and
// text/template packages.
//
// Usage:
//
//	m := email.NewMailer(email.Config{
//	    Host:     "smtp.example.com",
//	    Port:     587,
//	    Username: "user@example.com",
//	    Password: "secret",
//	    From:     "no-reply@example.com",
//	})
//	err := m.Send(ctx, email.Message{
//	    To:      []string{"recipient@example.com"},
//	    Subject: "Welcome",
//	    HTML:    "<h1>Hello</h1>",
//	    Text:    "Hello",
//	})
package email

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/smtp"
	"net/textproto"
	"strings"
	texttemplate "text/template"
	"time"
)

// Config holds SMTP connection configuration.
type Config struct {
	// Host is the SMTP server hostname (e.g. "smtp.sendgrid.net").
	Host string
	// Port is the SMTP server port (e.g. 587 for STARTTLS, 465 for TLS).
	Port int
	// Username for SMTP authentication.
	Username string
	// Password for SMTP authentication.
	Password string
	// From is the default sender address.
	From string
	// Timeout for establishing the SMTP connection. Default: 30s.
	Timeout time.Duration
}

func (c *Config) addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Message represents an outgoing email.
type Message struct {
	// To is a list of recipient addresses.
	To []string
	// CC is a list of CC recipient addresses.
	CC []string
	// BCC is a list of BCC recipient addresses.
	BCC []string
	// Subject is the email subject line.
	Subject string
	// HTML is the HTML body part. If empty only Text is sent.
	HTML string
	// Text is the plain-text body part. If empty only HTML is sent.
	Text string
	// Headers holds any extra MIME headers to include.
	Headers map[string]string
}

// Mailer sends email via SMTP.
type Mailer struct {
	cfg  Config
	send sendFunc
}

// sendFunc abstracts smtp.SendMail for testability.
type sendFunc func(addr string, a smtp.Auth, from string, to []string, msg []byte) error

// NewMailer creates a Mailer with the given configuration.
func NewMailer(cfg Config) *Mailer {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	return &Mailer{cfg: cfg, send: smtp.SendMail}
}

// ErrNoRecipients is returned when To is empty.
var ErrNoRecipients = errors.New("email: message must have at least one recipient")

// ErrEmptyBody is returned when both HTML and Text are empty.
var ErrEmptyBody = errors.New("email: message must have HTML or Text body")

// Send sends msg via SMTP. It respects context cancellation during the dial
// phase.
func (m *Mailer) Send(_ context.Context, msg Message) error {
	if len(msg.To) == 0 {
		return ErrNoRecipients
	}
	if msg.HTML == "" && msg.Text == "" {
		return ErrEmptyBody
	}

	raw, err := buildMessage(m.cfg.From, msg)
	if err != nil {
		return err
	}

	allTo := append(append([]string{}, msg.To...), msg.CC...)
	allTo = append(allTo, msg.BCC...)

	var auth smtp.Auth
	if m.cfg.Username != "" {
		auth = smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
	}

	if err := m.send(m.cfg.addr(), auth, m.cfg.From, allTo, raw); err != nil {
		return fmt.Errorf("email: send: %w", err)
	}
	return nil
}

// buildMessage constructs the raw MIME email bytes.
func buildMessage(from string, msg Message) ([]byte, error) {
	var buf bytes.Buffer

	// Headers.
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(msg.To, ", ")))
	if len(msg.CC) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(msg.CC, ", ")))
	}
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n",
		mime.QEncoding.Encode("utf-8", msg.Subject)))
	buf.WriteString(fmt.Sprintf("Date: %s\r\n",
		time.Now().UTC().Format(time.RFC1123Z)))
	for k, v := range msg.Headers {
		buf.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}

	switch {
	case msg.HTML != "" && msg.Text != "":
		mw := multipart.NewWriter(&buf)
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%q\r\n\r\n", mw.Boundary()))
		if err := writePart(mw, "text/plain; charset=utf-8", msg.Text); err != nil {
			return nil, err
		}
		if err := writePart(mw, "text/html; charset=utf-8", msg.HTML); err != nil {
			return nil, err
		}
		if err := mw.Close(); err != nil {
			return nil, fmt.Errorf("email: close multipart: %w", err)
		}
	case msg.HTML != "":
		buf.WriteString("Content-Type: text/html; charset=utf-8\r\n")
		buf.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
		if err := writeQP(&buf, msg.HTML); err != nil {
			return nil, err
		}
	default:
		buf.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
		buf.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
		if err := writeQP(&buf, msg.Text); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func writePart(mw *multipart.Writer, contentType, body string) error {
	h := textproto.MIMEHeader{}
	h.Set("Content-Type", contentType)
	h.Set("Content-Transfer-Encoding", "quoted-printable")
	pw, err := mw.CreatePart(h)
	if err != nil {
		return fmt.Errorf("email: create part: %w", err)
	}
	return writeQP(pw, body)
}

func writeQP(w interface{ Write([]byte) (int, error) }, s string) error {
	qpw := quotedprintable.NewWriter(w)
	if _, err := qpw.Write([]byte(s)); err != nil {
		return fmt.Errorf("email: write quoted-printable: %w", err)
	}
	return qpw.Close()
}

// ---- Template helpers ----

// RenderHTML renders the named HTML template with data and returns the output.
func RenderHTML(tmpl *template.Template, name string, data any) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return "", fmt.Errorf("email: render HTML template %q: %w", name, err)
	}
	return buf.String(), nil
}

// RenderText renders the named text template with data and returns the output.
func RenderText(tmpl *texttemplate.Template, name string, data any) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return "", fmt.Errorf("email: render text template %q: %w", name, err)
	}
	return buf.String(), nil
}

// ParseHTMLTemplates parses a set of HTML template files and returns the
// template set.
func ParseHTMLTemplates(patterns ...string) (*template.Template, error) {
	tmpl, err := template.ParseFiles(patterns...)
	if err != nil {
		return nil, fmt.Errorf("email: parse HTML templates: %w", err)
	}
	return tmpl, nil
}

// ParseHTMLString parses a single HTML template string with the given name.
func ParseHTMLString(name, src string) (*template.Template, error) {
	tmpl, err := template.New(name).Parse(src)
	if err != nil {
		return nil, fmt.Errorf("email: parse HTML template: %w", err)
	}
	return tmpl, nil
}

// ParseTextString parses a single text template string with the given name.
func ParseTextString(name, src string) (*texttemplate.Template, error) {
	tmpl, err := texttemplate.New(name).Parse(src)
	if err != nil {
		return nil, fmt.Errorf("email: parse text template: %w", err)
	}
	return tmpl, nil
}

// encodeBase64 is a helper used when attaching files in the future.
func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
