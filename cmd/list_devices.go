package cmd

import (
	"fmt"

	"github.com/erdaltsksn/cui"
	"github.com/fatih/color"
	"github.com/konradit/gowpd"
	"github.com/konradit/mmt/pkg/gopro"
	"github.com/spf13/cobra"
)

var listDevicesCmd = &cobra.Command{
	Use:   "list",
	Short: "List devices available for importing",
	Run: func(cmd *cobra.Command, args []string) {
		err := gowpd.Init()
		defer gowpd.Destroy()
		if err != nil {
			cui.Error(err.Error())
		}
		n := gowpd.GetDeviceCount()
		if n >= 1 {
			color.Yellow("📷 Devices:")
		}
		for i := 0; i < n; i++ {
			color.Cyan(fmt.Sprintf("\t🎥 %v - %v (%v) [%v]\n", i, gowpd.GetDeviceName(i), gowpd.GetDeviceDescription(i), gowpd.GetDeviceManufacturer(i)))
		}

		networkDevices, err := gopro.GetGoProNetworkAddresses()
		if err != nil {
			cui.Error(err.Error())
		}

		if len(networkDevices) >= 1 {
			color.Yellow("🔌 GoPro cameras via Connect (USB Ethernet):")
		}
		for i, devc := range networkDevices {
			color.White(fmt.Sprintf("\t📹 %d - %s (%s - %s)", i, devc.IP, devc.Info.Info.ModelName, devc.Info.Info.FirmwareVersion))
		}
	},
}

func init() {
	rootCmd.AddCommand(listDevicesCmd)
}
