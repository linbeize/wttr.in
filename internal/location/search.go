package location

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/chubin/wttr.in/internal/types"
)

type Provider interface {
	Name() string
	Query(location string) (*Location, error)
}

type Searcher struct {
	providers []Provider
}

// NewSearcher returns a new Searcher for the specified config.
func NewSearcher(config *Config) *Searcher {
	providers := []Provider{}

	for _, p := range config.NominatimServers {
		n := NewNominatim(p.Name, p.Type, p.URL, p.Token)
		if p.Lang != "" {
			n.WithLang(p.Lang)
		}

		providers = append(providers, n)
	}

	return &Searcher{
		providers: providers,
	}
}

// Search makes queries through all known providers,
// and returns response, as soon as it is not nil.
// If all responses were nil, the last response is returned.
func (s *Searcher) Search(location string) (*Location, error) {
	var (
		err    error
		result *Location
	)

	for _, p := range s.providers {
		logrus.Debugln("querying ", p.Name())
		result, err = p.Query(location)
		if result != nil && err == nil {
			return result, nil
		}
	}

	// No provider (or none of them) returned a usable location: reporting
	// (nil, nil) here would make callers dereference a nil *Location.
	if result == nil && err == nil {
		err = fmt.Errorf("%w: %s", types.ErrNotFound, location)
	}

	return result, err
}
