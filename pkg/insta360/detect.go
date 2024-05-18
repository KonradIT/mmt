package insta360

import (
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
		if utils.CameraGuess(partition.Mountpoint) == utils.Insta360.ToString() {
			return partition.Mountpoint, utils.SDCard, nil
		}
	}
	return "", "", mErrors.ErrNoCameraDetected
}
