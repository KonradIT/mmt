package camera

import (
	"os"
	"path/filepath"

	"github.com/konradit/mmt/pkg/utils"
	"github.com/spf13/viper"
)

// LocationProvider extracts GPS coordinates from a media file.
type LocationProvider interface {
	GetLocation(path string) (*utils.Location, error)
}

func orderFromConfig() []string {
	key := "location.order"
	viper.SetDefault(key, []string{"date", "location", "camera"})

	return viper.GetStringSlice(key)
}

func fallbackFromConfig() string {
	key := "location.fallback"
	viper.SetDefault(key, "NoLocation")

	return viper.GetString(key)
}

// GetOrder determines the destination folder based on sort options, location,
// date, and device name.
func GetOrder(sortoptions SortOptions, getLocation LocationProvider, osPathname, out, mediaDate, deviceName string) string {
	order := orderFromConfig()
	dayFolder := out

	for _, item := range order {
		switch item {
		case "date":
			dayFolder = filepath.Join(dayFolder, mediaDate)
		case "camera":
			if sortoptions.ByCamera {
				dayFolder = filepath.Join(dayFolder, deviceName)
			}
		case "location":
			if getLocation == nil || !sortoptions.ByLocation {
				continue
			}

			location := fallbackFromConfig()

			locationFromFile, locerr := getLocation.GetLocation(osPathname)
			if locerr == nil {
				reverseLocation, reverseerr := utils.ReverseLocation(*locationFromFile)
				if reverseerr == nil && reverseLocation != "" && reverseLocation != " " {
					location = reverseLocation
				}
			}

			dayFolder = filepath.Join(dayFolder, location)
		}
	}

	_ = os.MkdirAll(dayFolder, 0o755)

	return dayFolder
}
