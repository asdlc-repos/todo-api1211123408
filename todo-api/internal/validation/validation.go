package validation

import (
	"regexp"
	"strings"
	"unicode"
)

var emailRegex = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)

func ValidEmail(s string) bool {
	if len(s) == 0 || len(s) > 254 {
		return false
	}
	return emailRegex.MatchString(s)
}

func ValidPassword(pw string) bool {
	if len(pw) < 8 {
		return false
	}
	var hasUpper, hasLower, hasDigit bool
	for _, r := range pw {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	return hasUpper && hasLower && hasDigit
}

func Sanitize(s string) (string, bool) {
	if strings.ContainsRune(s, '\x00') {
		return "", false
	}
	return strings.TrimSpace(s), true
}
