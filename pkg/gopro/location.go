package gopro

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"

	"github.com/konradit/gopro-utils/telemetry"
	mErrors "github.com/konradit/mmt/pkg/errors"
	"github.com/konradit/mmt/pkg/utils"
	"github.com/konradit/mmt/pkg/videomanipulation"
	"golang.org/x/exp/slices"
)

type LocationService struct{}

func (LocationService) GetLocation(path string) (*utils.Location, error) {
	switch true {
	case strings.Contains(path, ".MP4"):
		return fromMP4(path)
	case strings.Contains(path, ".WAV"):
		return fromMP4(path[:len(path)-len(filepath.Ext(path))] + ".MP4")
	case strings.Contains(path, ".GPR"):
		return utils.LocationFromEXIF(path)
	case strings.Contains(path, ".JPG"):
		return utils.LocationFromEXIF(path)
	default:
		return nil, mErrors.ErrInvalidFile
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
			if telem.Speed == 0 || telem.Latitude == 0 || telem.Longitude == 0 || telem.GpsAccuracy > gpsMinAccuracyFromConfig() || !slices.Contains(gpsLockTypesFromConfig(), int(telem.GpsFix)) {
				continue
			}
			return &utils.Location{
				Latitude:  telem.Latitude,
				Longitude: telem.Longitude,
			}, nil
		}
		*lastEvent = *event
	}

	return nil, mErrors.ErrNoGPS
}
