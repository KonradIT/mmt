package dji

import (
	"github.com/spf13/viper"
)

const parent = "dji"

func srtFolderFromConfig() string {
	key := parent + ".srt"
	viper.SetDefault(key, "")

	return viper.GetString(key)
}
