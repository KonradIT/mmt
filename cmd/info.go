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
				var gpStatus = gopro.CameraStatus{}
				gpStatus, err = gopro.GetInfo(input)
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
				fmt.Printf("White Balance : %t\n", whiteBal != 0)
				if currentMode == 0 {
					var isoText = ""
					if isoMode == 1 {
						isoText = "Lock"
					} else {
						isoText = "Max"
					}
					fmt.Printf("ISO Mode : %s\n", isoText)
				} else {
					fmt.Printf("ISO Min : %d\n", getISO(isoMin))
				}
				fmt.Printf("ISO Limit : %d\n", getISO(isoLimit))
				fmt.Printf("Protune : %t\n", proTune != 0)
			}
		}
	},
}

func getISO(in int) int {
	switch in {
	case 0:
		return 800
	case 1:
		return 400
	case 2:
		return 200
	case 3:
		return 100
	default:
		return 0
	}
}

func init() {
	rootCmd.AddCommand(infoCmd)
	infoCmd.Flags().StringP("input", "i", "", "Input IP Address")
	infoCmd.Flags().StringP("camera", "c", "", "Camera type")
}
