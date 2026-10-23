package weather

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func newRequestWithHeaders(t *testing.T, remoteAddr string, headers map[string]string) *http.Request {
	t.Helper()

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = remoteAddr

	for k, v := range headers {
		r.Header.Set(k, v)
	}

	return r
}

func serviceWithProxies(t *testing.T, cidrs ...string) *WeatherService {
	t.Helper()

	s := &WeatherService{}
	s.SetTrustedProxies(cidrs)

	return s
}

// TestResolveClientIPNoProxiesConfigured is the safe default: forwarding
// headers must be ignored entirely, otherwise anyone can spoof their location.
func TestResolveClientIPNoProxiesConfigured(t *testing.T) {
	t.Parallel()

	s := serviceWithProxies(t)

	r := newRequestWithHeaders(t, "203.0.113.9:5678", map[string]string{
		"X-Forwarded-For": "1.2.3.4",
		"X-Real-IP":       "5.6.7.8",
	})

	require.Equal(t, "203.0.113.9", s.resolveClientIP(r))
}

// TestResolveClientIPUntrustedPeer ensures a direct client cannot bypass the
// trust check by sending the headers itself.
func TestResolveClientIPUntrustedPeer(t *testing.T) {
	t.Parallel()

	s := serviceWithProxies(t, "127.0.0.1/32")

	r := newRequestWithHeaders(t, "198.51.100.7:1234", map[string]string{
		"X-Forwarded-For": "58.216.1.1",
	})

	require.Equal(t, "198.51.100.7", s.resolveClientIP(r),
		"headers from an untrusted peer must be ignored")
}

// TestResolveClientIPSingleProxy covers the normal nginx setup: one trusted
// hop that appends the real client address.
func TestResolveClientIPSingleProxy(t *testing.T) {
	t.Parallel()

	s := serviceWithProxies(t, "127.0.0.1/32")

	r := newRequestWithHeaders(t, "127.0.0.1:4321", map[string]string{
		"X-Forwarded-For": "58.216.1.1",
	})

	require.Equal(t, "58.216.1.1", s.resolveClientIP(r))
}

// TestResolveClientIPSpoofedLeftmost is the regression this whole mechanism
// exists for: the client sends a forged X-Forwarded-For, nginx appends the
// real address, so the header becomes "forged, real". Reading from the right
// yields the real client; reading from the left (the old behaviour) yields
// whatever the attacker chose.
func TestResolveClientIPSpoofedLeftmost(t *testing.T) {
	t.Parallel()

	s := serviceWithProxies(t, "127.0.0.1/32")

	r := newRequestWithHeaders(t, "127.0.0.1:4321", map[string]string{
		"X-Forwarded-For": "1.1.1.1, 58.216.1.1",
	})

	require.Equal(t, "58.216.1.1", s.resolveClientIP(r))
}

// TestResolveClientIPMultiHop walks a chain of trusted proxies.
func TestResolveClientIPMultiHop(t *testing.T) {
	t.Parallel()

	s := serviceWithProxies(t, "127.0.0.1/32", "10.0.0.0/8")

	r := newRequestWithHeaders(t, "127.0.0.1:4321", map[string]string{
		"X-Forwarded-For": "58.216.1.1, 10.0.0.5, 10.0.0.6",
	})

	require.Equal(t, "58.216.1.1", s.resolveClientIP(r))
}

// TestResolveClientIPRealIPFallback covers proxies that only set X-Real-IP.
func TestResolveClientIPRealIPFallback(t *testing.T) {
	t.Parallel()

	s := serviceWithProxies(t, "127.0.0.1/32")

	r := newRequestWithHeaders(t, "127.0.0.1:4321", map[string]string{
		"X-Real-IP": "58.216.1.1",
	})

	require.Equal(t, "58.216.1.1", s.resolveClientIP(r))
}

// TestResolveClientIPIgnoresGarbage keeps a malformed header from becoming
// the client identity.
func TestResolveClientIPIgnoresGarbage(t *testing.T) {
	t.Parallel()

	s := serviceWithProxies(t, "127.0.0.1/32")

	r := newRequestWithHeaders(t, "127.0.0.1:4321", map[string]string{
		"X-Forwarded-For": "not-an-ip",
		"X-Real-IP":       "also-garbage",
	})

	require.Equal(t, "127.0.0.1", s.resolveClientIP(r))
}

// TestResolveClientIPIPv6 ensures IPv6 peers and values work, including the
// bracketed form RemoteAddr uses.
func TestResolveClientIPIPv6(t *testing.T) {
	t.Parallel()

	s := serviceWithProxies(t, "::1/128")

	r := newRequestWithHeaders(t, "[::1]:4321", map[string]string{
		"X-Forwarded-For": "2408:8000::1",
	})

	require.Equal(t, "2408:8000::1", s.resolveClientIP(r))
}

// TestResolveClientIPIPv4Mapped normalises ::ffff: addresses so the IP
// database (IPv4-only) can match them.
func TestResolveClientIPIPv4Mapped(t *testing.T) {
	t.Parallel()

	s := serviceWithProxies(t, "127.0.0.1/32")

	r := newRequestWithHeaders(t, "127.0.0.1:4321", map[string]string{
		"X-Forwarded-For": "::ffff:58.216.1.1",
	})

	require.Equal(t, "58.216.1.1", s.resolveClientIP(r))
}

// TestBuildCacheKeyUsesResolvedIP guards the cache-poisoning angle: two
// requests differing only in a forged header must not share a cache key
// unless the header is trusted.
func TestBuildCacheKeyUsesResolvedIP(t *testing.T) {
	t.Parallel()

	s := serviceWithProxies(t) // no trusted proxies

	r1 := newRequestWithHeaders(t, "203.0.113.9:1111", map[string]string{
		"User-Agent": "curl/8",
	})
	r2 := newRequestWithHeaders(t, "203.0.113.9:1111", map[string]string{
		"User-Agent":      "curl/8",
		"X-Forwarded-For": "1.2.3.4",
	})

	require.Equal(t, buildCacheKey(r1, s.resolveClientIP(r1)),
		buildCacheKey(r2, s.resolveClientIP(r2)),
		"a forged header must not change the cache key")
}

// TestNewProxyTrustAcceptsBareIP covers the convenience form.
func TestNewProxyTrustAcceptsBareIP(t *testing.T) {
	t.Parallel()

	trust, errs := newProxyTrust([]string{"127.0.0.1", "10.0.0.0/8", "", "bogus"})
	require.Len(t, errs, 1, "only the bogus entry should error")
	require.False(t, trust.empty())

	require.True(t, trust.contains(parseAddr("127.0.0.1")))
	require.True(t, trust.contains(parseAddr("10.9.9.9")))
	require.False(t, trust.contains(parseAddr("11.0.0.1")))
}
