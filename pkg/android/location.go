package android

import (
	"errors"
	"strings"

	"github.com/konradit/mmt/pkg/utils"
)

var errInvalidFile = errors.New("file invalid (not a video or photo)")

type LocationService struct{}

var ffprobe = utils.NewFFprobe(nil)

func (LocationService) GetLocation(path string) (*utils.Location, error) {
	switch true {
	case strings.Contains(strings.ToLower(path), ".mp4"):
		return ffprobe.GPSLocation(path)
	case strings.Contains(strings.ToLower(path), ".jpg"):
		return utils.LocationFromEXIF(path)
	default:
		return nil, errInvalidFile
	}
}
