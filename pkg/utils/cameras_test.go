package utils

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindFolderInPath(t *testing.T) {
	t.Run("Correct path", func(t *testing.T) {
		input := "C:\\Users\\konrad\\Videos\\Projects\\ElEscorialUAV\\San Lorenzo de El Escorial España\\DJI Device\\21-12-2022"
		output := "C:" + string(filepath.Separator) + "Users" + string(filepath.Separator) + "konrad" + string(filepath.Separator) + "Videos" + string(filepath.Separator) + "Projects" + string(filepath.Separator) + "ElEscorialUAV" + string(filepath.Separator) + "San Lorenzo de El Escorial España" + string(filepath.Separator) + "DJI Device"
		result, err := FindFolderInPath(input, "DJI Device")
		require.NoError(t, err)
		require.Equal(t, output, result)
	})
	t.Run("Correct path - UNIX", func(t *testing.T) {
		input := "/mnt/hdd2/videos/ElEscorialUAV/San Lorenzo de El Escorial España/DJI Device/21-12-2022"
		output := string(filepath.Separator) + "mnt" + string(filepath.Separator) + "hdd2" + string(filepath.Separator) + "videos" + string(filepath.Separator) + "ElEscorialUAV" + string(filepath.Separator) + "San Lorenzo de El Escorial España" + string(filepath.Separator) + "DJI Device"
		result, err := FindFolderInPath(input, "DJI Device")
		require.NoError(t, err)
		require.Equal(t, output, result)
	})
	t.Run("Directory not found", func(t *testing.T) {
		input := "C:\\Users\\konra\\Videos\\Projects\\ElEscorialUAV\\San Lorenzo de El Escorial España\\DJI Device\\21-12-2022"
		_, err := FindFolderInPath(input, "INVALID")
		require.Error(t, err)
	})
}
