package gopro

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/erdaltsksn/cui"
	"github.com/fatih/color"
	"github.com/k3a/html2text"
	"github.com/konradit/mmt/pkg/utils"
)

var FirmwareCatalogRemoteURL = "https://firmware-api.gopro.com/v2/firmware/catalog"

func UpdateCamera(sdcard string) error {
	req, err := http.NewRequest("GET", FirmwareCatalogRemoteURL, nil)
	if err != nil {
		return err
	}
	resp, err := utils.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	response := &FirmwareCatalog{}
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return err
	}

	versionContent, err := os.ReadFile(filepath.Join(sdcard, "MISC", fmt.Sprint(Version)))
	if err != nil {
		return err
	}

	gpVersion, err := readInfo(versionContent)
	if err != nil {
		return err
	}

	cameraID := fmt.Sprintf("%s.%s", strings.Split(gpVersion.FirmwareVersion, ".")[0], strings.Split(gpVersion.FirmwareVersion, ".")[1])

	for _, camera := range response.Cameras {
		if camera.ModelString != cameraID {
			continue
		}
		cameraVersion := strings.Replace(gpVersion.FirmwareVersion, cameraID+".", "", 1)

		if cameraVersion != camera.Version {
			color.Cyan("New update available!")
			color.Cyan("🎥 Firmware: [%s]", cameraVersion)
			color.Cyan("☁️ Firmware: [%s]", camera.Version)
			color.Yellow(">> Firmware release date: %s", camera.ReleaseDate)
			color.Yellow(html2text.HTML2Text(camera.ReleaseHTML))

			err = utils.DownloadFile(filepath.Join(sdcard, "UPDATE.zip"), camera.URL, nil, nil)
			if err != nil {
				return err
			}
			color.Cyan("Unzipping...")
			err = utils.Unzip(filepath.Join(sdcard, "UPDATE.zip"), filepath.Join(sdcard, "UPDATE"))
			if err != nil {
				return err
			}
			err = os.Remove(filepath.Join(sdcard, "UPDATE.zip"))
			if err != nil {
				return err
			}
			color.Cyan("Firmware extracted to SD card!")
			color.Cyan("Now eject the SD card and insert it into your camera")
			color.Cyan("then turn your camera on and wait for it to update")
		} else {
			cui.Warning("Firmware version is up to date.")
		}
	}

	return nil
}
