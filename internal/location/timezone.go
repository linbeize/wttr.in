package location

import (
	"fmt"
	"strconv"

	"github.com/ringsaturn/tzf"
	"github.com/sirupsen/logrus"
)

// timezoneFinder is created once: tzf.NewDefaultFinder loads the embedded
// timezone polygons, which is far too expensive to repeat per request.
var timezoneFinder = func() tzf.F {
	finder, err := tzf.NewDefaultFinder()
	if err != nil {
		logrus.WithError(err).Warn("timezone finder unavailable, falling back to UTC")
		return nil
	}

	return finder
}()

// TimezoneFor returns the IANA timezone name for the given coordinates,
// or "UTC" when it cannot be determined.
func TimezoneFor(lat, lon float64) string {
	if timezoneFinder == nil {
		return "UTC"
	}

	// tzf takes (lon, lat) — the reverse of the usual argument order.
	if name := timezoneFinder.GetTimezoneName(lon, lat); name != "" {
		return name
	}

	return "UTC"
}

// enrichLocationWithTimezone fills loc.Timezone from loc.Lat/loc.Lon.
func enrichLocationWithTimezone(loc *Location) error {
	if loc == nil {
		return fmt.Errorf("nil location")
	}

	lat, err := strconv.ParseFloat(loc.Lat, 64)
	if err != nil {
		return fmt.Errorf("invalid latitude in cached location: %w", err)
	}

	lon, err := strconv.ParseFloat(loc.Lon, 64)
	if err != nil {
		return fmt.Errorf("invalid longitude in cached location: %w", err)
	}

	loc.Timezone = TimezoneFor(lat, lon)

	return nil
}
