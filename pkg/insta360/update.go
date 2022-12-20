package insta360

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/k3a/html2text"
	"github.com/konradit/mmt/pkg/utils"
)

var FirmwareCatalogRemoteURL = "https://service.insta360.com/app-service/app/appDownload/getGroupApp?group=%s&X-Language=en-us"

func UpdateCamera(sdcard string, model string) error {
	client := &http.Client{}

	camera, err := CameraGet("insta360-" + model)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("GET", fmt.Sprintf(FirmwareCatalogRemoteURL, camera.String()), nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var response = &FirmwareDownloadList{}
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return err
	}

	for _, firmware := range response.Data.Apps {
		if firmware.Name != "Camera firmware" {
			continue
		}
		item := firmware.Items[len(firmware.Items)-1]
		color.Cyan("Insta360 Firmware:\n☁️ Version: [%s]", item.Version)
		updateTime := time.UnixMilli(item.UpdateTime)

		color.Yellow("📅 Release date: [%s]", updateTime.Format(time.RFC822))
		color.White(html2text.HTML2Text(item.Description))

		fwURL := item.Channels[0].DownloadURL

		err = utils.DownloadFile(filepath.Join(sdcard, filepath.Base(fwURL)), fwURL)
		if err != nil {
			return err
		}
		color.Cyan("Firmware downloaded to SD card!")
		color.Cyan("Now eject the SD card and insert it into your camera")
		color.Cyan("then turn your camera on and wait for it to update")
		break
	}
	return nil
}
