package main

import (
	"context"
	"errors"
	"html/template"
	"log"
	"net/http"

	"github.com/tobiasgleiter/forma"
	"github.com/tobiasgleiter/forma/adapters/formago"
)

// RegisterInput is parsed from the POST form body.
type RegisterInput struct {
	Username string `form:"username" required:"true" min:"3" max:"30"`
}

// RegisterOutput carries the result to the success template.
type RegisterOutput struct {
	Username string
}

var formTmpl = template.Must(template.New("form").Parse(`<!DOCTYPE html>
<html>
<head><title>Register</title></head>
<body>
{{- if .Output }}
  <h1>Welcome, {{ .Output.Username }}!</h1>
  <p>Your account has been created.</p>
{{- else }}
  <h1>Create account</h1>
  <form method="POST" action="/register">
    <p>
      <label>Username<br>
        <input name="username" value="{{ .Input.Username }}">
      </label>
      {{- if .Errors.username }}
      <br><span style="color:red">{{ .Errors.username }}</span>
      {{- end }}
    </p>
    <button>Register</button>
  </form>
{{- end }}
</body>
</html>`))

// takenUsernames simulates a set of already-registered usernames.
var takenUsernames = map[string]bool{"admin": true, "root": true}

func register(ctx context.Context, i *RegisterInput) (*RegisterOutput, error) {
	// Simulate a DB uniqueness check (e.g. a UNIQUE constraint violation).
	if takenUsernames[i.Username] {
		return nil, &forma.ValidationError{
			Field: map[string]string{
				"username": "That username is already taken.",
			},
		}
	}

	// Simulate persisting the new user.
	takenUsernames[i.Username] = true
	return &RegisterOutput{Username: i.Username}, nil
}

// ErrAlreadyExists is the sentinel your real DB layer would return.
var ErrAlreadyExists = errors.New("already exists")

func main() {
	mux := http.NewServeMux()
	html := forma.New(formago.New(mux), forma.Config{})

	// GET /register — serve the empty form.
	forma.Get(html, forma.Operation{
		Path:     "/register",
		Template: formTmpl,
	}, func(ctx context.Context, _ *RegisterInput) (*RegisterOutput, error) {
		return nil, nil
	})

	// POST /register — tag validation runs first; on success the handler checks
	// the uniqueness constraint and returns a ValidationError if it fires.
	// The framework re-renders formTmpl with Errors populated in both cases.
	// On success the same template renders the confirmation view via .Output.
	forma.Post(html, forma.Operation{
		Path:     "/register",
		Template: formTmpl,
	}, register)

	log.Println("example on http://localhost:8080/register")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
