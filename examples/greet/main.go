package main

import (
	"context"
	"log"
	"net/http"
	"text/template"

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
  {{- if .Output }}
  <p>Hello, {{ .Output.Name }}!</p>
  {{- end }}
</body>
</html>`))

func main() {
	mux := http.NewServeMux()
	html := forma.New(formago.New(mux), forma.Config{})

	// GET /greet — serve the empty form.
	forma.Register(html, forma.Operation[GreetOutput]{
		Method:   http.MethodGet,
		Path:     "/greet",
		Template: formTmpl,
	}, func(ctx context.Context, _ *GreetInput) (*GreetOutput, error) {
		return nil, nil
	})

	// POST /greet — validate input, then greet.
	// forma validates GreetInput tags (required, maxLength) before calling this handler.
	// On failure it re-renders the template with .Data.Errors populated.
	forma.Register(html, forma.Operation[GreetOutput]{
		Method:   http.MethodPost,
		Path:     "/greet",
		Template: formTmpl,
	}, func(ctx context.Context, i *GreetInput) (*GreetOutput, error) {
		return &GreetOutput{Name: i.Name}, nil
	})

	log.Println("example on http://localhost:8080/greet")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
