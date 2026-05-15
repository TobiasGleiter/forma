package forma

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
)

// Config holds the rendering dependencies for an HTML router.
//
// Use DefaultConfig to get the built-in renderer pre-wired. To swap in a
// custom renderer, set Renderer after calling DefaultConfig:
//
//	cfg := forma.DefaultConfig(logger)
//	cfg.Renderer = myRenderer
type Config struct {
	// Renderer renders all responses. If nil, the built-in HTML renderer is used.
	Renderer Renderer
	// ErrorTemplate is rendered for all framework-level errors. Defaults to
	// the built-in minimal fallback when nil.
	ErrorTemplate *template.Template
	// Logger is used to log handler errors. If nil, slog.Default() is used.
	Logger *slog.Logger
	// ErrorAttrs, when non-nil, is called for every PageError to produce
	// additional log attributes. Use it to extract fields from pe.Data such as
	// a TraceID or user-facing message. By default only status and error are
	// appended.
	ErrorAttrs func(ctx context.Context, pe *PageError) []slog.Attr
}

// DefaultConfig returns a Config with the built-in HTML renderer pre-wired.
// It uses slog.Default() as the logger, so it inherits any logger set via slog.SetDefault.
// Set Config fields after calling this to override defaults:
//
//	cfg := forma.DefaultConfig()
//	cfg.ErrorTemplate = cache[pages.Error.File]
func DefaultConfig() Config {
	logger := slog.Default()
	return Config{
		Renderer: newHTMLRenderer(logger),
		Logger:   logger,
	}
}

type htmlRenderer struct {
	logger *slog.Logger
}

func newHTMLRenderer(logger *slog.Logger) *htmlRenderer {
	return &htmlRenderer{logger: logger}
}

// Render executes tmpl and writes the result to w with the given status code.
// When data is nil the renderer builds a minimal struct with Message so error
// templates always have something to render.
func (re *htmlRenderer) Render(w http.ResponseWriter, r *http.Request, status int, tmpl *template.Template, data any) error {
	if data == nil {
		data = struct {
			Message string
		}{
			Message: http.StatusText(status),
		}
	}
	buf := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(buf, tmpl.Name(), data); err != nil {
		re.logger.ErrorContext(r.Context(), "respond: execute template", slog.String("error", err.Error()))
		return fmt.Errorf("execute template: %w", err)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if _, err := buf.WriteTo(w); err != nil {
		return fmt.Errorf("write response: %w", err)
	}
	return nil
}
