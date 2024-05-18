package gopro

import (
	"context"

	mErrors "github.com/konradit/mmt/pkg/errors"
	"github.com/konradit/mmt/pkg/utils"
	"github.com/shirou/gopsutil/disk"
)

func Detect() (string, utils.ConnectionType, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return "", "", err
	}
	for _, partition := range partitions {
		if utils.CameraGuess(partition.Mountpoint) == utils.GoPro.ToString() {
			return partition.Mountpoint, utils.SDCard, nil
		}
	}

	ctx := context.Background()
	networkDevices, err := GetGoProNetworkAddresses(ctx)
	if err != nil {
		return "", "", err
	}

	if len(networkDevices) > 0 {
		return networkDevices[0].IP, utils.Connect, nil
	}

	return "", "", mErrors.ErrNoCameraDetected
}
