package utils

import (
	"os"
	"path/filepath"
	"time"
)

type locationUtil interface {
	GetLocation(path string) (*Location, error)
}

type SortOptions struct {
	ByLocation         bool
	SkipAuxiliaryFiles bool
	ByCamera           bool
	DateFormat         string
	BufferSize         int
	Prefix             string
	DateRange          []time.Time
}

func GetOrder(sortoptions SortOptions, GetLocation locationUtil, osPathname, out, mediaDate, deviceName string) string {
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
			if GetLocation == nil {
				continue
			}
			location := fallbackFromConfig()
			locationFromFile, locerr := GetLocation.GetLocation(osPathname)
			if locerr == nil {
				reverseLocation, reverseerr := ReverseLocation(*locationFromFile)
				if reverseerr == nil {
					location = reverseLocation
				}
			}
			if sortoptions.ByLocation {
				dayFolder = filepath.Join(dayFolder, location)
			}
		default:
			// Not supported
		}
	}
	if _, err := os.Stat(dayFolder); os.IsNotExist(err) {
		_ = os.MkdirAll(dayFolder, 0755)
	}
	return dayFolder
}
