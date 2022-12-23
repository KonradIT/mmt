package utils

import (
	"fmt"

	"github.com/codingsince1985/geo-golang"
	"github.com/codingsince1985/geo-golang/openstreetmap"
)

type Location struct {
	Latitude  float64
	Longitude float64
}

func getPrettyAddress(address *geo.Address) string {
	if len(address.City) < 9 && address.State != "" {
		return fmt.Sprintf("%s, %s, %s", address.City, address.State, address.Country)
	}
	return fmt.Sprintf("%s, %s", address.City, address.Country)
}

func ReverseLocation(location Location) (string, error) {
	service := openstreetmap.Geocoder()

	address, err := service.ReverseGeocode(location.Latitude, location.Longitude)
	if err != nil {
		return "", err
	}
	return getPrettyAddress(address), nil
}
