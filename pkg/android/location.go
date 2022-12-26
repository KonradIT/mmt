package android

import (
	"strings"

	mErrors "github.com/konradit/mmt/pkg/errors"
	"github.com/konradit/mmt/pkg/utils"
)

type LocationService struct{}

var ffprobe = utils.NewFFprobe(nil)

func (LocationService) GetLocation(path string) (*utils.Location, error) {
	switch true {
	case strings.Contains(strings.ToLower(path), ".mp4"):
		return ffprobe.GPSLocation(path)
	case strings.Contains(strings.ToLower(path), ".jpg"):
		return utils.LocationFromEXIF(path)
	default:
		return nil, mErrors.ErrInvalidFile
	}
}
