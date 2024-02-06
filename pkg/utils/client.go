package utils

import (
	"net/http"
	"time"

	"github.com/spf13/viper"
	"github.com/hashicorp/go-retryablehttp"
)

func timeoutFromConfig() int {
	key := "network_timeout"
	viper.SetDefault(key, 4)
	return viper.GetInt(key)
}

var Client *http.Client

func init () {
	var retryableClient = retryablehttp.NewClient()
	retryableClient.Logger = nil
	Client = retryableClient.StandardClient()

}
