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
		input, err := cmd.Flags().GetString("input")
		if err != nil {
			cui.Error("Something went wrong", err)
		}

		camera, err := cmd.Flags().GetString("camera")
		if err != nil {
			cui.Error("Something went wrong", err)
		}
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
			break
		case utils.Insta360:
			err = insta360.UpdateCamera(input)
			if err != nil {
				cui.Error("Something went wrong", err)
			}
			break
		}

	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringP("input", "i", "", "Input directory for root sd card, eg: E:\\")
	updateCmd.Flags().StringP("camera", "c", "", "Camera type")

}
