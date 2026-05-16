package forma

import (
	"context"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

// errReader is an io.Reader that always returns an error, used to force ParseForm to fail.
type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("read error") }

// dispatchRouter captures the registered handler and exposes it for direct invocation.
type dispatchRouter struct{ h http.HandlerFunc }

func (r *dispatchRouter) Handle(_, _ string, h http.HandlerFunc) { r.h = h }

func (r *dispatchRouter) serve(rec *httptest.ResponseRecorder, req *http.Request) {
	r.h(rec, req)
}

func newTestHTML(t *testing.T) (*HTML, *dispatchRouter) {
	t.Helper()
	router := &dispatchRouter{}
	errorTmpl := template.Must(template.New("error").Parse(`error`))
	h := New(router, Config{ErrorTemplate: errorTmpl})
	return h, router
}

func TestRegister_ParseInputFailureReturns400(t *testing.T) {
	type Input struct{}
	type Output struct{}

	h, router := newTestHTML(t)
	Post(h, Operation{Path: "/", Template: template.Must(template.New("page").Parse(`ok`))},
		func(_ context.Context, in *Input) (*Output, error) { return &Output{}, nil },
	)

	req := httptest.NewRequest(http.MethodPost, "/", errReader{})
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	router.serve(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestRegister_IncompleteTemplatePanics(t *testing.T) {
	// Simulates the common layout pattern: template.New("root").ParseGlob(...)
	// where all files only contain {{define}} blocks. The root template is never
	// parsed itself, so its Tree is nil.
	tmpl := template.New("root")
	template.Must(tmpl.New("base").Parse(`<html>{{.}}</html>`))

	type Input struct{}
	type Output struct{}
	h, _ := newTestHTML(t)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic, got none")
		}
		msg, _ := r.(string)
		if !strings.Contains(msg, "incomplete") {
			t.Fatalf("unexpected panic message: %s", msg)
		}
	}()

	Get(h, Operation{Path: "/", Template: tmpl},
		func(_ context.Context, in *Input) (*Output, error) { return nil, nil },
	)
}

func TestRegister_IncompleteTemplateWithNameDoesNotPanic(t *testing.T) {
	tmpl := template.New("root")
	template.Must(tmpl.New("base").Parse(`<html>{{.}}</html>`))

	type Input struct{}
	type Output struct{}
	h, _ := newTestHTML(t)

	Get(h, Operation{Path: "/", Template: tmpl, TemplateName: "base"},
		func(_ context.Context, in *Input) (*Output, error) { return nil, nil },
	)
}

func TestRegister_GETRendersOutput(t *testing.T) {
	type Input struct {
		Name string `query:"name"`
	}
	type Output struct{ Message string }

	tmpl := template.Must(template.New("page").Parse(`{{.Output.Message}}`))
	h, router := newTestHTML(t)
	Get(h, Operation{Path: "/", Template: tmpl},
		func(_ context.Context, in *Input) (*Output, error) {
			return &Output{Message: "hello " + in.Name}, nil
		},
	)

	rec := httptest.NewRecorder()
	router.serve(rec, httptest.NewRequest(http.MethodGet, "/?name=world", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "hello world") {
		t.Fatalf("expected output in body, got %q", rec.Body.String())
	}
}

func TestRegister_POSTValidInputRendersSuccess(t *testing.T) {
	type Input struct {
		Name string `form:"name" required:""`
	}
	type Output struct{ Message string }

	tmpl := template.Must(template.New("page").Parse(`{{.Output.Message}}`))
	h, router := newTestHTML(t)
	Post(h, Operation{Path: "/", Template: tmpl},
		func(_ context.Context, in *Input) (*Output, error) {
			return &Output{Message: "saved " + in.Name}, nil
		},
	)

	body := url.Values{"name": {"alice"}}.Encode()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	router.serve(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "saved alice") {
		t.Fatalf("expected output in body, got %q", rec.Body.String())
	}
}

func TestRegister_POSTInvalidInputReturns422(t *testing.T) {
	type Input struct {
		Name string `form:"name" required:""`
	}
	type Output struct{}

	tmpl := template.Must(template.New("page").Parse(`{{range $k,$v := .Errors}}{{$k}}:{{$v}}{{end}}`))
	h, router := newTestHTML(t)
	handlerCalled := false
	Post(h, Operation{Path: "/", Template: tmpl},
		func(_ context.Context, in *Input) (*Output, error) {
			handlerCalled = true
			return &Output{}, nil
		},
	)

	body := url.Values{"name": {""}}.Encode()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	router.serve(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
	if handlerCalled {
		t.Fatal("expected handler not to be called on validation failure")
	}
	if !strings.Contains(rec.Body.String(), "name") {
		t.Fatalf("expected field error in body, got %q", rec.Body.String())
	}
}

func TestRegister_GETSkipsValidation(t *testing.T) {
	type Input struct {
		Name string `form:"name" required:""`
	}
	type Output struct{}

	tmpl := template.Must(template.New("page").Parse(`ok`))
	h, router := newTestHTML(t)
	handlerCalled := false
	Get(h, Operation{Path: "/", Template: tmpl},
		func(_ context.Context, in *Input) (*Output, error) {
			handlerCalled = true
			return &Output{}, nil
		},
	)

	rec := httptest.NewRecorder()
	router.serve(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if !handlerCalled {
		t.Fatal("expected handler to be called — GET should skip validation")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRegister_HandlerPageErrorRendersCorrectStatus(t *testing.T) {
	type Input struct{}
	type Output struct{}

	h, router := newTestHTML(t)
	Get(h, Operation{Path: "/", Template: template.Must(template.New("page").Parse(`ok`))},
		func(_ context.Context, in *Input) (*Output, error) {
			return nil, &PageError{Code: http.StatusNotFound}
		},
	)

	rec := httptest.NewRecorder()
	router.serve(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestRegister_HandlerGenericErrorReturns500(t *testing.T) {
	type Input struct{}
	type Output struct{}

	h, router := newTestHTML(t)
	Get(h, Operation{Path: "/", Template: template.Must(template.New("page").Parse(`ok`))},
		func(_ context.Context, in *Input) (*Output, error) {
			return nil, errors.New("something went wrong")
		},
	)

	rec := httptest.NewRecorder()
	router.serve(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestRegister_RedirectOnSuccess(t *testing.T) {
	type Input struct{}
	type Output struct{}

	h, router := newTestHTML(t)
	Post(h, Operation{
		Path:        "/",
		Template:    template.Must(template.New("page").Parse(`ok`)),
		RedirectURL: "/success",
	},
		func(_ context.Context, in *Input) (*Output, error) {
			return &Output{}, nil
		},
	)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	router.serve(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/success" {
		t.Fatalf("expected Location: /success, got %q", loc)
	}
}

func TestSetField(t *testing.T) {
	set := func(dest any, s string) {
		setField(reflect.ValueOf(dest).Elem(), s)
	}

	t.Run("string trimmed", func(t *testing.T) {
		var v string
		set(&v, "  hello  ")
		if v != "hello" {
			t.Fatalf("expected \"hello\", got %q", v)
		}
	})

	t.Run("int parsed", func(t *testing.T) {
		var v int
		set(&v, "42")
		if v != 42 {
			t.Fatalf("expected 42, got %d", v)
		}
	})

	t.Run("float64 parsed", func(t *testing.T) {
		var v float64
		set(&v, "3.14")
		if v != 3.14 {
			t.Fatalf("expected 3.14, got %f", v)
		}
	})

	t.Run("bool true values", func(t *testing.T) {
		for _, s := range []string{"true", "on", "1"} {
			var v bool
			set(&v, s)
			if !v {
				t.Fatalf("expected true for %q", s)
			}
		}
	})
	t.Run("bool false for anything else", func(t *testing.T) {
		var v bool
		set(&v, "yes")
		if v {
			t.Fatal("expected false for \"yes\"")
		}
	})

	t.Run("time.Time parsed from date", func(t *testing.T) {
		var v time.Time
		set(&v, "2024-06-01")
		if v.Year() != 2024 || v.Month() != 6 || v.Day() != 1 {
			t.Fatalf("expected 2024-06-01, got %v", v)
		}
	})
}

func TestMergeValidatorErrors(t *testing.T) {
	type noValidator struct{}

	t.Run("non-Validator returned unchanged", func(t *testing.T) {
		in := map[string]string{"field": "err"}
		out := mergeValidatorErrors(&noValidator{}, in)
		if len(out) != 1 || out["field"] != "err" {
			t.Fatalf("expected unchanged map, got %v", out)
		}
	})

	t.Run("Validator errors merged into existing map", func(t *testing.T) {
		v := &validatorStub{errors: map[string]string{"cross": "invalid"}}
		base := map[string]string{"field": "err"}
		out := mergeValidatorErrors(v, base)
		if out["cross"] != "invalid" {
			t.Fatalf("expected cross-field error merged, got %v", out)
		}
		if out["field"] != "err" {
			t.Fatalf("expected original errors preserved, got %v", out)
		}
	})

	t.Run("Validator errors added when fieldErrs is nil", func(t *testing.T) {
		v := &validatorStub{errors: map[string]string{"cross": "invalid"}}
		out := mergeValidatorErrors(v, nil)
		if out["cross"] != "invalid" {
			t.Fatalf("expected cross-field error added to nil map, got %v", out)
		}
	})
}

// validatorStub implements Validator for testing mergeValidatorErrors.
type validatorStub struct{ errors map[string]string }

func (v *validatorStub) Validate() map[string]string { return v.errors }

func TestParseInput(t *testing.T) {
	t.Run("query tag populated", func(t *testing.T) {
		type Input struct {
			Q string `query:"q"`
		}
		req := httptest.NewRequest(http.MethodGet, "/?q=hello", nil)
		in := new(Input)
		if err := parseInput(req, in); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if in.Q != "hello" {
			t.Fatalf("expected \"hello\", got %q", in.Q)
		}
	})

	t.Run("form tag populated from POST body", func(t *testing.T) {
		type Input struct {
			Name string `form:"name"`
		}
		body := url.Values{"name": {"alice"}}.Encode()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		in := new(Input)
		if err := parseInput(req, in); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if in.Name != "alice" {
			t.Fatalf("expected \"alice\", got %q", in.Name)
		}
	})

	t.Run("path tag populated", func(t *testing.T) {
		type Input struct {
			ID string `path:"id"`
		}
		mux := http.NewServeMux()
		var in Input
		mux.HandleFunc("GET /item/{id}", func(w http.ResponseWriter, r *http.Request) {
			if err := parseInput(r, &in); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
		req := httptest.NewRequest(http.MethodGet, "/item/42", nil)
		mux.ServeHTTP(httptest.NewRecorder(), req)
		if in.ID != "42" {
			t.Fatalf("expected \"42\", got %q", in.ID)
		}
	})
}

// failingRenderer always returns an error from Render.
type failingRenderer struct{}

func (f *failingRenderer) Render(w http.ResponseWriter, r *http.Request, status int, tmpl *template.Template, data any) error {
	return errors.New("render failed")
}

func TestRenderError(t *testing.T) {
	errorTmpl := template.Must(template.New("error").Parse(`{{if .}}custom{{else}}Error{{end}}`))

	t.Run("renders with given status code", func(t *testing.T) {
		h := &HTML{renderer: newHTMLRenderer(slog.Default()), errorPage: errorTmpl}
		rec := httptest.NewRecorder()
		h.renderError(rec, httptest.NewRequest(http.MethodGet, "/", nil), http.StatusNotFound, nil)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("falls back to 500 when renderer fails", func(t *testing.T) {
		h := &HTML{renderer: &failingRenderer{}, errorPage: errorTmpl}
		rec := httptest.NewRecorder()
		h.renderError(rec, httptest.NewRequest(http.MethodGet, "/", nil), http.StatusNotFound, nil)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 fallback, got %d", rec.Code)
		}
	})
}

func TestLogPageError(t *testing.T) {
	t.Run("errorAttrs called with PageError", func(t *testing.T) {
		pe := &PageError{Code: http.StatusForbidden}
		called := false
		h := &HTML{
			logger: slog.Default(),
			errorAttrs: func(_ context.Context, got *PageError) []slog.Attr {
				called = true
				if got != pe {
					t.Errorf("expected the same PageError, got %v", got)
				}
				return nil
			},
		}
		h.logPageError(context.Background(), pe)
		if !called {
			t.Fatal("expected errorAttrs to be called")
		}
	})

	t.Run("nil errorAttrs does not panic", func(t *testing.T) {
		h := &HTML{logger: slog.Default(), errorAttrs: nil}
		h.logPageError(context.Background(), &PageError{Code: http.StatusInternalServerError})
	})
}
