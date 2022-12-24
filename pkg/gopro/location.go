package gopro

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"github.com/konradit/gopro-utils/telemetry"
	"github.com/konradit/mmt/pkg/utils"
	"github.com/konradit/mmt/pkg/videomanipulation"
)

var errNoGPS = errors.New("No GPS(5) data found")
var errInvalidFile = errors.New("file invalid (not MP4 or JPG)")
var noGPSFix = 9999

type LocationService struct{}

func (LocationService) GetLocation(path string) (*utils.Location, error) {
	switch true {
	case strings.Contains(path, ".MP4"):
		return fromMP4(path)
	case strings.Contains(path, ".JPG"):
		return utils.LocationFromPhoto(path)
	default:
		return nil, errInvalidFile
	}
}
func fromMP4(videoPath string) (*utils.Location, error) {
	vman := videomanipulation.New()
	data, err := vman.ExtractGPMF(videoPath)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(*data)

	lastEvent := &telemetry.TELEM{}

	for {
		event, err := telemetry.Read(reader)
		if err != nil && err != io.EOF {
			return nil, err
		} else if err == io.EOF || event == nil {
			break
		}

		if lastEvent.IsZero() {
			*lastEvent = *event
			event.Clear()
			continue
		}

		err = lastEvent.FillTimes(event.Time.Time)
		if err != nil {
			return nil, err
		}

		telems := lastEvent.ShitJson()
		for _, telem := range telems {
			if telem.Latitude == 0 && telem.Longitude == 0 || telem.GpsAccuracy == uint16(noGPSFix) {
				continue
			}
			return &utils.Location{
				Latitude:  telem.Latitude,
				Longitude: telem.Longitude,
			}, nil
		}
		*lastEvent = *event
	}

	return nil, errNoGPS
}
