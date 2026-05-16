package main

import (
	"context"
	"html/template"
	"log"
	"net/http"

	"github.com/tobiasgleiter/forma"
	"github.com/tobiasgleiter/forma/adapters/formago"
)

// GreetInput is parsed from the POST form body.
// Validation tags run automatically before the handler is called.
type GreetInput struct {
	Name string `form:"name" required:"true" maxLength:"20"`
}

// GreetOutput holds the result passed to the template on success.
type GreetOutput struct {
	Name string
}

// Template data shape: .Input, .Output, .Errors (map field→Name), .URL.
var formTmpl = template.Must(template.New("greet").Parse(`<!DOCTYPE html>
<html>
<head><title>Greet</title></head>
<body>
  <form method="POST" action="/greet">
    <label>
      Name:
      <input name="name" value="{{ .Input.Name }}">
    </label>
    {{- if .Errors.name }}
    <p style="color:red">{{ .Errors.name }}</p>
    {{- end }}
    <button type="submit">Greet</button>
  </form>
</body>
</html>`))

var successTmpl = template.Must(template.New("success").Parse(`<!DOCTYPE html>
<html>
<head><title>Greet</title></head>
<body>
  <p>Greeted successfully!</p>
  <a href="/greet">Greet someone else</a>
</body>
</html>`))

func main() {
	mux := http.NewServeMux()
	html := forma.New(formago.New(mux), forma.Config{})

	// GET /greet — serve the empty form.
	forma.Get(html, forma.Operation{
		Path:     "/greet",
		Template: formTmpl,
	}, func(ctx context.Context, _ *GreetInput) (*GreetOutput, error) {
		return nil, nil
	})

	// POST /greet — validate input, then redirect to the success page.
	// forma validates GreetInput tags (required, maxLength) before calling this handler.
	// On failure it re-renders the template with .Errors populated.
	forma.Post(html, forma.Operation{
		Path:        "/greet",
		Template:    formTmpl,
		RedirectURL: "/greet/success",
	}, func(ctx context.Context, i *GreetInput) (*GreetOutput, error) {
		return &GreetOutput{Name: i.Name}, nil
	})

	// GET /greet/success — success page after redirect.
	forma.Get(html, forma.Operation{
		Path:     "/greet/success",
		Template: successTmpl,
	}, func(ctx context.Context, _ *GreetOutput) (*GreetOutput, error) {
		return nil, nil
	})

	log.Println("example on http://localhost:8080/greet")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
