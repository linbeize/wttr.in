package location

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/chubin/wttr.in/internal/types"
)

type Nominatim struct {
	name  string
	url   string
	token string
	typ   string
	// lang is the preferred language for geocoding results (accept-language).
	lang string
}

type locationQuerier interface {
	Query(*Nominatim, string) (*Location, error)
}

func NewNominatim(name, typ, url, token string) *Nominatim {
	return &Nominatim{
		name:  name,
		url:   url,
		token: token,
		typ:   typ,
	}
}

// WithLang sets the preferred geocoding language and returns the provider.
func (n *Nominatim) WithLang(lang string) *Nominatim {
	n.lang = lang
	return n
}

func (n *Nominatim) Name() string {
	return fmt.Sprintf("%s (%s)", n.name, n.typ)
}

func (n *Nominatim) Query(location string) (*Location, error) {
	var data locationQuerier

	switch n.typ {
	case "iq":
		data = &locationIQ{}
	case "opencage":
		data = &locationOpenCage{}
	case "osm":
		data = &locationOSM{}
	default:
		return nil, fmt.Errorf("%s: %w", n.name, types.ErrUnknownLocationService)
	}

	return data.Query(n, location)
}

// geocodingClient is shared so that connections to the geocoder are reused.
// A timeout is mandatory: a hanging geocoder would otherwise block the
// whole weather response.
var geocodingClient = &http.Client{
	Timeout: 8 * time.Second,
}

// geocodingUserAgent identifies this service to public geocoders.
// The public Nominatim endpoint rejects requests without a User-Agent.
const geocodingUserAgent = "wttr.in/2.0 (+https://wttr.in)"

func makeQuery(url string, result interface{}) error {
	var errResponse struct {
		Error   string
		Message string
	}

	log.Debugln("nominatim:", url)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", geocodingUserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := geocodingClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: http %d: %s", types.ErrUpstream, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if err = json.Unmarshal(body, &errResponse); err == nil {
		if errResponse.Error != "" {
			return fmt.Errorf("%w: %s", types.ErrUpstream, errResponse.Error)
		}

		if errResponse.Message != "" {
			return fmt.Errorf("%w: %s", types.ErrUpstream, errResponse.Message)
		}
	}

	log.Debugln("nominatim: response: ", string(body))

	if err = json.Unmarshal(body, &result); err != nil {
		return err
	}

	return nil
}
