package validation

// List of user-facing form validation error messages.
// Each format string takes the field label as the first %s argument.
const (
	MsgRequired     = "%s is required"
	MsgMaxLength    = "%s must not exceed %s characters"
	MsgMinLength    = "%s must be at least %s characters"
	MsgEnum         = "%s must be one of: %s"
	MsgEmail        = "%s must be a valid email address"
	MsgISO4217      = "%s must be a valid ISO 4217 currency code"
	MsgIANATimezone = "%s must be a valid IANA timezone"
	MsgMin          = "%s must be at least %s"
	MsgMax          = "%s must be at most %s"
	MsgMultipleOf   = "%s must be a multiple of %s"
	MsgMinDate      = "%s must be on or after %s"
	MsgMaxDate      = "%s must be on or before %s"
)
