package query

import "testing"

func TestIsValidLanguageCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		code string
		want bool
	}{
		{"en", true},
		{"zh", true},
		{"de", true},
		{"zh-cn", true},
		{"zh-tw", true},
		{"zh-CN", true},
		{"", false},
		{"e", false},
		{"english", false},
		{"zh-cn-extra", false},
		{"v2", false},
		{"1a", false},
	}

	for _, tt := range tests {
		if got := isValidLanguageCode(tt.code); got != tt.want {
			t.Errorf("isValidLanguageCode(%q) = %v, want %v", tt.code, got, tt.want)
		}
	}
}

func TestNormalizeLanguageTag(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"zh":      "zh-cn",
		"zh-CN":   "zh-cn",
		"zh-Hans": "zh-cn",
		"zh-SG":   "zh-cn",
		"zh-TW":   "zh-tw",
		"zh-Hant": "zh-tw",
		"zh-HK":   "zh-tw",
		"en":      "en",
		"de":      "de",
		"":        "",
	}

	for in, want := range tests {
		if got := normalizeLanguageTag(in); got != want {
			t.Errorf("normalizeLanguageTag(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLanguageFromAcceptLanguage(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"zh-CN,zh;q=0.9,en;q=0.8": "zh-cn",
		"zh-TW":                   "zh-tw",
		"en-US,en;q=0.9":          "en-us",
		"de":                      "de",
		"":                        "",
		";;;":                     "",
		"*":                       "",
	}

	for in, want := range tests {
		if got := languageFromAcceptLanguage(in); got != want {
			t.Errorf("languageFromAcceptLanguage(%q) = %q, want %q", in, got, want)
		}
	}
}
