package location

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseCoordinates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in     string
		lat    float64
		lon    float64
		wantOK bool
	}{
		{"39.9057,116.3913", 39.9057, 116.3913, true},
		{"-78.46,106.79", -78.46, 106.79, true},
		{"39.9 116.4", 39.9, 116.4, true},
		{"  31.7728 , 119.947  ", 31.7728, 119.947, true},
		// Not coordinates:
		{"", 0, 0, false},
		{"94107", 0, 0, false},
		{"New York", 0, 0, false},
		{"北京", 0, 0, false},
		{"1,2,3", 0, 0, false},
		// Null island is a placeholder, not a place.
		{"0,0", 0, 0, false},
		// Out of range.
		{"91,0", 0, 0, false},
		{"0,181", 0, 0, false},
		{"abc,def", 0, 0, false},
	}

	for _, tt := range tests {
		lat, lon, ok := parseCoordinates(tt.in)
		require.Equal(t, tt.wantOK, ok, "input %q", tt.in)

		if tt.wantOK {
			require.InDelta(t, tt.lat, lat, 1e-9, "lat for %q", tt.in)
			require.InDelta(t, tt.lon, lon, 1e-9, "lon for %q", tt.in)
		}
	}
}

func TestResolveCoordinatesRejectsInvalid(t *testing.T) {
	t.Parallel()

	_, err := ResolveCoordinates(91, 0)
	require.Error(t, err)

	_, err = ResolveCoordinates(0, 181)
	require.Error(t, err)
}

func TestResolveCoordinatesTimezone(t *testing.T) {
	t.Parallel()

	loc, err := ResolveCoordinates(39.9057, 116.3913)
	require.NoError(t, err)
	require.Equal(t, "Asia/Shanghai", loc.Timezone)

	loc, err = ResolveCoordinates(37.4129, -122.1157)
	require.NoError(t, err)
	require.Equal(t, "America/Los_Angeles", loc.Timezone)
}

// TestSearcherWithoutProviders ensures a missing geocoder configuration does
// not surface as a nil location with a nil error (which used to panic in the
// caller).
func TestSearcherWithoutProviders(t *testing.T) {
	t.Parallel()

	s := NewSearcher(&Config{})
	loc, err := s.Search("北京")

	require.Error(t, err)
	require.Nil(t, loc)
}
