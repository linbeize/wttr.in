package util

import "strings"

func InSlice(what string, where []string) bool {
	for _, item := range where {
		if item == what {
			return true
		}
	}
	return false
}

// NonEmptyStrings drops empty values and joins the rest with sep.
// Used to build composite location strings without dangling separators
// (e.g. "北京, 北京, CN" vs "北京, , CN").
func NonEmptyStrings(sep string, values ...string) string {
	nonEmpty := make([]string, 0, len(values))
	for _, v := range values {
		if v = strings.TrimSpace(v); v != "" {
			nonEmpty = append(nonEmpty, v)
		}
	}

	return strings.Join(nonEmpty, sep)
}

func HasPrefixInSlice(what string, where []string, prefix string) bool {
	for _, item := range where {
		item = item + prefix
		if strings.HasPrefix(what, item) {
			return true
		}
	}
	return false
}
