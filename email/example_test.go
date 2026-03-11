package email_test

import (
	"fmt"
	"html/template"
	texttemplate "text/template"

	"github.com/Saver-Street/cat-shared-lib/email"
)

func ExampleParseHTMLString() {
	tmpl, err := email.ParseHTMLString("welcome", "<h1>Hello {{.Name}}</h1>")
	if err != nil {
		panic(err)
	}
	out, err := email.RenderHTML(tmpl, "welcome", map[string]string{"Name": "Alice"})
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
	// Output:
	// <h1>Hello Alice</h1>
}

func ExampleParseTextString() {
	tmpl, err := email.ParseTextString("note", "Hi {{.Name}}, your code is {{.Code}}.")
	if err != nil {
		panic(err)
	}
	out, err := email.RenderText(tmpl, "note", map[string]string{
		"Name": "Bob",
		"Code": "ABC123",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
	// Output:
	// Hi Bob, your code is ABC123.
}

func ExampleRenderHTML() {
	tmpl := template.Must(template.New("greet").Parse("<p>Welcome, {{.}}!</p>"))
	out, _ := email.RenderHTML(tmpl, "greet", "Jordan")
	fmt.Println(out)
	// Output:
	// <p>Welcome, Jordan!</p>
}

func ExampleRenderText() {
	tmpl := texttemplate.Must(texttemplate.New("greet").Parse("Welcome, {{.}}!"))
	out, _ := email.RenderText(tmpl, "greet", "Jordan")
	fmt.Println(out)
	// Output:
	// Welcome, Jordan!
}

func ExampleNewMessage() {
	msg := email.NewMessage().
		To("alice@example.com", "bob@example.com").
		Subject("Welcome").
		HTML("<h1>Hello</h1>").
		Text("Hello").
		Header("X-Priority", "1").
		Build()

	fmt.Println(msg.Subject)
	fmt.Println(msg.To)
	// Output:
	// Welcome
	// [alice@example.com bob@example.com]
}

func ExampleNewMailer() {
	m := email.NewMailer(email.Config{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user@example.com",
		Password: "secret",
		From:     "no-reply@example.com",
	})
	// m is ready to use with m.Send(ctx, msg)
	fmt.Printf("%T\n", m)
	// Output:
	// *email.Mailer
}
