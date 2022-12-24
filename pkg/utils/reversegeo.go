package utils

import (
	"fmt"

	"github.com/codingsince1985/geo-golang"
	"github.com/codingsince1985/geo-golang/openstreetmap"
	"github.com/spf13/viper"
)

type Location struct {
	Latitude  float64
	Longitude float64
}

func formatFromConfig() int {
	return viper.GetInt("location.format")
}

func fallbackFromConfig() string {
	key := "location.fallback"
	viper.SetDefault(key, "NoLocation")
	return viper.GetString(key)
}

func orderFromConfig() []string {
	key := "location.order"
	viper.SetDefault(key, []string{"date", "location", "device"})
	return viper.GetStringSlice(key)
}

type locationFormat interface {
	format(*geo.Address) string
}

type format1 struct{}

func (format1) format(address *geo.Address) string {
	if len(address.City) < 9 && address.State != "" {
		return fmt.Sprintf("%s, %s, %s", address.City, address.State, address.Country)
	}
	return fmt.Sprintf("%s, %s", address.City, address.Country)
}

type format2 struct{}

func (format2) format(address *geo.Address) string {
	return address.Country
}

func getPrettyAddress(format locationFormat, address *geo.Address) string {
	return format.format(address)
}

func ReverseLocation(location Location) (string, error) {
	service := openstreetmap.Geocoder()

	address, err := service.ReverseGeocode(location.Latitude, location.Longitude)
	if err != nil {
		return "", err
	}

	format := formatFromConfig()
	switch format {
	case 1:
		return getPrettyAddress(format1{}, address), nil
	case 2:
		return getPrettyAddress(format2{}, address), nil
	}
	return getPrettyAddress(format1{}, address), nil
}
