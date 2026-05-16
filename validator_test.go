package forma

import (
	"reflect"
	"testing"
	"time"
)

// sf builds a reflect.StructField with the given tag for testing.
func sf(tag string) reflect.StructField {
	type S struct{ F string }
	f := reflect.TypeFor[S]().Field(0)
	f.Tag = reflect.StructTag(tag)
	return f
}

func TestMaxLengthMsg(t *testing.T) {
	t.Run("over limit returns error", func(t *testing.T) {
		if maxLengthMsg(sf(`max:"3"`), "abcd", "F") == "" {
			t.Fatal("expected error for value over max")
		}
	})
	t.Run("at limit returns empty", func(t *testing.T) {
		if maxLengthMsg(sf(`max:"3"`), "abc", "F") != "" {
			t.Fatal("expected no error at max limit")
		}
	})
}

func TestMinLengthMsg(t *testing.T) {
	t.Run("under limit returns error", func(t *testing.T) {
		if minLengthMsg(sf(`min:"5"`), "ab", "F") == "" {
			t.Fatal("expected error for value under min")
		}
	})
	t.Run("at limit returns empty", func(t *testing.T) {
		if minLengthMsg(sf(`min:"5"`), "abcde", "F") != "" {
			t.Fatal("expected no error at min limit")
		}
	})
	t.Run("empty value skipped", func(t *testing.T) {
		// Empty non-required fields must not trigger min.
		if minLengthMsg(sf(`min:"5"`), "", "F") != "" {
			t.Fatal("expected min to be skipped for empty value")
		}
	})
}

func TestEnumMsg(t *testing.T) {
	t.Run("value in enum returns empty", func(t *testing.T) {
		if enumMsg(sf(`enum:"a,b,c"`), "b", "F") != "" {
			t.Fatal("expected no error for value in enum")
		}
	})
	t.Run("value not in enum returns error", func(t *testing.T) {
		if enumMsg(sf(`enum:"a,b,c"`), "d", "F") == "" {
			t.Fatal("expected error for value not in enum")
		}
	})
	t.Run("whitespace around enum values trimmed", func(t *testing.T) {
		if enumMsg(sf(`enum:"a, b, c"`), "b", "F") != "" {
			t.Fatal("expected whitespace around enum values to be trimmed")
		}
	})
	t.Run("case sensitive", func(t *testing.T) {
		if enumMsg(sf(`enum:"a,b"`), "A", "F") == "" {
			t.Fatal("expected enum match to be case-sensitive")
		}
	})
}

func TestEmailMsg(t *testing.T) {
	t.Run("valid address returns empty", func(t *testing.T) {
		if emailMsg(sf(`email:"true"`), "user@example.com", "F") != "" {
			t.Fatal("expected no error for valid email")
		}
	})
	t.Run("missing @ returns error", func(t *testing.T) {
		if emailMsg(sf(`email:"true"`), "notanemail", "F") == "" {
			t.Fatal("expected error for address without @")
		}
	})
	t.Run("missing domain returns error", func(t *testing.T) {
		if emailMsg(sf(`email:"true"`), "user@", "F") == "" {
			t.Fatal("expected error for address without domain")
		}
	})
}

func TestIsoMsg(t *testing.T) {
	t.Run("valid ISO 4217 code returns empty", func(t *testing.T) {
		if isoMsg(sf(`iso:"4217"`), "EUR", "F") != "" {
			t.Fatal("expected no error for valid ISO 4217 code")
		}
	})
	t.Run("lowercase code accepted", func(t *testing.T) {
		if isoMsg(sf(`iso:"4217"`), "eur", "F") != "" {
			t.Fatal("expected lowercase ISO 4217 code to be accepted")
		}
	})
	t.Run("unknown code returns error", func(t *testing.T) {
		if isoMsg(sf(`iso:"4217"`), "ZZZ", "F") == "" {
			t.Fatal("expected error for unknown ISO 4217 code")
		}
	})
}

func TestTimezoneMsg(t *testing.T) {
	t.Run("valid IANA timezone returns empty", func(t *testing.T) {
		if timezoneMsg(sf(`timezone:"iana"`), "Europe/Berlin", "F") != "" {
			t.Fatal("expected no error for valid IANA timezone")
		}
	})
	t.Run("invalid timezone returns error", func(t *testing.T) {
		if timezoneMsg(sf(`timezone:"iana"`), "Mars/Olympus", "F") == "" {
			t.Fatal("expected error for invalid IANA timezone")
		}
	})
}

func TestValidateInt(t *testing.T) {
	run := func(tag string, val int64) map[string]string {
		type S struct{ N int }
		f := reflect.TypeFor[S]().Field(0)
		f.Tag = reflect.StructTag(tag)
		errs := map[string]string{}
		validateInt(f, val, "n", "N", errs)
		return errs
	}

	t.Run("below min returns error", func(t *testing.T) {
		if len(run(`min:"10"`, 5)) == 0 {
			t.Fatal("expected error for value below min")
		}
	})
	t.Run("at min returns no error", func(t *testing.T) {
		if len(run(`min:"10"`, 10)) != 0 {
			t.Fatal("expected no error at min")
		}
	})
	t.Run("above max returns error", func(t *testing.T) {
		if len(run(`max:"10"`, 11)) == 0 {
			t.Fatal("expected error for value above max")
		}
	})
	t.Run("at max returns no error", func(t *testing.T) {
		if len(run(`max:"10"`, 10)) != 0 {
			t.Fatal("expected no error at max")
		}
	})
	t.Run("not multiple of returns error", func(t *testing.T) {
		if len(run(`multipleOf:"3"`, 7)) == 0 {
			t.Fatal("expected error for value not a multiple")
		}
	})
	t.Run("is multiple of returns no error", func(t *testing.T) {
		if len(run(`multipleOf:"3"`, 9)) != 0 {
			t.Fatal("expected no error for valid multiple")
		}
	})
}

func TestValidateFloat(t *testing.T) {
	run := func(tag string, val float64) map[string]string {
		type S struct{ N float64 }
		f := reflect.TypeFor[S]().Field(0)
		f.Tag = reflect.StructTag(tag)
		errs := map[string]string{}
		validateFloat(f, val, "n", "N", errs)
		return errs
	}

	t.Run("below min returns error", func(t *testing.T) {
		if len(run(`min:"1.5"`, 1.0)) == 0 {
			t.Fatal("expected error for value below min")
		}
	})
	t.Run("at min returns no error", func(t *testing.T) {
		if len(run(`min:"1.5"`, 1.5)) != 0 {
			t.Fatal("expected no error at min")
		}
	})
	t.Run("above max returns error", func(t *testing.T) {
		if len(run(`max:"1.5"`, 2.0)) == 0 {
			t.Fatal("expected error for value above max")
		}
	})
}

func TestSourceKey(t *testing.T) {
	field := func(tag string) reflect.StructField {
		type S struct{ F string }
		f := reflect.TypeFor[S]().Field(0)
		f.Tag = reflect.StructTag(tag)
		return f
	}

	t.Run("form tag returned", func(t *testing.T) {
		if got := sourceKey(field(`form:"name"`)); got != "name" {
			t.Fatalf("expected \"name\", got %q", got)
		}
	})
	t.Run("query tag returned", func(t *testing.T) {
		if got := sourceKey(field(`query:"q"`)); got != "q" {
			t.Fatalf("expected \"q\", got %q", got)
		}
	})
	t.Run("path tag returned", func(t *testing.T) {
		if got := sourceKey(field(`path:"id"`)); got != "id" {
			t.Fatalf("expected \"id\", got %q", got)
		}
	})
	t.Run("form takes priority over query", func(t *testing.T) {
		if got := sourceKey(field(`form:"a" query:"b"`)); got != "a" {
			t.Fatalf("expected form tag \"a\" to win, got %q", got)
		}
	})
	t.Run("no source tag returns empty", func(t *testing.T) {
		if got := sourceKey(field(`json:"name"`)); got != "" {
			t.Fatalf("expected empty string for non-source tag, got %q", got)
		}
	})
}

func TestValidateString(t *testing.T) {
	run := func(tag, val string) map[string]string {
		type S struct{ Name string }
		f := reflect.TypeFor[S]().Field(0)
		f.Tag = reflect.StructTag(tag)
		errs := map[string]string{}
		validateString(f, val, "name", "Name", errs)
		return errs
	}

	t.Run("required empty returns error", func(t *testing.T) {
		if len(run(`form:"name" required:""`, "")) == 0 {
			t.Fatal("expected error for empty required string")
		}
	})
	t.Run("required stops further checks", func(t *testing.T) {
		// Empty required field: no additional errors beyond required itself.
		errs := run(`form:"name" required:"" min:"5"`, "")
		if len(errs) != 1 {
			t.Fatalf("expected exactly 1 error, got %d: %v", len(errs), errs)
		}
	})
	t.Run("empty non-required skips all checks", func(t *testing.T) {
		if len(run(`form:"name" min:"5" max:"2"`, "")) != 0 {
			t.Fatal("expected all checks skipped for empty non-required string")
		}
	})
	t.Run("first failing check wins", func(t *testing.T) {
		// Value violates min first, then enum — only one error recorded.
		errs := run(`form:"name" min:"10" enum:"a,b"`, "c")
		if len(errs) != 1 {
			t.Fatalf("expected exactly 1 error, got %d: %v", len(errs), errs)
		}
	})
	t.Run("label tag overrides message", func(t *testing.T) {
		errs := run(`form:"name" required:"" label:"Enter your name"`, "")
		if errs["name"] != "Enter your name" {
			t.Fatalf("expected label override, got %q", errs["name"])
		}
	})
}

func TestValidateInput(t *testing.T) {
	t.Run("no source tags returns nil", func(t *testing.T) {
		type Input struct {
			Internal string
		}
		if validateInput(&Input{Internal: ""}) != nil {
			t.Fatal("expected nil for struct with no source tags")
		}
	})
	t.Run("all fields valid returns nil", func(t *testing.T) {
		type Input struct {
			Name string `form:"name" required:""`
		}
		if validateInput(&Input{Name: "alice"}) != nil {
			t.Fatal("expected nil when all fields pass validation")
		}
	})
	t.Run("invalid field returned in map", func(t *testing.T) {
		type Input struct {
			Name string `form:"name" required:""`
		}
		errs := validateInput(&Input{Name: ""})
		if errs == nil || errs["name"] == "" {
			t.Fatal("expected error for missing required field")
		}
	})
	t.Run("only failing fields appear in map", func(t *testing.T) {
		type Input struct {
			Name  string `form:"name" required:""`
			Email string `form:"email" required:""`
		}
		errs := validateInput(&Input{Name: "alice", Email: ""})
		if len(errs) != 1 {
			t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
		}
		if errs["email"] == "" {
			t.Fatal("expected error keyed by \"email\"")
		}
	})
	t.Run("fields without source tags skipped", func(t *testing.T) {
		type Input struct {
			Name     string `form:"name" required:""`
			Internal string `required:""`
		}
		errs := validateInput(&Input{Name: "alice", Internal: ""})
		if errs != nil {
			t.Fatalf("expected Internal field without source tag to be skipped, got %v", errs)
		}
	})
	t.Run("int field dispatched", func(t *testing.T) {
		type Input struct {
			Age int `form:"age" min:"18"`
		}
		if validateInput(&Input{Age: 10}) == nil {
			t.Fatal("expected error for int below min")
		}
	})
	t.Run("float64 field dispatched", func(t *testing.T) {
		type Input struct {
			Score float64 `form:"score" max:"1.0"`
		}
		if validateInput(&Input{Score: 1.5}) == nil {
			t.Fatal("expected error for float64 above max")
		}
	})
	t.Run("time.Time field dispatched", func(t *testing.T) {
		type Input struct {
			Date time.Time `form:"date" required:""`
		}
		if validateInput(&Input{Date: time.Time{}}) == nil {
			t.Fatal("expected error for zero required time.Time")
		}
	})
}

func TestValidateTime(t *testing.T) {
	run := func(tag string, val time.Time) map[string]string {
		type S struct{ D time.Time }
		f := reflect.TypeFor[S]().Field(0)
		f.Tag = reflect.StructTag(tag)
		errs := map[string]string{}
		validateTime(f, val, "d", "D", errs)
		return errs
	}

	t.Run("zero time with required returns error", func(t *testing.T) {
		if len(run(`form:"d" required:""`, time.Time{})) == 0 {
			t.Fatal("expected error for zero time with required tag")
		}
	})
	t.Run("zero time without required skips range checks", func(t *testing.T) {
		if len(run(`form:"d" min:"2020-01-01"`, time.Time{})) != 0 {
			t.Fatal("expected zero time without required to skip range checks")
		}
	})
	t.Run("before min returns error", func(t *testing.T) {
		d := time.Date(2019, 12, 31, 0, 0, 0, 0, time.UTC)
		if len(run(`form:"d" min:"2020-01-01"`, d)) == 0 {
			t.Fatal("expected error for date before min")
		}
	})
	t.Run("at min returns no error", func(t *testing.T) {
		d := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		if len(run(`form:"d" min:"2020-01-01"`, d)) != 0 {
			t.Fatal("expected no error for date at min")
		}
	})
	t.Run("after max returns error", func(t *testing.T) {
		d := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
		if len(run(`form:"d" max:"2025-01-01"`, d)) == 0 {
			t.Fatal("expected error for date after max")
		}
	})
	t.Run("malformed date tag silently ignored", func(t *testing.T) {
		d := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		if len(run(`form:"d" min:"not-a-date"`, d)) != 0 {
			t.Fatal("expected malformed date tag to be silently ignored")
		}
	})
}
