package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/tobiasgleiter/forma"
	"github.com/tobiasgleiter/forma/adapters/formago"
)

// SearchInput is parsed from the GET query string.
type SearchInput struct {
	Search string `query:"q"`
}

// SearchOutput holds the filtered results passed to the template.
type SearchOutput struct {
	Items []string
}

var catalog = []string{"Go", "Rust", "TypeScript", "PostgreSQL", "SQLite", "Redis", "Docker", "Kubernetes"}

var searchTmpl = template.Must(template.New("search").Parse(`<!DOCTYPE html>
<html>
<head><title>Search</title></head>
<body>
  <h1>Catalog Search</h1>
  <form method="GET" action="/">
    <input name="q" value="{{ .Input.Search }}" placeholder="Search...">
    <button type="submit">Search</button>
  </form>
  <ul>
    {{- range .Output.Items }}
    <li>{{ . }}</li>
    {{- end }}
  </ul>
</body>
</html>`))

func main() {
	mux := http.NewServeMux()
	html := forma.New(formago.New(mux), forma.Config{})

	// GET / — render the search form and filter results by the "q" query param.
	forma.Register(html, forma.Operation[SearchOutput]{
		Method:   http.MethodGet,
		Path:     "/",
		Template: searchTmpl,
	}, func(ctx context.Context, i *SearchInput) (*SearchOutput, error) {
		var results []string
		for _, name := range catalog {
			if strings.Contains(strings.ToLower(name), strings.ToLower(i.Search)) {
				results = append(results, name)
			}
		}
		return &SearchOutput{Items: results}, nil
	})

	log.Println("example on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
