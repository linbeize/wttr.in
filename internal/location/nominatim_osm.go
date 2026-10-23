package location

import (
	"fmt"
	"net/url"

	"github.com/chubin/wttr.in/internal/types"
)

// locationOSM queries a plain Nominatim instance (e.g. the public
// nominatim.openstreetmap.org endpoint).
//
// Unlike the commercial providers this one has no token, and it accepts
// accept-language, which is what makes Chinese place names work
// ("北京", "朝阳区", "乌鲁木齐").
type locationOSM []struct {
	Lat     string `json:"lat"`
	Lon     string `json:"lon"`
	Name    string `json:"name"`
	Address struct {
		Country     string `json:"country"`
		CountryCode string `json:"country_code"`
	} `json:"address"`
	//nolint:tagliatelle
	DisplayName string `json:"display_name"`
}

func (data *locationOSM) Query(n *Nominatim, location string) (*Location, error) {
	lang := n.lang
	if lang == "" {
		lang = "zh"
	}

	query := fmt.Sprintf(
		"%s?q=%s&format=jsonv2&limit=1&addressdetails=1&accept-language=%s",
		n.url, url.QueryEscape(location), url.QueryEscape(lang))

	if n.token != "" {
		query += "&email=" + url.QueryEscape(n.token)
	}

	if err := makeQuery(query, data); err != nil {
		return nil, fmt.Errorf("%s: %w", n.name, err)
	}

	if len(*data) != 1 {
		return nil, fmt.Errorf("%w: %s: invalid response", types.ErrUpstream, n.name)
	}

	nl := &(*data)[0]

	return &Location{
		Lat:      nl.Lat,
		Lon:      nl.Lon,
		Fullname: nl.DisplayName,
	}, nil
}

// ResolveCoordinates builds a Location straight from a lat/lon pair,
// bypassing geocoding entirely.
//
// Used for IP-based auto detection: the IP database already knows the
// coordinates, and reverse geocoding them through a third-party service
// would only add latency and failure modes.
func ResolveCoordinates(lat, lon float64) (*Location, error) {
	if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		return nil, fmt.Errorf("%w: invalid coordinates %v,%v", types.ErrInvalidCacheEntry, lat, lon)
	}

	loc := &Location{
		Lat: fmt.Sprintf("%.6f", lat),
		Lon: fmt.Sprintf("%.6f", lon),
	}

	if err := enrichLocationWithTimezone(loc); err != nil {
		return nil, err
	}

	return loc, nil
}
