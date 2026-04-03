package android

import (
	"path/filepath"
	"strings"

	mErrors "github.com/konradit/mmt/pkg/errors"
	"github.com/konradit/mmt/pkg/utils"
)

type LocationService struct{}

var ffprobe = utils.NewFFprobe(nil)

func (LocationService) GetLocation(path string) (*utils.Location, error) {
	switch strings.ToUpper(filepath.Ext(path)) {
	case ".MP4":
		return ffprobe.GPSLocation(path)
	case ".JPG":
		return utils.LocationFromEXIF(path)
	default:
		return nil, mErrors.ErrInvalidFile
	}
}
