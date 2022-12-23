package dji

import (
	"errors"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"github.com/konradit/mmt/pkg/utils"
)

var latitude = regexp.MustCompile(`\[latitude: ([+-]?(\d+\.?\d*)|(\.\d+))\]`)
var longitude = regexp.MustCompile(`\[longitude: ([+-]?(\d+\.?\d*)|(\.\d+))\]`)
var errInvalidFormat = errors.New("SRT file invalid format")
var errInvalidFile = errors.New("file invalid (not JPG or SRT)")

func GetLocation(path string) (*utils.Location, error) {
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

	latMatches := latitude.FindAllStringSubmatch(string(content), -1)

	if len(latMatches) == 0 {
		return nil, errInvalidFormat
	}

	lonMatches := longitude.FindAllStringSubmatch(string(content), -1)

	if len(lonMatches) == 0 {
		return nil, errInvalidFormat
	}

	latAsFloat, err := strconv.ParseFloat(latMatches[0][1], 64)
	if err != nil {
		return nil, err
	}
	lonAsFloat, err := strconv.ParseFloat(lonMatches[0][1], 64)
	if err != nil {
		return nil, err
	}

	return &utils.Location{Latitude: latAsFloat, Longitude: lonAsFloat}, nil
}
