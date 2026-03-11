package email

// MessageBuilder provides a fluent API for constructing [Message] values.
// Create one with [NewMessage] and chain methods before calling [MessageBuilder.Build].
type MessageBuilder struct {
	msg Message
}

// NewMessage returns a new MessageBuilder.
func NewMessage() *MessageBuilder {
	return &MessageBuilder{}
}

// To sets the recipient list.
func (b *MessageBuilder) To(addrs ...string) *MessageBuilder {
	b.msg.To = addrs
	return b
}

// CC sets the CC recipient list.
func (b *MessageBuilder) CC(addrs ...string) *MessageBuilder {
	b.msg.CC = addrs
	return b
}

// BCC sets the BCC recipient list.
func (b *MessageBuilder) BCC(addrs ...string) *MessageBuilder {
	b.msg.BCC = addrs
	return b
}

// Subject sets the email subject.
func (b *MessageBuilder) Subject(s string) *MessageBuilder {
	b.msg.Subject = s
	return b
}

// HTML sets the HTML body.
func (b *MessageBuilder) HTML(body string) *MessageBuilder {
	b.msg.HTML = body
	return b
}

// Text sets the plain-text body.
func (b *MessageBuilder) Text(body string) *MessageBuilder {
	b.msg.Text = body
	return b
}

// Header sets a custom MIME header.
func (b *MessageBuilder) Header(key, value string) *MessageBuilder {
	if b.msg.Headers == nil {
		b.msg.Headers = make(map[string]string)
	}
	b.msg.Headers[key] = value
	return b
}

// Build returns the constructed Message.
func (b *MessageBuilder) Build() Message {
	return b.msg
}
