package gopro

import (
	"github.com/spf13/viper"
)

const parent = "gopro"

func gpsMinAccuracyFromConfig() uint {
	key := parent + ".gps_accuracy"
	viper.SetDefault(key, 500)

	return viper.GetUint(key)
}

func gpsMaxAltitudeFromConfig() float64 {
	key := parent + ".gps_max_altitude"
	viper.SetDefault(key, 8000)

	return float64(viper.GetUint(key))
}

func gpsCountryCodesFromConfig() []string {
	key := parent + ".gps_country_codes"
	viper.SetDefault(key, []string{}) // 3d lock, 2d lock

	return viper.GetStringSlice(key)
}

func gpsMaxCountryCodesFromConfig() int {
	key := parent + ".gps_max_country_codes"
	viper.SetDefault(key, 5)

	return viper.GetInt(key)
}
