package gopro

import (
	"fmt"

	"github.com/spf13/viper"
)

const parent = "gopro"

func gpsLockTypesFromConfig() []int {
	key := fmt.Sprintf("%s.gps_fix", parent)
	viper.SetDefault(key, []int{2, 3}) // 3d lock, 2d lock
	return viper.GetIntSlice(key)
}

func gpsMinAccuracyFromConfig() uint16 {
	key := fmt.Sprintf("%s.gps_accuracy", parent)
	viper.SetDefault(key, 500)
	return uint16(viper.GetUint(key))
}
