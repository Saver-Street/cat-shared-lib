// Package email provides a simple SMTP mailer with multipart MIME email
// construction and Go template rendering for HTML and plain-text bodies.
//
// Create a [Mailer] with [NewMailer] and a [Config] containing SMTP host,
// port, and sender address.  Build a [Message] with recipients, subject, and
// HTML or plain-text body, then call [Mailer.Send] to deliver it.  Messages
// are encoded with quoted-printable transfer encoding and proper MIME headers.
//
// Template helpers [ParseHTMLTemplates], [ParseHTMLString], and
// [ParseTextString] parse Go templates, while [RenderHTML] and [RenderText]
// execute them into strings suitable for use in [Message.HTML] and
// [Message.Text].
package email
