// Package forma provides a typed HTML handler registration framework.
// Handlers declare input/output as plain structs; the framework handles
// parsing, validation, rendering, and redirects.
//
// Design: encoding routing, parsing, validation, and rendering in one
// generic Register call means each route is a plain function
// (func(context.Context, *Input) (*Output, error)) with no HTTP awareness.
// Handlers are easy to test and reason about independently of the transport.
package forma

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Router is the minimal interface for registering HTTP handlers.
type Router interface {
	Handle(method, path string, h http.HandlerFunc)
}

// Renderer renders HTTP responses. A single method handles all response types:
// success (2xx), validation errors (422), and application errors (4xx/5xx).
// The status code is the discriminator; the Renderer decides how to render each.
//
// One method instead of separate Render/Error methods keeps the interface
// minimal. It also lets the Renderer pull request-scoped data (trace IDs, auth
// context) from r for error pages, without the framework needing to know
// about those conventions, the framework passes nil data for its own errors
// and the Renderer fills in whatever it needs from the request context.
//
// If Render returns an error the framework falls back to http.Error. Calling
// Render again would risk an infinite loop, so the caller must not retry.
type Renderer interface {
	Render(w http.ResponseWriter, r *http.Request, status int, tmpl *template.Template, data any) error
}

// PageError is returned by handler functions to render a specific HTTP status
// code and custom template data on the error page. It implements the error
// interface so handlers keep the standard func(context.Context, *I) (*O, error)
// signature.
//
// Data is passed directly to the error page template so it can carry any
// fields the template needs: Message, TraceID, support links, etc.
//
// Example:
//
//	return nil, &forma.PageError{
//	    Code: http.StatusNotFound,
//	    Data: MyErrorData{Message: "Data not found."},
//	}
type PageError struct {
	Code int
	Data any
}

// Error satisfies the error interface. The message is the standard HTTP status
// text for the code, which is always human-readable without extra fields.
func (e *PageError) Error() string {
	return http.StatusText(e.Code)
}

// PageData is the envelope passed to all form templates.
//
// URL is the request URL, available for building links or reading query params.
//
// Input holds parsed form/query/path values and is always populated, even on
// validation re-renders, so templates can re-fill submitted fields.
//
// Output holds the handler's domain result. It is nil on validation re-renders
// because fn is skipped when validation fails (see [Register]).
//
// Errors maps source-tag field names to validation messages; nil on success.
type PageData[I, O any] struct {
	URL    *url.URL
	Input  *I
	Output *O
	Errors map[string]string
}

// HTML wraps a Router and holds shared rendering state.
type HTML struct {
	router     Router
	renderer   Renderer
	errorPage  *template.Template
	logger     *slog.Logger
	errorAttrs func(ctx context.Context, pe *PageError) []slog.Attr
}

// New returns an HTML router backed by router and the renderer in cfg.
//
// cfg.ErrorTemplate is rendered for all framework-level errors: parse
// failures, handler errors, and PageError returns. A single error page is
// shared across all routes because error presentation is an application-wide
// concern. Pass a nil ErrorTemplate to use the built-in minimal fallback.
//
// When the framework triggers a non-PageError, it passes nil data to Render.
// The View is responsible for building the error page data from the request
// context (trace IDs, auth session, etc.).
func New(router Router, cfg Config) *HTML {
	var errorPage = template.Must(template.New("base").Parse(`Error`))
	if cfg.ErrorTemplate != nil {
		errorPage = cfg.ErrorTemplate
	}
	var logger = slog.Default()
	if cfg.Logger != nil {
		logger = cfg.Logger
	}
	var renderer Renderer = newHTMLRenderer(logger)
	if cfg.Renderer != nil {
		renderer = cfg.Renderer
	}
	h := &HTML{
		router:     router,
		renderer:   renderer,
		errorPage:  errorPage,
		logger:     logger,
		errorAttrs: cfg.ErrorAttrs,
	}
	return h
}

// renderError is the framework's internal error path. data is nil for
// framework-triggered errors (View builds page data from context) and non-nil
// when the handler returned a PageError with custom template data.
//
// If Render itself returns an error we fall back to http.Error. Calling Render
// again would risk an infinite render loop, so the fallback is intentionally
// dependency-free.
func (m *HTML) renderError(w http.ResponseWriter, r *http.Request, code int, data any) {
	if err := m.renderer.Render(w, r, code, m.errorPage, data); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (m *HTML) logPageError(ctx context.Context, pe *PageError) {
	var args []any
	if m.errorAttrs != nil {
		for _, a := range m.errorAttrs(ctx, pe) {
			args = append(args, a)
		}
	}
	m.logger.ErrorContext(ctx, "handler error", args...)
}

// Register an operation handler for an HTML router. The handler must be a
// function that takes a context and a pointer to the input struct, and returns
// a pointer to the output struct and an error. The input struct fields are
// populated from path, query, and form parameters via struct tags. The output
// struct is passed as template data when rendering the response page.
//
// Example:
//
//	forma.Register(html, forma.Operation[GreetingOutput]{
//		Method:   http.MethodGet,
//		Path:     "/greeting/{name}",
//		Template: greetingTmpl,
//	}, func(ctx context.Context, input *GreetingInput) (*GreetingOutput, error) {
//		return &GreetingOutput{Message: "Hello, " + input.Name + "!"}, nil
//	})
func Register[I, O any](m *HTML, op Operation[O], fn func(context.Context, *I) (*O, error)) {
	m.router.Handle(op.Method, op.Path, func(w http.ResponseWriter, r *http.Request) {
		in := new(I)
		if err := parseInput(r, in); err != nil {
			m.renderError(w, r, http.StatusBadRequest, nil)
			return
		}

		isPost := r.Method == http.MethodPost
		isPut := r.Method == http.MethodPut
		isPatch := r.Method == http.MethodPatch

		var fieldErrs map[string]string
		if isPost || isPut || isPatch {
			fieldErrs = mergeValidatorErrors(in, validateInput(in))
		}
		if len(fieldErrs) > 0 {
			td := &PageData[I, O]{URL: r.URL, Input: in, Errors: fieldErrs}
			if err := m.renderer.Render(w, r, op.validationCode(), op.entrypoint(), td); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		out, err := fn(r.Context(), in)
		if err != nil {
			if pe, ok := errors.AsType[*PageError](err); ok {
				m.logPageError(r.Context(), pe)
				m.renderError(w, r, pe.Code, pe.Data)
			} else {
				m.logger.ErrorContext(r.Context(), "handler error", slog.Int("status", http.StatusInternalServerError), slog.String("error", err.Error()))
				m.renderError(w, r, http.StatusInternalServerError, nil)
			}
			return
		}

		if redirectTo := op.redirectURL(out); redirectTo != "" {
			http.Redirect(w, r, redirectTo, http.StatusSeeOther)
			return
		}

		td := &PageData[I, O]{URL: r.URL, Input: in, Output: out}
		if err := m.renderer.Render(w, r, op.successCode(), op.entrypoint(), td); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	})
}

// parseInput populates v from r using path/query/form struct tags.
func parseInput(r *http.Request, v any) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form: %w", err)
	}
	rv := reflect.ValueOf(v).Elem()
	rt := rv.Type()
	for i := range rt.NumField() {
		f, fv := rt.Field(i), rv.Field(i)
		if tag, ok := f.Tag.Lookup("path"); ok {
			setField(fv, r.PathValue(tag))
		}
		if tag, ok := f.Tag.Lookup("query"); ok {
			setField(fv, r.URL.Query().Get(tag))
		}
		if tag, ok := f.Tag.Lookup("form"); ok {
			setField(fv, r.PostForm.Get(tag))
		}
	}
	return nil
}

// mergeValidatorErrors calls v.Validate() if implemented and folds its results
// into errors. Validator runs after tag validation so it can assume basic
// constraints already passed.
func mergeValidatorErrors(v any, fieldErrs map[string]string) map[string]string {
	impl, ok := v.(Validator)
	if !ok {
		return fieldErrs
	}
	for k, msg := range impl.Validate() {
		if fieldErrs == nil {
			fieldErrs = make(map[string]string)
		}
		fieldErrs[k] = msg
	}
	return fieldErrs
}

// setField converts s to the kind of fv and sets it.
// time.Time fields are parsed from the HTML date input format "2006-01-02".
func setField(fv reflect.Value, s string) {
	switch fv.Kind() {
	case reflect.String:
		fv.SetString(strings.TrimSpace(s))
	case reflect.Int, reflect.Int64:
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			fv.SetInt(n)
		}
	case reflect.Float64:
		if n, err := strconv.ParseFloat(s, 64); err == nil {
			fv.SetFloat(n)
		}
	case reflect.Bool:
		isTrue := s == "true"
		isOn := s == "on"
		is1 := s == "1"

		fv.SetBool(isTrue || isOn || is1)
	case reflect.Struct:
		setStructField(fv, s)
	}
}

func setStructField(fv reflect.Value, s string) {
	if fv.Type() == reflect.TypeFor[time.Time]() {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			fv.Set(reflect.ValueOf(t))
		}
	}
}
