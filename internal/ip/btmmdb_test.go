package ip_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/chubin/wttr.in/internal/ip"
)

// btDBPath is provided by the environment when the China-accurate database
// is available; the tests are skipped otherwise so CI stays self-contained.
const btDBPathEnv = "BT_MMDB_PATH"

func openBT(t *testing.T) *IPLocatorBT {
	t.Helper()

	path := os.Getenv(btDBPathEnv)
	if path == "" {
		t.Skipf("%s not set, skipping bt.cn-ip tests", btDBPathEnv)
	}

	locator, err := NewIPLocatorBT(path)
	require.NoError(t, err)

	t.Cleanup(func() { _ = locator.Close() })

	return locator
}

func TestIPLocatorBTChinaCity(t *testing.T) {
	locator := openBT(t)

	tests := []struct {
		ip       string
		country  string
		province string
		city     string
	}{
		{"58.216.1.1", "CN", "江苏", "常州"},
		{"114.114.114.114", "CN", "江苏", "南京"},
		{"223.5.5.5", "CN", "浙江", "杭州"},
		{"183.232.231.174", "CN", "广东", "广州"},
	}

	for _, tt := range tests {
		data, err := locator.GetIPData(tt.ip)
		require.NoError(t, err, tt.ip)
		require.Equal(t, tt.country, data.CountryCode, tt.ip)
		require.Equal(t, tt.province, data.Region, tt.ip)
		require.Equal(t, tt.city, data.City, tt.ip)

		// Coordinates must be numeric and non-empty so that the weather
		// lookup can use them directly.
		require.NotEmpty(t, data.Latitude, tt.ip)
		require.NotEmpty(t, data.Longitude, tt.ip)
	}
}

func TestIPLocatorBTIPv6Unsupported(t *testing.T) {
	locator := openBT(t)

	// The database has no IPv6 tree; the locator must report an error so the
	// chain falls through to GeoIP2 instead of returning empty geodata.
	_, err := locator.GetIPData("2408:8000::1")
	require.Error(t, err)
}

func TestIPLocatorBTInvalidInput(t *testing.T) {
	locator := openBT(t)

	_, err := locator.GetIPData("not-an-ip")
	require.Error(t, err)
}

func TestIPLocatorBTLoopbackHasNoData(t *testing.T) {
	locator := openBT(t)

	// 127.0.0.1 must never resolve to a real place: it would make local
	// development look like a city in China.
	_, err := locator.GetIPData("127.0.0.1")
	require.Error(t, err)
}
