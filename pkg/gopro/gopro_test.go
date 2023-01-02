package gopro

import (
	"fmt"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestCameraInfoParse(t *testing.T) {

	fileContents := `{
		"info version":"2.0"
		,"firmware version":"H22.01.02.01.00"
		,"wifi mac":"24000000000"
		,"camera type":"HERO11 Black"
		,"camera serial number":"C3471000000000"
		}
		`
	gopro := fstest.MapFS{
		"MISC/version.txt": {
			Data: []byte(fileContents),
		},
		"MISC\\version.txt": {
			Data: []byte(fileContents),
		},
	}

	versionContent, err := gopro.ReadFile(filepath.Join(".", "MISC", fmt.Sprint(Version)))
	require.NoError(t, err)

	gpVersion, err := readInfo(versionContent)
	require.NoError(t, err)

	require.Equal(t, "C3471000000000", gpVersion.CameraSerialNumber)
	require.Equal(t, "HERO11 Black", gpVersion.CameraType)
	require.Equal(t, "H22.01.02.01.00", gpVersion.FirmwareVersion)
	require.Equal(t, "24000000000", gpVersion.WifiMac)
	require.Equal(t, "2.0", gpVersion.InfoVersion)

}
