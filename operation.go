package forma

import (
	"html/template"
	"net/http"
)

// Operation holds per-route configuration for Get and Post handlers.
//
// SuccessCode and ValidationCode are per-operation rather than global because
// different routes have different HTTP semantics: a resource-creating POST
// should return 201, a search form is fine with 200. Zero means use the
// framework default (200 and 422 respectively).
//
// RedirectURL configures a static POST-redirect-GET target. It is unused by
// Get handlers.
type Operation struct {
	Path     string
	Template *template.Template

	// TemplateName is the named template to execute within Template.
	// Defaults to Template.Name() when empty. Set this when the entry point
	// is a layout defined as {{ define "layout" }} inside a multi-file template
	// set where Template.Name() is a throwaway root name.
	TemplateName string

	// SuccessCode overrides the default 200 for successful renders.
	// Common override: http.StatusCreated (201) for resource-creating POSTs.
	SuccessCode int

	// ValidationCode overrides the default 422 for validation-error re-renders.
	ValidationCode int

	// RedirectURL is a static POST-redirect-GET target. Unused by Get.
	RedirectURL string
}

// Operationf holds per-route configuration for Postf handlers, which derive
// the redirect URL from the handler output.
//
// Redirect takes priority over the embedded Operation.RedirectURL when both
// are set. Use RedirectURL for static targets (e.g. a list page); use Redirect
// when the URL depends on the output (e.g. a newly created resource's detail page).
type Operationf[O any] struct {
	Operation

	// Redirect derives the redirect URL from the handler output. Takes
	// priority over Operation.RedirectURL. No redirect occurs when it returns "".
	Redirect func(*O) string
}

func (op Operationf[O]) redirectURL(out *O) string {
	if op.Redirect != nil {
		return op.Redirect(out)
	}
	return op.Operation.RedirectURL
}

// entrypoint returns the template to execute. When TemplateName is set it
// resolves the named sub-template via Lookup so tmpl.Name() inside the
// renderer returns the correct name. Falls back to Template when the name
// isn't found or TemplateName is empty.
func (op Operation) entrypoint() *template.Template {
	if op.TemplateName == "" {
		return op.Template
	}
	if t := op.Template.Lookup(op.TemplateName); t != nil {
		return t
	}
	return op.Template
}

func (op Operation) successCode() int {
	if op.SuccessCode != 0 {
		return op.SuccessCode
	}
	return http.StatusOK
}

func (op Operation) validationCode() int {
	if op.ValidationCode != 0 {
		return op.ValidationCode
	}
	return http.StatusUnprocessableEntity
}
