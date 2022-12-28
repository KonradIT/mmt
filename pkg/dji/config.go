package dji

import (
	"fmt"

	"github.com/spf13/viper"
)

const parent = "dji"

func srtFolderFromConfig() string {
	key := fmt.Sprintf("%s.srt", parent)
	viper.SetDefault(key, "")
	return viper.GetString(key)
}
