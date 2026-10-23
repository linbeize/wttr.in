package weather

import "github.com/sirupsen/logrus"

// SetTrustedProxies configures which reverse proxies may set the client IP
// headers. Invalid CIDRs are logged and skipped rather than aborting startup:
// a typo in one entry should not take the whole service down, but it must be
// visible in the logs.
func (s *WeatherService) SetTrustedProxies(cidrs []string) {
	trust, errs := newProxyTrust(cidrs)
	for _, err := range errs {
		logrus.WithError(err).Warn("ignoring invalid trustedProxies entry")
	}

	s.trust = trust

	if trust.empty() {
		logrus.Info("no trusted proxies configured: client IP taken from RemoteAddr")

		return
	}

	logrus.WithField("proxies", cidrs).Info("trusting forwarding headers from configured proxies")
}
