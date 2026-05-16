package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"html/template"
	"log"
	"net/http"

	"github.com/tobiasgleiter/forma"
	"github.com/tobiasgleiter/forma/adapters/formago"
)

// nonceKey is the context key used to pass the nonce between middleware and handler.
type nonceKey struct{}

// nonceTmpl renders a page that includes the nonce in both the CSP meta tag and the script tag.
// Template data shape: .Meta (PageMeta with Nonce), .Input, .Output.
var nonceTmpl = template.Must(template.New("nonce").Parse(`<!DOCTYPE html>
<html>
<head>
  <title>Nonce added via custom middleware</title>
  <meta http-equiv="Content-Security-Policy" content="script-src 'nonce-{{.Meta.Nonce}}'">
</head>
<body>
  <p>Open DevTools (F12) and inspect the &lt;script&gt; tag — its <code>nonce</code> attribute should match the CSP header.</p>
  <script nonce="{{.Meta.Nonce}}">console.log("nonce example: script executed")</script>
</body>
</html>`))

// NonceInput has no fields — this page requires no user input.
type NonceInput struct{}

// NonceOutput has no fields — the page only displays the nonce from .Meta.
type NonceOutput struct{}

// nonceMiddleware generates a random nonce per request and stores it in the context.
func nonceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, 16)
		rand.Read(b)
		nonce := base64.StdEncoding.EncodeToString(b)
		ctx := context.WithValue(r.Context(), nonceKey{}, nonce)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// PageMeta carries per-request metadata injected into every template via Config.Meta.
type PageMeta struct {
	Nonce string
}

func main() {
	mux := http.NewServeMux()
	html := forma.New(formago.New(mux), forma.Config{
		Meta: func(r *http.Request) any {
			nonce, _ := r.Context().Value(nonceKey{}).(string)
			return PageMeta{Nonce: nonce}
		},
	})

	// GET / — render the page; Config.Meta pulls the nonce from the request context.
	forma.Get(html, forma.Operation{
		Path:     "/",
		Template: nonceTmpl,
	}, func(ctx context.Context, i *NonceInput) (*NonceOutput, error) {
		return &NonceOutput{}, nil
	})

	log.Println("example on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nonceMiddleware(mux)))
}
