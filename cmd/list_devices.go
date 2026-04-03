package cmd

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/konradit/mmt/pkg/camera"
	"github.com/konradit/mmt/pkg/gopro"
	"github.com/shirou/gopsutil/disk"
	"github.com/spf13/cobra"
)

var listDevicesCmd = &cobra.Command{
	Use:   "list",
	Short: "List devices available for importing",
	Run: func(_ *cobra.Command, _ []string) {
		partitions, _ := disk.Partitions(false)

		if len(partitions) >= 1 {
			color.Yellow("📷 Devices:")
		}
		for _, partition := range partitions {
			guessed := camera.Guess(partition.Mountpoint)
			name := ""
			if guessed != nil {
				name = guessed.Name()
			}
			color.Cyan(fmt.Sprintf("\t🎥 %v (%v)\n", partition.Mountpoint, name))
		}

		ctx := context.Background()
		networkDevices, err := gopro.GetGoProNetworkAddresses(ctx)
		if err != nil {
			// Network detection is best-effort; don't fail hard.
			return
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
