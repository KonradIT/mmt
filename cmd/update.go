package cmd

import (
	"github.com/erdaltsksn/cui"
	"github.com/konradit/mmt/pkg/gopro"
	"github.com/konradit/mmt/pkg/insta360"
	"github.com/konradit/mmt/pkg/utils"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update camera firmware",
	Run: func(cmd *cobra.Command, args []string) {
		input := getFlagString(cmd, "input", "")
		camera := getFlagString(cmd, "camera", "")
		c, err := utils.CameraGet(camera)
		if err != nil {
			cui.Error("Something went wrong", err)
		}
		switch c {
		case utils.GoPro:
			err = gopro.UpdateCamera(input)
			if err != nil {
				cui.Error("Something went wrong", err)
			}
		case utils.Insta360:
			model := getFlagString(cmd, "model", "")
			err = insta360.UpdateCamera(input, model)
			if err != nil {
				cui.Error("Something went wrong", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringP("input", "i", "", "Input directory for root sd card, eg: E:\\")
	updateCmd.Flags().StringP("camera", "c", "", "Camera type")
	updateCmd.Flags().StringP("model", "m", "", "Model type (for insta360): oner, onex, onex2")
}
