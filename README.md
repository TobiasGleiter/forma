<a href="#">
	<picture>
		<img alt="Forma Logo" src="docs/forma.png" />
	</picture>
</a>

<br />

[![Forma](https://img.shields.io/badge/Powered%20By-FORMA-00ff07)](https://github.com/tobiasgleiter/forma) [![Go Report Card](https://goreportcard.com/badge/github.com/tobiasgleiter/forma)](https://goreportcard.com/report/github.com/tobiasgleiter/forma) [![CI](https://github.com/tobiasgleiter/forma/actions/workflows/ci.yml/badge.svg)](https://github.com/tobiasgleiter/forma/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/License-MIT-00ff07.svg)](./LICENSE)

- [What is forma?](#intro)

<a name="intro"></a>

Where [Huma](https://huma.rocks/) defines your JSON API. [Forma](https://forma.rocks/) defines your HTML pages. The same philosophy, the other side of the wire.

Forma is a modern, declarative, type-safe server-side rendering framework for Go. It brings the same `Register`-based handler pattern that Huma uses for REST APIs to HTML template rendering: annotated structs, automatic parameter binding, and pages that stay in sync with your types.

Goals of this project:

- Be the natural companion to Huma for teams that need both an API and a web UI
- Zero boilerplate: Declare a struct, bind your params, render your template
- Guard rails to catch bad input at the boundary, not deep in handler logic
- No magic: Just Go types, struct tags, and `html/template`

Features include:

- `forma.Register` — mirrors Huma's operation registration, but renders an HTML page
  - `path:"name"` binding for URL path parameters
  - `query:"name"` binding for query string parameters
  - `form:"name"` binding for `application/x-www-form-urlencoded` POST fields
- Automatic input validation: Invalid input returns HTTP 422, handler errors render a default error page
- Output struct passed directly as template data — no manual `.Execute` calls
- Router-agnostic works with stdlib `net/http`
- Zero external dependencies beyond the Go standard library

Inspired by [Huma](https://huma.rocks/). Built to sit right beside it.

## Why forma?

Huma solved the hard parts of building JSON REST APIs in Go. But server-rendered HTML endpoints still require the same boilerplate: manual `r.PathValue(...)`, `r.FormValue(...)`, hand-written template execution, custom error pages.

Forma applies the same struct-tag, auto-bind, type-safe philosophy to HTML rendering. If you already think in Huma, forma should feel immediately familiar.

| | Huma | Forma |
|---|---|---|
| Output format | JSON | HTML (via `html/template`) |
| Handler pattern | `func(ctx, *Input) (*Output, error)` | `func(ctx, *Input) (*Output, error)` |
| Parameter binding | path, query, header, cookie | path, query, form |
| Validation | Automatic, tag-driven | Automatic, tag-driven |
| Error response | Structured JSON | HTTP 422 / error page |
| Router | Bring your own | Bring your own |

---

Be sure to star the project if you find it useful!
