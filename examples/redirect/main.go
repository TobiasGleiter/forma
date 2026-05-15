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
type GreetInput struct {
	Name string `form:"name" required:"true" maxLength:"20"`
}

// NameInput is parsed from the path for the success page.
type NameInput struct {
	Name string `path:"name"`
}

// GreetOutput carries the result to the template.
type GreetOutput struct {
	Name string
}

var formTmpl = template.Must(template.New("form").Parse(`<!DOCTYPE html>
<html>
<head><title>Greet</title></head>
<body>
  <h1>Greet</h1>
  <form method="POST" action="/greet">
    <label>Name: <input name="name" value="{{ .Input.Name }}" placeholder="Your name"></label>
    {{- if .Errors.name }}
    <p style="color:red">{{ .Errors.name }}</p>
    {{- end }}
    <button>Submit</button>
  </form>
</body>
</html>`))

var successTmpl = template.Must(template.New("success").Parse(`<!DOCTYPE html>
<html>
<head><title>Greet</title></head>
<body>
  <h1>Hello, {{ .Output.Name }}!</h1>
  <a href="/greet">Greet someone else</a>
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

	// POST /greet — validate, then redirect to the success page.
	//
	// Redirect derives the URL from the handler output (dynamic).
	// Swap it for a static target with:
	//   RedirectURL: "/greet/success",
	forma.Register(html, forma.Operation[GreetOutput]{
		Method:   http.MethodPost,
		Path:     "/greet",
		Template: formTmpl,
		Redirect: func(o *GreetOutput) string { return "/greet/" + o.Name },
	}, func(ctx context.Context, i *GreetInput) (*GreetOutput, error) {
		return &GreetOutput{Name: i.Name}, nil
	})

	// GET /greet/{name} — success page after redirect.
	forma.Register(html, forma.Operation[GreetOutput]{
		Method:   http.MethodGet,
		Path:     "/greet/{name}",
		Template: successTmpl,
	}, func(ctx context.Context, i *NameInput) (*GreetOutput, error) {
		return &GreetOutput{Name: i.Name}, nil
	})

	log.Println("example on http://localhost:8080/greet")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
