package main

import (
	"context"
	"embed"
	"html/template"
	"log"
	"net/http"

	"github.com/tobiasgleiter/forma"
	"github.com/tobiasgleiter/forma/adapters/formago"
)

//go:embed *.html
var files embed.FS

type Page struct{}

// template.New("root").ParseFS(...) leaves "root" unparsed (Tree == nil), so
// Config.TemplateName or Operation.TemplateName must name the defined entry-point block.
var homeTmpl = template.Must(template.New("root").ParseFS(files, "layout.html", "home.html"))
var aboutTmpl = template.Must(template.New("root").ParseFS(files, "layout.html", "about.html"))
var plainTmpl = template.Must(template.New("root").ParseFS(files, "layout.html", "plain.html"))

func main() {
	mux := http.NewServeMux()
	html := forma.New(formago.New(mux), forma.Config{
		TemplateName: "base", // default entry-point for all routes
	})

	forma.Get(html, forma.Operation{
		Path:     "/",
		Template: homeTmpl,
	}, func(_ context.Context, _ *Page) (*Page, error) {
		return nil, nil
	})

	forma.Get(html, forma.Operation{
		Path:     "/about",
		Template: aboutTmpl,
	}, func(_ context.Context, _ *Page) (*Page, error) {
		return nil, nil
	})

	// TemplateName overrides the global "base" for this route alone.
	forma.Get(html, forma.Operation{
		Path:         "/plain",
		Template:     plainTmpl,
		TemplateName: "minimal",
	}, func(_ context.Context, _ *Page) (*Page, error) {
		return nil, nil
	})

	log.Println("example on http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
