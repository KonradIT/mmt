package insta360

import (
	"github.com/konradit/mmt/pkg/camera"
	mErrors "github.com/konradit/mmt/pkg/errors"
	"github.com/shirou/gopsutil/disk"
)

// Detect scans partitions for Insta360 storage.
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

	return "", "", mErrors.ErrNoCameraDetected
}
