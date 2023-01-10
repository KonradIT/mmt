package gopro

/* GoPro Connect - API exposed over USB Ethernet */

import (
        "fmt"
        /*
	mErrors "github.com/konradit/mmt/pkg/errors"
	"github.com/konradit/mmt/pkg/utils"
	"github.com/shirou/gopsutil/disk"
        */
)

func GetInfo(in string) error {
	var gpStatus = &goProCameraStatus{}
	err := caller(in, "gp/gpControl/status", gpStatus)
	if err != nil {
               fmt.Println(err)
               return err
	}
        fmt.Println(gpStatus)
        return nil
}

