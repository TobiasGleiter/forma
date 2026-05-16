---
description: Use built-in validators and the Validator interface to validate input data with whatever rules you need.
---

# Custom Validation

## Built-in Validators

Forma validates struct fields automatically using struct tags. Validation runs on POST, PUT, and PATCH requests before the handler is called. Fields without a source tag (`form`, `query`, or `path`) are silently skipped.

### String validators

| Tag | Example | Description |
|---|---|---|
| `required` | `required:"true"` | Fails if the value is empty |
| `min` | `min:"2"` | Minimum character count |
| `max` | `max:"20"` | Maximum character count |
| `enum` | `enum:"foo,bar,baz"` | Value must match one of the comma-separated options |
| `email` | `email:"true"` | Must be a valid email address |
| `iso` | `iso:"4217"` | Must be a valid ISO 4217 currency code |
| `timezone` | `timezone:"iana"` | Must be a valid IANA timezone name |

### Numeric validators (int, int64, float64)

| Tag | Example | Description |
|---|---|---|
| `min` | `min:"1"` | Minimum value (inclusive) |
| `max` | `max:"100"` | Maximum value (inclusive) |
| `multipleOf` | `multipleOf:"5"` | Value must be a multiple of n |

### Date validators (time.Time)

| Tag | Example | Description |
|---|---|---|
| `required` | `required:"true"` | Fails if the value is zero |
| `min` | `min:"2020-01-01"` | Must be on or after this date (`YYYY-MM-DD`) |
| `max` | `max:"2099-12-31"` | Must be on or before this date (`YYYY-MM-DD`) |

### Overriding the error message

Add a `label` tag to replace the generated error message entirely:

```go title="code.go"
type BookingInput struct {
    StartDate time.Time `form:"start_date" required:"true" label:"Pick a start date"`
}
```

### Example

```go title="code.go"
type RegisterInput struct {
    Username string    `form:"username" required:"true" min:"3" max:"20"`
    Email    string    `form:"email"    required:"true" email:"true"`
    Age      int       `form:"age"      min:"18" max:"120"`
    Currency string    `form:"currency" iso:"4217"`
    Plan     string    `form:"plan"     enum:"free,pro,enterprise"`
}
```

## Custom Validators

For cross-field rules that tags cannot express, implement the `Validator` interface on the input struct. `Validate` is called after tag validation passes, so you can assume basic constraints already hold.

```go title="code.go"
type Validator interface {
    Validate() map[string]string
}
```

The returned map keys must match the source tag values used in the template to display errors (`form`, `query`, or `path` tag values).

### Example

```go title="code.go"
type DateRangeInput struct {
    StartDate time.Time `form:"start_date" required:"true"`
    EndDate   time.Time `form:"end_date"   required:"true"`
}

func (i *DateRangeInput) Validate() map[string]string {
    if !i.StartDate.IsZero() && !i.EndDate.IsZero() && !i.EndDate.After(i.StartDate) {
        return map[string]string{
            "end_date": "End date must be after start date",
        }
    }
    return nil
}
```

The template accesses these errors the same way as tag errors:

```html title="template.html"
<input type="date" name="end_date">
{{- if .Errors.end_date }}
<p>{{ .Errors.end_date }}</p>
{{- end }}
```
