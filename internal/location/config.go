package location

type Config struct {
	LocationCacheType string `yaml:"locationCacheType"`
	LocationCacheDB   string `yaml:"locationCacheDb"`
	LocationCache     string `yaml:"locationCache"`
	NominatimServers  []struct {
		Name  string `yaml:"name"`
		Type  string `yaml:"type"`
		URL   string `yaml:"url"`
		Token string `yaml:"token"`
		// Lang is the preferred result language (accept-language) for
		// geocoders that support it, e.g. "zh" for Nominatim.
		Lang string `yaml:"lang"`
	} `yaml:"nominatim"`
	// CoordinatesOnly skips geocoding entirely and treats the location
	// string as "lat,lon" when it parses as a coordinate pair.
	CoordinatesOnly bool `yaml:"coordinatesOnly"`
}
