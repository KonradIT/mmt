package utils

import (
	"net/http"
	"time"

	"github.com/spf13/viper"
)

func timeoutFromConfig() int {
	key := "network_timeout"
	viper.SetDefault(key, 4)
	return viper.GetInt(key)
}

var Client = &http.Client{
	Timeout: time.Duration(timeoutFromConfig()) * time.Second,
}
