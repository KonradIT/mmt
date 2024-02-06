package insta360

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/k3a/html2text"
	"github.com/konradit/mmt/pkg/utils"
)

var FirmwareCatalogRemoteURL = "https://openapi.insta360.com/website/appDownload/getGroupApp?group=%s&X-Language=en-us"

func UpdateCamera(sdcard string, model string) error {
	camera, err := CameraGet("insta360-" + model)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("GET", fmt.Sprintf(FirmwareCatalogRemoteURL, camera.String()), nil)
	if err != nil {
		return err
	}
	resp, err := utils.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	response := &FirmwareDownloadList{}
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return err
	}

	for _, firmware := range response.Data.Apps {
		for _, item := range firmware.Items {
			if item.Platform == "insta360" {
				color.Cyan("☁️ Firmware: [%s]", item.Version)
				color.Yellow(">> Firmware release date: %s", item.UpdateTime)
				color.White(html2text.HTML2Text(item.Description))

				fwURL := item.Channels[0].DownloadURL
				err = utils.DownloadFile(filepath.Join(sdcard, strings.Split(fwURL, "/")[len(strings.Split(fwURL, "/"))-1]), fwURL, nil, nil)
				if err != nil {
					return err
				}
				color.Cyan("Firmware downloaded to SD card!")
				color.Cyan("Now eject the SD card and insert it into your camera")
				color.Cyan("then turn your camera on and wait for it to update")
			}
		}
	}
	return nil
}
