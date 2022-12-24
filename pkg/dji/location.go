package dji

import (
	"errors"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"github.com/konradit/mmt/pkg/utils"
)

type LatLongPair struct {
	Latitude, Longitude *regexp.Regexp
	Indicator           string
}

var allDrones = map[string]LatLongPair{
	"NewMavics": {
		Latitude:  regexp.MustCompile(`\[latitude[ ]?: ([+-]?(\d+\.?\d*)|(\.\d+))\]`),
		Longitude: regexp.MustCompile(`\[long[t]?itude[ ]?: ([+-]?(\d+\.?\d*)|(\.\d+))\]`), // DJI and their typos...
	},

	"OldMavics": {
		Latitude:  regexp.MustCompile(`GPS[ ]?\(([+-]?(\d+\.?\d*)|(\.\d+))`),
		Longitude: regexp.MustCompile(`,[ ]?([+-]?(\d+\.?\d*)|(\.\d+)),[ ]?\d+\)`),
	},
}

var errInvalidFormat = errors.New("SRT file invalid format (could not read from predefined presets)")
var errInvalidFile = errors.New("file invalid (not JPG or SRT)")

type LocationService struct{}

func (LocationService) GetLocation(path string) (*utils.Location, error) {
	switch true {
	case strings.Contains(path, ".MP4") || strings.Contains(path, ".SRT"):
		return fromSRT(path)
	case strings.Contains(path, ".JPG") || strings.Contains(path, ".DNG"):
		return utils.LocationFromPhoto(path)
	default:
		return nil, errInvalidFile
	}
}
func fromSRT(srtPath string) (*utils.Location, error) {
	content, err := ioutil.ReadFile(strings.Replace(srtPath, ".MP4", ".SRT", -1))
	if err != nil {
		return nil, err
	}

	latAsFloat, lonAsFloat := float64(0), float64(0)

	for _, drone := range allDrones {
		latMatches := drone.Latitude.FindAllStringSubmatch(string(content), -1)

		if len(latMatches) == 0 {
			continue
		}

		lonMatches := drone.Longitude.FindAllStringSubmatch(string(content), -1)

		if len(lonMatches) == 0 {
			continue
		}

		latAsFloat, err = strconv.ParseFloat(latMatches[0][1], 64)
		if err != nil {
			return nil, err
		}
		lonAsFloat, err = strconv.ParseFloat(lonMatches[0][1], 64)
		if err != nil {
			return nil, err
		}
		return &utils.Location{Latitude: latAsFloat, Longitude: lonAsFloat}, nil
	}
	return nil, errInvalidFormat
}
