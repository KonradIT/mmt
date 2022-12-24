package utils

import (
	"os"

	"github.com/rwcarlsen/goexif/exif"
)

func LocationFromEXIF(photoPath string) (*Location, error) {
	f, err := os.Open(photoPath)
	if err != nil {
		return nil, err
	}
	x, decodeerr := exif.Decode(f)
	if decodeerr != nil {
		return nil, decodeerr
	}

	lat, lon, locerr := x.LatLong()
	if locerr != nil {
		return nil, locerr
	}
	return &Location{Latitude: lat, Longitude: lon}, nil
}
