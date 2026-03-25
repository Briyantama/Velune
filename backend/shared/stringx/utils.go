package stringx

import "strings"

func TrimSpace(s string) string {
	return strings.TrimSpace(s)
}

func TrimPrefix(s, prefix string) string {
	return strings.TrimPrefix(s, prefix)
}

func TrimSuffix(s, suffix string) string {
	return strings.TrimSuffix(s, suffix)
}

func Lower(s string) string {
	return strings.ToLower(s)
}

func Upper(s string) string {
	return strings.ToUpper(s)
}

func HasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

func HasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

func NewReplacer(old, new string) *strings.Replacer {
	return strings.NewReplacer(old, new)
}

func TrimRight(s, suffix string) string {
	return strings.TrimRight(s, suffix)
}

func TrimLeft(s, prefix string) string {
	return strings.TrimLeft(s, prefix)
}

func StringsEqualTrue(s string) bool {
	switch strings.TrimSpace(strings.ToLower(s)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}