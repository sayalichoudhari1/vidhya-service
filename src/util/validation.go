package util

import (
	"net/mail"
	"strings"
)

// IsStrEmpty reports whether s is empty once whitespace is trimmed.
func IsStrEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsValidEmail reports whether s is a syntactically valid email address.
func IsValidEmail(s string) bool {
	if IsStrEmpty(s) {
		return false
	}
	_, err := mail.ParseAddress(s)
	return err == nil
}

// IsValidPhone reports whether s looks like a plausible phone number: only
// digits plus the common separators (+, -, spaces, parentheses), with
// between 7 and 15 digits overall (E.164 max length). Phone is one of the
// fields used to detect duplicate person records (see NormalizePhone), so
// it must be present and well-formed whenever a person entity is created.
func IsValidPhone(s string) bool {
	if IsStrEmpty(s) {
		return false
	}
	digits := 0
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
			digits++
		case r == '+' || r == '-' || r == ' ' || r == '(' || r == ')':
			// allowed separators
		default:
			return false
		}
	}
	return digits >= 7 && digits <= 15
}

// NormalizePhone strips everything except digits and a leading '+', so
// "+1 (555) 123-4567" and "15551234567" compare as the same number for
// duplicate-detection purposes.
func NormalizePhone(s string) string {
	out := make([]rune, 0, len(s))
	for i, r := range s {
		switch {
		case r >= '0' && r <= '9':
			out = append(out, r)
		case r == '+' && i == 0:
			out = append(out, r)
		}
	}
	return string(out)
}
