package email

import (
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestMessageBuilder_FullBuild(t *testing.T) {
	msg := NewMessage().
		To("alice@example.com", "bob@example.com").
		CC("cc@example.com").
		BCC("bcc@example.com").
		Subject("Hello").
		HTML("<h1>Hi</h1>").
		Text("Hi").
		Header("X-Priority", "1").
		Build()

	testkit.AssertEqual(t, len(msg.To), 2)
	testkit.AssertEqual(t, msg.To[0], "alice@example.com")
	testkit.AssertEqual(t, msg.To[1], "bob@example.com")
	testkit.AssertEqual(t, len(msg.CC), 1)
	testkit.AssertEqual(t, len(msg.BCC), 1)
	testkit.AssertEqual(t, msg.Subject, "Hello")
	testkit.AssertEqual(t, msg.HTML, "<h1>Hi</h1>")
	testkit.AssertEqual(t, msg.Text, "Hi")
	testkit.AssertEqual(t, msg.Headers["X-Priority"], "1")
}

func TestMessageBuilder_MinimalBuild(t *testing.T) {
	msg := NewMessage().
		To("user@example.com").
		Subject("Test").
		Text("body").
		Build()

	testkit.AssertEqual(t, len(msg.To), 1)
	testkit.AssertEqual(t, msg.Subject, "Test")
	testkit.AssertEqual(t, msg.Text, "body")
	testkit.AssertEqual(t, msg.HTML, "")
	testkit.AssertNil(t, msg.Headers)
}

func TestMessageBuilder_Chaining(t *testing.T) {
	b := NewMessage()
	result := b.To("a@b.com").Subject("s").Text("t")
	testkit.AssertTrue(t, result == b) // Same pointer returned.
}

func TestMessageBuilder_MultipleHeaders(t *testing.T) {
	msg := NewMessage().
		Header("X-One", "1").
		Header("X-Two", "2").
		Build()

	testkit.AssertEqual(t, msg.Headers["X-One"], "1")
	testkit.AssertEqual(t, msg.Headers["X-Two"], "2")
}

func TestMessageBuilder_EmptyBuild(t *testing.T) {
	msg := NewMessage().Build()
	testkit.AssertNil(t, msg.To)
	testkit.AssertEqual(t, msg.Subject, "")
}

func BenchmarkMessageBuilder(b *testing.B) {
	for b.Loop() {
		NewMessage().
			To("alice@example.com").
			Subject("Hello").
			HTML("<p>body</p>").
			Text("body").
			Header("X-Priority", "1").
			Build()
	}
}
