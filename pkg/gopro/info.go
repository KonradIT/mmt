package gopro

/* GoPro Connect - API exposed over USB Ethernet */

import (
        /*
	mErrors "github.com/konradit/mmt/pkg/errors"
	"github.com/konradit/mmt/pkg/utils"
	"github.com/shirou/gopsutil/disk"
        */
)

func GetInfo(in string) (CameraStatus, error) {
	var gpStatus = &CameraStatus{}
	err := caller(in, "gp/gpControl/status", gpStatus)
	if err != nil {
               return *gpStatus, err
	}
        return *gpStatus, nil
}

