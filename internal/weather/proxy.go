package weather

import (
	"errors"
	"net"
	"net/http"
	"net/netip"
	"strings"
)

// ProxyConfig controls how client IPs are derived when the service sits
// behind a reverse proxy.
//
// The default (no proxies configured) trusts no forwarding header and uses
// RemoteAddr, which is the only safe choice when the service is exposed
// directly: anybody can send a forged X-Forwarded-For.
type ProxyConfig struct {
	// TrustedProxies lists CIDRs of proxies allowed to set the forwarding
	// headers, e.g. ["127.0.0.1/32", "::1/128"] for a proxy on the same host
	// or ["10.0.0.0/8"] for a container network.
	TrustedProxies []string `yaml:"trustedProxies"`
}

// proxyTrust holds the parsed trusted proxy ranges.
type proxyTrust struct {
	prefixes []netip.Prefix
}

// newProxyTrust parses the configured CIDRs. Invalid entries are ignored
// (and reported by the caller) so a typo cannot silently disable the proxy
// support the operator asked for.
func newProxyTrust(raw []string) (*proxyTrust, []error) {
	pt := &proxyTrust{}
	var errs []error

	for _, entry := range raw {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		// Accept a bare IP as a single-host range: "127.0.0.1" -> /32.
		if !strings.Contains(entry, "/") {
			addr, err := netip.ParseAddr(entry)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			pt.prefixes = append(pt.prefixes, netip.PrefixFrom(addr, addr.BitLen()))

			continue
		}

		prefix, err := netip.ParsePrefix(entry)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		pt.prefixes = append(pt.prefixes, prefix.Masked())
	}

	return pt, errs
}

// empty reports whether no trusted proxies are configured.
func (p *proxyTrust) empty() bool {
	return p == nil || len(p.prefixes) == 0
}

// contains reports whether addr is one of the trusted proxies.
func (p *proxyTrust) contains(addr netip.Addr) bool {
	if p == nil {
		return false
	}

	for _, prefix := range p.prefixes {
		if prefix.Contains(addr) {
			return true
		}
	}

	return false
}

// resolveClientIP returns the client address of r.
//
// Forwarding headers are honoured only when the direct peer (RemoteAddr) is a
// configured trusted proxy. In that case the chain is walked from right to
// left and the first address that is *not* itself a trusted proxy wins —
// proxies append entries, so the rightmost untrusted hop is the one the
// proxy actually observed. Trusting the leftmost entry (the previous
// behaviour) let any client spoof its location with a single header.
func (s *WeatherService) resolveClientIP(r *http.Request) string {
	direct := remoteHost(r.RemoteAddr)

	if s.trust.empty() || !s.trust.contains(parseAddr(direct)) {
		// Not behind a (configured) proxy: the peer is the client.
		if direct == "" {
			return r.RemoteAddr
		}

		return direct
	}

	if ip := clientFromForwardedFor(r.Header.Get("X-Forwarded-For"), s.trust); ip != "" {
		return ip
	}

	if realIP := netipAddrHeader(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}

	if direct == "" {
		return r.RemoteAddr
	}

	return direct
}

// clientFromForwardedFor extracts the client address from an X-Forwarded-For
// value, walking right to left and skipping trusted proxies.
func clientFromForwardedFor(header string, trust *proxyTrust) string {
	if header == "" {
		return ""
	}

	parts := strings.Split(header, ",")
	// Walk backwards: the rightmost entries were appended by the closest hops.
	for i := len(parts) - 1; i >= 0; i-- {
		candidate := strings.TrimSpace(parts[i])
		if candidate == "" {
			continue
		}

		addr := parseAddr(candidate)
		if !addr.IsValid() {
			continue
		}

		if trust.contains(addr) {
			// Another proxy in the chain: keep looking left.
			continue
		}

		return addr.String()
	}

	return ""
}

// netipAddrHeader validates a single-address header value.
func netipAddrHeader(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	if addr := parseAddr(value); addr.IsValid() {
		return addr.String()
	}

	return ""
}

// parseAddr parses an address, tolerating IPv6 zone and brackets.
func parseAddr(s string) netip.Addr {
	s = strings.TrimSpace(s)
	if s == "" {
		return netip.Addr{}
	}

	// Strip an optional ":port" (X-Forwarded-For never has one, X-Real-IP
	// sometimes does).
	if host, _, err := net.SplitHostPort(s); err == nil {
		s = host
	}

	addr, err := netip.ParseAddr(s)
	if err != nil {
		return netip.Addr{}
	}

	return addr.Unmap()
}

// remoteHost extracts the host part of a RemoteAddr value.
func remoteHost(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}

	return host
}

// errNoClientIP is returned when the client address cannot be determined.
var errNoClientIP = errors.New("cannot determine client IP")
