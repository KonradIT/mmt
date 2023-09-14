package gopro

/* GoPro Connect - API exposed over USB Ethernet */

import (
	"fmt"

	"github.com/konradit/mmt/pkg/utils"
)

func GetInfo(in string) (CameraStatus, error) {
	var gpStatus = &CameraStatus{}
	err := caller(in, "gp/gpControl/status", gpStatus)
	if err != nil {
		return *gpStatus, err
	}
	return *gpStatus, nil
}

func GetISO(in int) int {
	switch in {
	case 0:
		return 800
	case 1:
		return 400
	case 2:
		return 200
	case 3:
		return 100
	default:
		return 0
	}
}

func GetVidRes(in int) string {
	switch in {
	case 1:
		return "4K"
	case 2:
		return "4K SuperView"
	case 4:
		return "2.7K"
	case 5:
		return "2.7K SuperView"
	case 6:
		return "2.7K 4:3"
	case 7:
		return "1440"
	case 8:
		return "1080 SuperView"
	case 9:
		return "1080"
	case 10:
		return "960"
	case 11:
		return "720 SuperView"
	case 12:
		return "720"
	case 13:
		return "WVGA"
	default:
		return ""
	}
}

func GetWhiteBalance(in int) string {
	switch in {
	case 0:
		return "Auto"
	case 1:
		return "3000K"
	case 5:
		return "4000K"
	case 6:
		return "4800K"
	case 2:
		return "5500K"
	case 7:
		return "6000K"
	case 3:
		return "6500K"
	case 4:
		return "Native"
	default:
		return ""
	}
}

func GetFileInfo(in string) (*utils.Result, error) {
	var result utils.Result
	returned, err := fromMP4(in)
	if returned != nil {
		fmt.Printf("\n\treturned: %f %f\n", returned.Latitude, returned.Longitude)
	}
	return &result, err
}
