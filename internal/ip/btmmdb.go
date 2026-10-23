package ip

import (
	"errors"
	"fmt"
	"net/netip"
	"strconv"

	"github.com/oschwald/maxminddb-golang/v2"

	"github.com/chubin/wttr.in/internal/domain"
)

// btRecord mirrors a single record of the bt.cn-ip database.
//
// The database is not MaxMind-shaped: names are Chinese and the fields are
// flat strings, so it cannot be read through geoip2.Reader.City().
type btRecord struct {
	Location struct {
		Continent   string `maxminddb:"continent"`
		Country     string `maxminddb:"country"`
		CountryEN   string `maxminddb:"country_en"`
		CountryCode string `maxminddb:"country_ens"`
		Province    string `maxminddb:"province"`
		City        string `maxminddb:"city"`
		Latitude    string `maxminddb:"latitude"`
		Longitude   string `maxminddb:"longitude"`
	} `maxminddb:"location"`
	Organization struct {
		Name string `maxminddb:"name"`
	} `maxminddb:"organization"`
	Spider struct {
		Name string `maxminddb:"name"`
	} `maxminddb:"spider"`
}

// IPLocatorBT locates IP addresses using the bt.cn-ip mmdb (China-accurate).
//
// Mounted before the MaxMind locator: it is the authoritative source for
// mainland China addresses, and IPv6 lookups (unsupported by this database)
// fall through to the next locator in the chain.
type IPLocatorBT struct {
	db *maxminddb.Reader
}

// NewIPLocatorBT opens the bt.cn-ip database at dbPath.
func NewIPLocatorBT(dbPath string) (*IPLocatorBT, error) {
	db, err := maxminddb.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open bt.cn-ip database: %w", err)
	}

	return &IPLocatorBT{db: db}, nil
}

// Close releases the underlying database reader resources.
func (l *IPLocatorBT) Close() error {
	if l.db != nil {
		return l.db.Close()
	}

	return nil
}

// GetIPData implements the IPLocator interface.
func (l *IPLocatorBT) GetIPData(ipStr string) (*domain.IPData, error) {
	addr, err := netip.ParseAddr(ipStr)
	if err != nil {
		return nil, fmt.Errorf("invalid IP address: %w", err)
	}

	// The database is IPv4-only in content; drop the ::ffff: prefix so that
	// IPv4-mapped clients still resolve instead of being treated as IPv6.
	addr = addr.Unmap()

	if !addr.Is4() {
		return nil, errors.New("bt.cn-ip database covers IPv4 only")
	}

	// Private and reserved ranges resolve to a "保留" (reserved) placeholder
	// record with no ISO code and no coordinates. Treating that as real
	// geodata would label local clients with a bogus country.
	if !addr.IsGlobalUnicast() || addr.IsPrivate() || addr.IsLoopback() {
		return nil, errors.New("no geolocation data found for this IP")
	}

	var record btRecord

	res := l.db.Lookup(addr)
	if !res.Found() {
		return nil, errors.New("no geolocation data found for this IP")
	}

	if err := res.Decode(&record); err != nil {
		return nil, fmt.Errorf("lookup failed: %w", err)
	}

	// A usable record must carry a real ISO country code. Placeholder rows
	// ("保留", empty code) are rejected so the chain falls through.
	if !isISOCountryCode(record.Location.CountryCode) {
		return nil, errors.New("no geolocation data found for this IP")
	}

	data := &domain.IPData{
		IP:          ipStr,
		CountryCode: record.Location.CountryCode,
		Country:     record.Location.Country,
		Region:      record.Location.Province,
		City:        record.Location.City,
	}

	// Coordinates come back as strings; normalise them so downstream
	// ParseFloat calls never see an empty or malformed value.
	if lat, lon, ok := normalizeCoordPair(record.Location.Latitude, record.Location.Longitude); ok {
		data.Latitude = lat
		data.Longitude = lon
	}

	return data, nil
}

// normalizeCoordPair validates and reshapes a lat/lon string pair.
func normalizeCoordPair(lat, lon string) (string, string, bool) {
	latF, err := strconv.ParseFloat(lat, 64)
	if err != nil {
		return "", "", false
	}

	lonF, err := strconv.ParseFloat(lon, 64)
	if err != nil {
		return "", "", false
	}

	// Reject the (0,0) null island placeholder and out-of-range values.
	if latF == 0 && lonF == 0 {
		return "", "", false
	}

	if latF < -90 || latF > 90 || lonF < -180 || lonF > 180 {
		return "", "", false
	}

	return strconv.FormatFloat(latF, 'f', 6, 64), strconv.FormatFloat(lonF, 'f', 6, 64), true
}

// isISOCountryCode reports whether code looks like an ISO 3166-1 alpha-2 code.
func isISOCountryCode(code string) bool {
	if len(code) != 2 {
		return false
	}

	for i := 0; i < 2; i++ {
		if code[i] < 'A' || code[i] > 'Z' {
			return false
		}
	}

	return true
}
