package forma

import (
	"html/template"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Renderer == nil {
		t.Fatal("expected Renderer to be set")
	}
	if cfg.Logger == nil {
		t.Fatal("expected Logger to be set")
	}
}

func TestHTMLRenderer_Render(t *testing.T) {
	re := newHTMLRenderer(slog.Default())
	tmpl := template.Must(template.New("page").Parse(`<p>{{.Message}}</p>`))

	t.Run("nil data uses status text", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		if err := re.Render(rec, req, http.StatusNotFound, tmpl, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "Not Found") {
			t.Fatalf("expected status text in body, got %q", rec.Body.String())
		}
	})

	t.Run("provided data rendered", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		data := struct{ Message string }{Message: "hello"}

		if err := re.Render(rec, req, http.StatusOK, tmpl, data); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(rec.Body.String(), "hello") {
			t.Fatalf("expected \"hello\" in body, got %q", rec.Body.String())
		}
	})

	t.Run("content-type header set", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		re.Render(rec, req, http.StatusOK, tmpl, struct{ Message string }{})
		if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
			t.Fatalf("expected text/html content-type, got %q", ct)
		}
	})

	t.Run("template error returned", func(t *testing.T) {
		broken := template.Must(template.New("bad").Parse(`{{.NoSuchField.Nested}}`))
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		err := re.Render(rec, req, http.StatusOK, broken, struct{}{})
		if err == nil {
			t.Fatal("expected error for broken template")
		}
	})
}
