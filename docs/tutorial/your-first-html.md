---
description: Learn how to create your first Forma HTML page.
---

# Your First HTML Page

Let's build a simple page that greets people. We will take the person's name as a form input and render an HTML page with a greeting message. Here's the high-level design:

```
Request:
GET  /greet  → render empty form
POST /greet  → validate input, render greeting
```

## Input & Output

Start by making a new file `main.go` and adding the input and output models for the greet operation:

```go title="main.go" linenums="1"
package main

// GreetInput is parsed from the POST form body.
// Validation tags run automatically before the handler is called.
type GreetInput struct {
	Name string `form:"name" required:"true" max:"20"`
}

// GreetOutput holds the result passed to the template on success.
type GreetOutput struct {
	Name string
}
```

The `form` tag tells Forma which form field to bind. The `required` and `max` tags are validated automatically before your handler is called.

You should now have a directory structure that looks like:

```
my-app/
  |-- go.mod
  |-- go.sum
  |-- main.go
```

## Template

Add an HTML template that renders the form and, when output is present, displays the greeting:

```go title="main.go" linenums="1"
package main

import "html/template"

// ...

// Template data shape: .Input, .Output, .Errors (map field to message), .URL.
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
```

The template receives `.Input`, `.Output`, and `.Errors`. On a validation failure Forma re-renders the template with `.Errors` populated so the user sees inline error messages.

## Router & App

Create the router, wrap it with Forma, and register both operations:

```go title="main.go" linenums="1"
package main

import (
	"context"
	"html/template"
	"log"
	"net/http"

	"github.com/tobiasgleiter/forma"
	"github.com/tobiasgleiter/forma/adapters/formago"
)

type GreetInput struct {
	Name string `form:"name" required:"true" max:"20"`
}

type GreetOutput struct {
	Name string
}

var formTmpl = template.Must(template.New("greet").Parse(`...`))

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

	// POST /greet — validate input, then greet.
	// Forma validates GreetInput tags (required, max) before calling this handler.
	// On failure it re-renders the template with .Errors populated.
	forma.Post(html, forma.Operation{
		Path:     "/greet",
		Template: formTmpl,
	}, func(ctx context.Context, i *GreetInput) (*GreetOutput, error) {
		return &GreetOutput{Name: i.Name}, nil
	})

	log.Println("example on http://localhost:8080/greet")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
```

Download dependencies for `main.go`:

```bash
$ go mod tidy
```

Congratulations! This is a fully functional Forma page!

## Running the App

Start the server:

```bash
$ go run .
```

Open [http://localhost:8080/greet](http://localhost:8080/greet) in your browser. You should see a form with a name field. Submit it to see the greeting, or leave it empty to see the validation error.

## Review

Congratulations! You just learned:

- Creating Forma input and output models with validation tags
- Writing an HTML template using `.Input`, `.Output`, and `.Errors`
- Registering GET and POST operations with `forma.Get` and `forma.Post`
- How Forma automatically validates input and re-renders on failure

Read on to learn how to level up your pages with even more features.
