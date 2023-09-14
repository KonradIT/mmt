package cmd

import (
	"fmt"

	"github.com/erdaltsksn/cui"
	"github.com/konradit/mmt/pkg/gopro"
	"github.com/konradit/mmt/pkg/utils"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use: "info",
	Run: func(cmd *cobra.Command, args []string) {
		input := getFlagString(cmd, "input")
		camera := getFlagString(cmd, "camera")
		customCameraOpts := make(map[string]interface{})
		if useGoPro, err := cmd.Flags().GetBool("use_gopro"); err == nil && useGoPro {
			detectedGoPro, connectionType, err := gopro.Detect()
			if err != nil {
				cui.Error(err.Error())
			}
			input = detectedGoPro
			customCameraOpts["connection"] = string(connectionType)
			camera = "gopro"
		}
		if camera != "" && input != "" {
			c, err := utils.CameraGet(camera)
			if err != nil {
				cui.Error("Something went wrong", err)
			}
			switch c {
			case utils.GoPro:
				if customCameraOpts["connection"] == "" {
					connection := getFlagString(cmd, "connection")
					if connection == "" {
						connection = "sd_card"
					}
					customCameraOpts["connection"] = connection
				}
				switch customCameraOpts["connection"] {
				case "connect":
					printGpStatus(input)
				default:
					gopro.GetFileInfo(input)
				}
			}
		}
	},
}

func printGpStatus(input string) {
	var gpStatus = gopro.CameraStatus{}
	gpStatus, _ = gopro.GetInfo(input)
	fmt.Printf("SSID : %s\n", gpStatus.Status.WiFiSSID)
	currentMode := gpStatus.Status.CurrentMode
	var modeName = "Video"
	var whiteBal = gpStatus.Settings.Num11
	var isoMode = gpStatus.Settings.Num74
	var isoLimit = gpStatus.Settings.Num13
	var isoMin = 0
	var proTune = gpStatus.Settings.Num10
	switch currentMode {
	case 1:
		modeName = "Photo"
		whiteBal = gpStatus.Settings.Num22
		isoMin = gpStatus.Settings.Num75
		isoLimit = gpStatus.Settings.Num24
		proTune = gpStatus.Settings.Num21
	case 2:
		modeName = "MultiShot"
		whiteBal = gpStatus.Settings.Num35
		isoMin = gpStatus.Settings.Num76
		isoLimit = gpStatus.Settings.Num37
		proTune = gpStatus.Settings.Num34
	}
	fmt.Printf("Mode : %s\n", modeName)
	fmt.Printf("White Balance : %s\n", gopro.GetWhiteBalance(whiteBal))
	if currentMode == 0 {
		fmt.Printf("Resolution : %s\n", gopro.GetVidRes(gpStatus.Settings.VideoResolutions))
		var isoText = ""
		if isoMode == 1 {
			isoText = "Lock"
		} else {
			isoText = "Max"
		}
		fmt.Printf("ISO Mode : %s\n", isoText)
	} else {
		fmt.Printf("ISO Min : %d\n", gopro.GetISO(isoMin))
	}
	fmt.Printf("ISO Limit : %d\n", gopro.GetISO(isoLimit))
	fmt.Printf("Protune : %t\n", proTune != 0)
}

func init() {
	rootCmd.AddCommand(infoCmd)
	infoCmd.Flags().StringP("input", "i", "", "Input IP Address")
	infoCmd.Flags().StringP("camera", "c", "", "Camera type")
	infoCmd.Flags().StringP("connection", "x", "", "Connexion type: `sd_card`, `connect` (GoPro-specific)")
}
