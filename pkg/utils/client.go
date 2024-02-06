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
	retryableClient.Backoff = retryablehttp.LinearJitterBackoff
	var timeout = time.Duration(timeoutFromConfig()) * time.Second
	retryableClient.RetryWaitMin = timeout / 10
	retryableClient.RetryWaitMax = timeout
	Client = retryableClient.StandardClient()

}
