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
		if maxLengthMsg(sf(`maxLength:"3"`), "abcd", "F") == "" {
			t.Fatal("expected error for value over maxLength")
		}
	})
	t.Run("at limit returns empty", func(t *testing.T) {
		if maxLengthMsg(sf(`maxLength:"3"`), "abc", "F") != "" {
			t.Fatal("expected no error at maxLength limit")
		}
	})
}

func TestMinLengthMsg(t *testing.T) {
	t.Run("under limit returns error", func(t *testing.T) {
		if minLengthMsg(sf(`minLength:"5"`), "ab", "F") == "" {
			t.Fatal("expected error for value under minLength")
		}
	})
	t.Run("at limit returns empty", func(t *testing.T) {
		if minLengthMsg(sf(`minLength:"5"`), "abcde", "F") != "" {
			t.Fatal("expected no error at minLength limit")
		}
	})
	t.Run("empty value skipped", func(t *testing.T) {
		// Empty non-required fields must not trigger minLength.
		if minLengthMsg(sf(`minLength:"5"`), "", "F") != "" {
			t.Fatal("expected minLength to be skipped for empty value")
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
