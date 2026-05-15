package forma

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tobiasgleiter/forma/validation"
)

// iso4217Codes is the set of active ISO 4217 currency codes.
var iso4217Codes = map[string]struct{}{
	"AED": {}, "AFN": {}, "ALL": {}, "AMD": {}, "ANG": {}, "AOA": {}, "ARS": {}, "AUD": {},
	"AWG": {}, "AZN": {}, "BAM": {}, "BBD": {}, "BDT": {}, "BGN": {}, "BHD": {}, "BIF": {},
	"BMD": {}, "BND": {}, "BOB": {}, "BOV": {}, "BRL": {}, "BSD": {}, "BTN": {}, "BWP": {},
	"BYN": {}, "BZD": {}, "CAD": {}, "CDF": {}, "CHE": {}, "CHF": {}, "CHW": {}, "CLF": {},
	"CLP": {}, "CNY": {}, "COP": {}, "COU": {}, "CRC": {}, "CUC": {}, "CUP": {}, "CVE": {},
	"CZK": {}, "DJF": {}, "DKK": {}, "DOP": {}, "DZD": {}, "EGP": {}, "ERN": {}, "ETB": {},
	"EUR": {}, "FJD": {}, "FKP": {}, "GBP": {}, "GEL": {}, "GHS": {}, "GIP": {}, "GMD": {},
	"GNF": {}, "GTQ": {}, "GYD": {}, "HKD": {}, "HNL": {}, "HTG": {}, "HUF": {}, "IDR": {},
	"ILS": {}, "INR": {}, "IQD": {}, "IRR": {}, "ISK": {}, "JMD": {}, "JOD": {}, "JPY": {},
	"KES": {}, "KGS": {}, "KHR": {}, "KMF": {}, "KPW": {}, "KRW": {}, "KWD": {}, "KYD": {},
	"KZT": {}, "LAK": {}, "LBP": {}, "LKR": {}, "LRD": {}, "LSL": {}, "LYD": {}, "MAD": {},
	"MDL": {}, "MGA": {}, "MKD": {}, "MMK": {}, "MNT": {}, "MOP": {}, "MRU": {}, "MUR": {},
	"MVR": {}, "MWK": {}, "MXN": {}, "MXV": {}, "MYR": {}, "MZN": {}, "NAD": {}, "NGN": {},
	"NIO": {}, "NOK": {}, "NPR": {}, "NZD": {}, "OMR": {}, "PAB": {}, "PEN": {}, "PGK": {},
	"PHP": {}, "PKR": {}, "PLN": {}, "PYG": {}, "QAR": {}, "RON": {}, "RSD": {}, "RUB": {},
	"RWF": {}, "SAR": {}, "SBD": {}, "SCR": {}, "SDG": {}, "SEK": {}, "SGD": {}, "SHP": {},
	"SLE": {}, "SLL": {}, "SOS": {}, "SRD": {}, "SSP": {}, "STN": {}, "SVC": {}, "SYP": {},
	"SZL": {}, "THB": {}, "TJS": {}, "TMT": {}, "TND": {}, "TOP": {}, "TRY": {}, "TTD": {},
	"TWD": {}, "TZS": {}, "UAH": {}, "UGX": {}, "USD": {}, "USN": {}, "UYI": {}, "UYU": {},
	"UYW": {}, "UZS": {}, "VED": {}, "VES": {}, "VND": {}, "VUV": {}, "WST": {}, "XAF": {},
	"XAG": {}, "XAU": {}, "XBA": {}, "XBB": {}, "XBC": {}, "XBD": {}, "XCD": {}, "XDR": {},
	"XOF": {}, "XPD": {}, "XPF": {}, "XPT": {}, "XSU": {}, "XTS": {}, "XUA": {}, "XXX": {},
	"YER": {}, "ZAR": {}, "ZMW": {}, "ZWL": {},
}

var emailRE = regexp.MustCompile(`\A[^@\s]+@[^@\s]+\.[^@\s]+\z`)

// Validator is optionally implemented by input structs for cross-field rules
// that struct tags cannot express. Returned errors are merged with tag-level errors.
type Validator interface {
	Validate() map[string]string
}

// validateInput inspects struct tags on v and returns a field → message error map.
// Returns nil when all fields pass.
func validateInput(v any) map[string]string {
	errors := map[string]string{}
	rv := reflect.ValueOf(v).Elem()
	rt := rv.Type()
	for i := range rt.NumField() {
		f, fv := rt.Field(i), rv.Field(i)

		// Only validate fields that declare an HTTP source tag.
		// The source tag value doubles as the error map key so templates can
		// match errors back to the input element they describe.
		key := sourceKey(f)
		if key == "" {
			continue
		}

		label := f.Name

		switch fv.Kind() {
		case reflect.String:
			validateString(f, fv.String(), key, label, errors)
		case reflect.Int, reflect.Int64:
			validateInt(f, fv.Int(), key, label, errors)
		case reflect.Float64:
			validateFloat(f, fv.Float(), key, label, errors)
		case reflect.Struct:
			if fv.Type() == reflect.TypeFor[time.Time]() {
				validateTime(f, fv.Interface().(time.Time), key, label, errors)
			}
		}
	}
	if len(errors) == 0 {
		return nil
	}
	return errors
}

// msgOrLabel returns the label tag value as a complete error message when set,
// otherwise returns defaultMsg. This lets callers override the generated
// message for fields where the constraint wording doesn't fit the domain.
func msgOrLabel(f reflect.StructField, defaultMsg string) string {
	if l := f.Tag.Get("label"); l != "" {
		return l
	}
	return defaultMsg
}

// sourceKey returns the HTTP input tag (form > query > path) used as the error
// map key. Fields without a source tag are not validated and they have no name
// by which a template could display their error.
func sourceKey(f reflect.StructField) string {
	for _, tag := range []string{"form", "query", "path"} {
		if v := f.Tag.Get(tag); v != "" {
			return v
		}
	}
	return ""
}

func validateString(f reflect.StructField, val, key, label string, errors map[string]string) {
	if _, ok := f.Tag.Lookup("required"); ok && val == "" {
		errors[key] = msgOrLabel(f, fmt.Sprintf(validation.MsgRequired, label))
		return
	}
	// Each check is its own function so this dispatcher stays below the
	// cyclomatic complexity limit; the logic stays close to the constraint it encodes.
	for _, msg := range []string{
		maxLengthMsg(f, val, label),
		minLengthMsg(f, val, label),
		enumMsg(f, val, label),
		emailMsg(f, val, label),
		isoMsg(f, val, label),
		timezoneMsg(f, val, label),
	} {
		if msg != "" {
			errors[key] = msg
			break
		}
	}
}

func maxLengthMsg(f reflect.StructField, val, label string) string {
	maxLength := f.Tag.Get("maxLength")
	if maxLength == "" {
		return ""
	}
	if n, _ := strconv.Atoi(maxLength); len(val) > n {
		return msgOrLabel(f, fmt.Sprintf(validation.MsgMaxLength, label, maxLength))
	}
	return ""
}

func minLengthMsg(f reflect.StructField, val, label string) string {
	minLength := f.Tag.Get("minLength")
	if minLength == "" || val == "" {
		return ""
	}
	if n, _ := strconv.Atoi(minLength); len(val) < n {
		return msgOrLabel(f, fmt.Sprintf(validation.MsgMinLength, label, minLength))
	}
	return ""
}

func enumMsg(f reflect.StructField, val, label string) string {
	enum := f.Tag.Get("enum")
	if enum == "" || val == "" {
		return ""
	}
	for a := range strings.SplitSeq(enum, ",") {
		if val == strings.TrimSpace(a) {
			return ""
		}
	}
	return msgOrLabel(f, fmt.Sprintf(validation.MsgEnum, label, enum))
}

func emailMsg(f reflect.StructField, val, label string) string {
	if f.Tag.Get("email") != "true" || val == "" {
		return ""
	}
	if !emailRE.MatchString(val) {
		return msgOrLabel(f, fmt.Sprintf(validation.MsgEmail, label))
	}
	return ""
}

func isoMsg(f reflect.StructField, val, label string) string {
	standard := f.Tag.Get("iso")
	if standard == "" || val == "" {
		return ""
	}
	if standard == "4217" {
		if _, ok := iso4217Codes[strings.ToUpper(val)]; !ok {
			return msgOrLabel(f, fmt.Sprintf(validation.MsgISO4217, label))
		}
	}
	return ""
}

func timezoneMsg(f reflect.StructField, val, label string) string {
	if f.Tag.Get("timezone") != "iana" || val == "" {
		return ""
	}
	if _, err := time.LoadLocation(val); err != nil {
		return msgOrLabel(f, fmt.Sprintf(validation.MsgIANATimezone, label))
	}
	return ""
}

func validateInt(f reflect.StructField, n int64, key, label string, errors map[string]string) {
	if minStr := f.Tag.Get("min"); minStr != "" {
		if v, _ := strconv.ParseInt(minStr, 10, 64); n < v {
			errors[key] = msgOrLabel(f, fmt.Sprintf(validation.MsgMin, label, minStr))
		}
	}
	if maxStr := f.Tag.Get("max"); maxStr != "" {
		if v, _ := strconv.ParseInt(maxStr, 10, 64); n > v {
			errors[key] = msgOrLabel(f, fmt.Sprintf(validation.MsgMax, label, maxStr))
		}
	}
	if mul := f.Tag.Get("multipleOf"); mul != "" {
		if m, _ := strconv.ParseInt(mul, 10, 64); m != 0 && n%m != 0 {
			errors[key] = msgOrLabel(f, fmt.Sprintf(validation.MsgMultipleOf, label, mul))
		}
	}
}

func validateFloat(f reflect.StructField, n float64, key, label string, errors map[string]string) {
	if minStr := f.Tag.Get("min"); minStr != "" {
		if v, _ := strconv.ParseFloat(minStr, 64); n < v {
			errors[key] = msgOrLabel(f, fmt.Sprintf(validation.MsgMin, label, minStr))
		}
	}
	if maxStr := f.Tag.Get("max"); maxStr != "" {
		if v, _ := strconv.ParseFloat(maxStr, 64); n > v {
			errors[key] = msgOrLabel(f, fmt.Sprintf(validation.MsgMax, label, maxStr))
		}
	}
}

func validateTime(f reflect.StructField, t time.Time, key, label string, errors map[string]string) {
	if _, ok := f.Tag.Lookup("required"); ok && t.IsZero() {
		errors[key] = msgOrLabel(f, fmt.Sprintf(validation.MsgRequired, label))
		return
	}
	// A zero time with no required tag is treated as "not provided" therefore skip range checks.
	if t.IsZero() {
		return
	}
	if minLength := f.Tag.Get("min"); minLength != "" {
		if minT, err := time.Parse("2006-01-02", minLength); err == nil && t.Before(minT) {
			errors[key] = msgOrLabel(f, fmt.Sprintf(validation.MsgMinDate, label, minLength))
		}
	}
	if maxLength := f.Tag.Get("max"); maxLength != "" {
		if maxT, err := time.Parse("2006-01-02", maxLength); err == nil && t.After(maxT) {
			errors[key] = msgOrLabel(f, fmt.Sprintf(validation.MsgMaxDate, label, maxLength))
		}
	}
}
