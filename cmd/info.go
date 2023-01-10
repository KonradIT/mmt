package cmd

import (
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
				err = gopro.GetInfo(input)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
	infoCmd.Flags().StringP("input", "i", "", "Input IP Address")
	infoCmd.Flags().StringP("camera", "c", "", "Camera type")
}
