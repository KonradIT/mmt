package gopro

import (
	"bytes"
	"io"
	"math"
	"path/filepath"
	"strings"

	"github.com/codingsince1985/geo-golang/openstreetmap"
	"github.com/konradit/gopro-utils/telemetry"
	mErrors "github.com/konradit/mmt/pkg/errors"
	"github.com/konradit/mmt/pkg/utils"
	"github.com/konradit/mmt/pkg/videomanipulation"
	"golang.org/x/exp/slices"
)

type LocationService struct{}

type Location struct {
	latitude  float64
	longitude float64
}

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

	GPSNum := 0
	reader := bytes.NewReader(*data)

	lastEvent := &telemetry.TELEM{}
	coordinates := []Location{}

GetLocation:
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
			if telem.Altitude > gpsMaxAltitudeFromConfig() || telem.Latitude == 0 || telem.Longitude == 0 || telem.GpsAccuracy > gpsMinAccuracyFromConfig() {
				continue
			}

			CountryCodes := gpsCountryCodesFromConfig()
			if len(CountryCodes) != 0 {
				service := openstreetmap.Geocoder()

				address, err := service.ReverseGeocode(telem.Latitude, telem.Longitude)
				if err != nil || !slices.Contains(CountryCodes, address.CountryCode) {
					continue
				}

				GPSNum++
			}

			coordinates = append(coordinates, Location{telem.Latitude, telem.Longitude})

			if GPSNum > gpsMaxCountryCodesFromConfig() {
				break GetLocation
			}
		}
		*lastEvent = *event
	}

	if len(coordinates) == 0 {
		return nil, mErrors.ErrNoGPS
	}

	closestLocation := getClosestLocation(coordinates)

	return &utils.Location{
		Latitude:  closestLocation.latitude,
		Longitude: closestLocation.longitude,
	}, nil
}

func getClosestLocation(locations []Location) Location {
	// Find the nearest and repeat location
	counts := make(map[Location]int)
	for _, loc := range locations {
		counts[loc]++
	}

	mostFrequentLocation := Location{}
	maxCount := 0
	for loc, count := range counts {
		if count > maxCount {
			mostFrequentLocation = loc
			maxCount = count
		} else if count == maxCount {
			// If there are multiple locations with the same frequency, choose the closest one
			distanceToLoc := distance(locations[0], loc)
			distanceToMostFrequent := distance(locations[0], mostFrequentLocation)
			if distanceToLoc < distanceToMostFrequent {
				mostFrequentLocation = loc
			}
		}
	}

	return mostFrequentLocation
}

func distance(loc1 Location, loc2 Location) float64 {
	lat1 := degreesToRadians(loc1.latitude)
	lon1 := degreesToRadians(loc1.longitude)
	lat2 := degreesToRadians(loc2.latitude)
	lon2 := degreesToRadians(loc2.longitude)

	deltaLat := lat2 - lat1
	deltaLon := lon2 - lon1

	a := math.Pow(math.Sin(deltaLat/2), 2) + math.Cos(lat1)*math.Cos(lat2)*math.Pow(math.Sin(deltaLon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return 6371 * c // radius of the earth in km
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}
