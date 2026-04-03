package dji

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	mErrors "github.com/konradit/mmt/pkg/errors"
	"github.com/konradit/mmt/pkg/utils"
)

type LatLongPair struct {
	Latitude, Longitude *regexp.Regexp
	Indicator           string
}

type ACType int

const (
	NewAircraft ACType = iota
	OldAircraft
)

var allDrones = map[ACType]LatLongPair{
	NewAircraft: {
		Latitude:  regexp.MustCompile(`\[latitude[ ]?: ([+-]?(\d+\.?\d*)|(\.\d+))\]`),
		Longitude: regexp.MustCompile(`\[long[t]?itude[ ]?: ([+-]?(\d+\.?\d*)|(\.\d+))\]`), // DJI and their typos...
	},

	OldAircraft: {
		Latitude:  regexp.MustCompile(`GPS[ ]?\(([+-]?(\d+\.?\d*)|(\.\d+))`),
		Longitude: regexp.MustCompile(`,[ ]?([+-]?(\d+\.?\d*)|(\.\d+)),[ ]?\d+\)`),
	},
}

type LocationService struct{}

func (LocationService) GetLocation(path string) (*utils.Location, error) {
	switch strings.ToUpper(filepath.Ext(path)) {
	case ".MP4", ".SRT":
		return fromSRT(path)
	case ".JPG", ".DNG":
		return utils.LocationFromEXIF(path)
	default:
		return nil, mErrors.ErrInvalidFile
	}
}

func fromSRT(srtPath string) (*utils.Location, error) {
	fs, err := os.Open(strings.ReplaceAll(srtPath, ".MP4", ".SRT"))
	if err != nil {
		return nil, err
	}
	defer fs.Close()

	reader := bufio.NewReader(fs)
	limitedSizeReader := io.LimitReader(reader, 2048)

	content, err := io.ReadAll(limitedSizeReader)
	if err != nil {
		return nil, err
	}

	var latAsFloat, lonAsFloat float64

	for _, drone := range allDrones {
		latMatches := drone.Latitude.FindAllStringSubmatch(string(content), -1)

		lonMatches := drone.Longitude.FindAllStringSubmatch(string(content), -1)

		if len(lonMatches) == 0 || len(latMatches) == 0 {
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

	return nil, mErrors.ErrNoRecognizedSRTFormat
}
