package ip

type Config struct {
	IPCacheDB   string `yaml:"ipCacheDb"`
	IPCache     string `yaml:"ipCache"`
	IPCacheType string `yaml:"ipCacheType"`
	GeoIP2      string `yaml:"geoip2"`
	// BTGeoIP points to a bt.cn-ip mmdb (China-accurate, IPv4 only).
	// When set it is consulted before GeoIP2.
	BTGeoIP string `yaml:"btGeoip"`
}
