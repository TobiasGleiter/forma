package main

import (
	"context"
	"html/template"
	"log"
	"net/http"

	"github.com/tobiasgleiter/forma"
	"github.com/tobiasgleiter/forma/adapters/formago"
)

// LookupInput is parsed from the path parameter.
type LookupInput struct {
	Name string `path:"name"`
}

// LookupOutput holds the result passed to the template on success.
type LookupOutput struct {
	Name  string
	Email string
}

// ErrorData is passed to the error template via PageError.Data.
type ErrorData struct {
	Message string
}

var users = map[string]LookupOutput{
	"alice": {Name: "Alice", Email: "alice@example.com"},
	"bob":   {Name: "Bob", Email: "bob@example.com"},
}

var profileTmpl = template.Must(template.New("profile").Parse(`<!DOCTYPE html>
<html>
<head><title>Profile</title></head>
<body>
  <h1>{{ .Output.Name }}</h1>
  <p>{{ .Output.Email }}</p>
</body>
</html>`))

var errorTmpl = template.Must(template.New("error").Parse(`<!DOCTYPE html>
<html>
<head><title>Error</title></head>
<body>
  <p>{{ if .}}{{ .Message }}{{ else }}Something went wrong.{{ end }}</p>
</body>
</html>`))

func main() {
	mux := http.NewServeMux()
	html := forma.New(formago.New(mux), forma.Config{
		ErrorTemplate: errorTmpl,
	})

	// GET /users/{name} — look up a user by name.
	// Returns 404 with a human-readable message when the user is not found.
	forma.Get(html, forma.Operation{
		Path:     "/users/{name}",
		Template: profileTmpl,
	}, func(ctx context.Context, i *LookupInput) (*LookupOutput, error) {
		u, ok := users[i.Name]
		if !ok {
			return nil, &forma.PageError{
				Code: http.StatusNotFound,
				Data: ErrorData{Message: "User \"" + i.Name + "\" not found."},
			}
		}
		return &u, nil
	})

	log.Println("example on http://localhost:8080/users/alice")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
