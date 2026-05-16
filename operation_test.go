package forma

import (
	"html/template"
	"net/http"
	"testing"
)

func TestPageError_Error(t *testing.T) {
	pe := &PageError{Code: http.StatusNotFound}
	if pe.Error() != "Not Found" {
		t.Fatalf("expected %q, got %q", "Not Found", pe.Error())
	}
}

func TestOperation_SuccessCode(t *testing.T) {
	t.Run("defaults to 200", func(t *testing.T) {
		op := Operation[struct{}]{}
		if op.successCode() != http.StatusOK {
			t.Fatalf("expected 200, got %d", op.successCode())
		}
	})
	t.Run("returns override", func(t *testing.T) {
		op := Operation[struct{}]{SuccessCode: http.StatusCreated}
		if op.successCode() != http.StatusCreated {
			t.Fatalf("expected 201, got %d", op.successCode())
		}
	})
}

func TestOperation_ValidationCode(t *testing.T) {
	t.Run("defaults to 422", func(t *testing.T) {
		op := Operation[struct{}]{}
		if op.validationCode() != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", op.validationCode())
		}
	})
	t.Run("returns override", func(t *testing.T) {
		op := Operation[struct{}]{ValidationCode: http.StatusBadRequest}
		if op.validationCode() != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", op.validationCode())
		}
	})
}

func TestOperation_RedirectURL(t *testing.T) {
	type O struct{}
	out := &O{}

	t.Run("static URL returned", func(t *testing.T) {
		op := Operation[O]{RedirectURL: "/list"}
		if op.redirectURL(out) != "/list" {
			t.Fatalf("expected \"/list\", got %q", op.redirectURL(out))
		}
	})
	t.Run("Redirect func takes priority", func(t *testing.T) {
		op := Operation[O]{
			RedirectURL: "/list",
			Redirect:    func(*O) string { return "/detail" },
		}
		if op.redirectURL(out) != "/detail" {
			t.Fatalf("expected \"/detail\", got %q", op.redirectURL(out))
		}
	})
	t.Run("Redirect func returning empty skips redirect", func(t *testing.T) {
		op := Operation[O]{Redirect: func(*O) string { return "" }}
		if op.redirectURL(out) != "" {
			t.Fatalf("expected empty string, got %q", op.redirectURL(out))
		}
	})
}

func TestOperation_Entrypoint(t *testing.T) {
	base := template.Must(template.New("base").Parse(`base`))
	multi := template.Must(template.New("root").Parse(`root{{define "layout"}}layout{{end}}`))

	t.Run("no TemplateName returns Template", func(t *testing.T) {
		op := Operation[struct{}]{Template: base}
		if op.entrypoint() != base {
			t.Fatal("expected Template to be returned when TemplateName is empty")
		}
	})
	t.Run("TemplateName resolves named sub-template", func(t *testing.T) {
		op := Operation[struct{}]{Template: multi, TemplateName: "layout"}
		if op.entrypoint().Name() != "layout" {
			t.Fatalf("expected \"layout\", got %q", op.entrypoint().Name())
		}
	})
	t.Run("unknown TemplateName falls back to Template", func(t *testing.T) {
		op := Operation[struct{}]{Template: base, TemplateName: "nonexistent"}
		if op.entrypoint() != base {
			t.Fatal("expected fallback to Template for unknown TemplateName")
		}
	})
}
