package config

// SiteConfig holds site-wide defaults that are not tied to a subsystem.
type SiteConfig struct {
	// DefaultLang is the language used when the request does not specify one
	// (no ?lang=, no language subdomain, no Accept-Language header).
	// Empty falls back to "en", the historical behaviour.
	DefaultLang string `yaml:"defaultLang"`

	// DefaultLocation is used when neither the request nor the client IP
	// yields a location (e.g. loopback clients during local development).
	DefaultLocation string `yaml:"defaultLocation"`
}
