package utils

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	mErrors "github.com/konradit/mmt/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestFindFolderInPath(t *testing.T) {
	t.Run("Correct path - UNIX", func(t *testing.T) {
		input := "/mnt/hdd2/videos/ElEscorialUAV/San Lorenzo de El Escorial España/DJI Device/21-12-2022"
		output := fmt.Sprintf("%[1]smnt%[1]shdd2%[1]svideos%[1]sElEscorialUAV%[1]sSan Lorenzo de El Escorial España%[1]sDJI Device", string(filepath.Separator))
		result, err := FindFolderInPath(input, "DJI Device")
		require.NoError(t, err)
		require.Equal(t, output, result)
	})
	t.Run("Directory not found", func(t *testing.T) {
		input := "C:\\Users\\konra\\Videos\\Projects\\ElEscorialUAV\\San Lorenzo de El Escorial España\\DJI Device\\21-12-2022"
		_, err := FindFolderInPath(input, "INVALID")
		require.Error(t, err)
		require.ErrorContains(t, err, mErrors.ErrNotFound("INVALID").Error())
	})
	if runtime.GOOS == "windows" {
		t.Run("Correct path - Windows", func(t *testing.T) {
			input := "26-12-2022\\San Lorenzo de El Escorial España\\DJI Device"
			output := "26-12-2022\\San Lorenzo de El Escorial España\\DJI Device"
			result, err := FindFolderInPath(input, "DJI Device")
			require.NoError(t, err)
			require.Equal(t, output, result)
		})
		t.Run("Correct path - Windows", func(t *testing.T) {
			input := "C:\\Users\\konra\\Videos\\Projects\\ElEscorialUAV\\San Lorenzo de El Escorial España\\DJI Device\\21-12-2022"
			output := "C:\\Users\\konra\\Videos\\Projects\\ElEscorialUAV\\San Lorenzo de El Escorial España\\DJI Device"
			result, err := FindFolderInPath(input, "DJI Device")
			require.NoError(t, err)
			require.Equal(t, output, result)
		})
	}
}
