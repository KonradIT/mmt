package gopro

import (
	"context"

	"github.com/konradit/mmt/pkg/camera"
	mErrors "github.com/konradit/mmt/pkg/errors"
	"github.com/shirou/gopsutil/disk"
)

// Detect implements camera.Camera by checking for GoPro SD cards and
// network-connected cameras.
func (Entrypoint) Detect() (string, camera.ConnectionType, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return "", "", err
	}
	for _, partition := range partitions {
		if (Entrypoint{}).GuessFromPath(partition.Mountpoint) {
			return partition.Mountpoint, camera.SDCard, nil
		}
	}

	ctx := context.Background()
	networkDevices, err := GetGoProNetworkAddresses(ctx)
	if err != nil {
		return "", "", err
	}

	if len(networkDevices) > 0 {
		return networkDevices[0].IP, camera.Connect, nil
	}

	return "", "", mErrors.ErrNoCameraDetected
}
