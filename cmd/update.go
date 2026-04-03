package cmd

import (
	"fmt"

	"github.com/erdaltsksn/cui"
	"github.com/konradit/mmt/pkg/camera"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update camera firmware",
	Run: func(cmd *cobra.Command, _ []string) {
		input := getFlagString(cmd, "input", "")
		cameraType := getFlagString(cmd, "camera", "")
		cam, err := camera.Get(cameraType)
		if err != nil {
			cui.Error("Something went wrong", err)
		}
		updater, ok := cam.(camera.Updater)
		if !ok {
			cui.Error(fmt.Sprintf("camera %q does not support firmware updates", cameraType))
			return
		}
		model := getFlagString(cmd, "model", "")
		err = updater.UpdateFirmware(input, camera.UpdateOptions{Model: model})
		if err != nil {
			cui.Error("Something went wrong", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringP("input", "i", "", "Input directory for root sd card, eg: E:\\")
	updateCmd.Flags().StringP("camera", "c", "", "Camera type")
	updateCmd.Flags().StringP("model", "m", "", "Model type (for insta360): oner, onex, onex2")
}
