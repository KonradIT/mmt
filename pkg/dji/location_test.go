package dji

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/posener/gitfs"
	"github.com/posener/gitfs/fsutil"
	"github.com/stretchr/testify/require"
)

func TestParseSRT(t *testing.T) {
	ctx := context.Background()

	drones := map[string]int{
		"air2s.srt":          290,
		"MAVIC3.srt":         290,
		"mavic_2pro_new.SRT": 290,
		"mavic_pro.SRT":      160,
		"Mini_SE.SRT":        1470,
		"mavic_mini.SRT":     145,
		"p4_rtk.SRT":         250,
	}
	fs, err := gitfs.New(ctx, "github.com/JuanIrache/DJI_SRT_Parser/samples",
		gitfs.OptGlob("*.srt", "*.SRT"))
	require.NoError(t, err)

	walk := fsutil.Walk(fs, "")
	for walk.Step() {
		maxSize, is := drones[walk.Path()]
		if !strings.Contains(walk.Path(), "empty") && is {
			fmt.Printf("\tTesting [%s]", walk.Path())

			remoteFile, err := fs.Open(walk.Path())
			require.NoError(t, err)
			localFile, err := ioutil.TempFile(".", walk.Path())
			require.NoError(t, err)
			defer os.Remove(localFile.Name())

			buf := make([]byte, maxSize) // roughly 290 bytes per SRT entry

			_, err = remoteFile.Read(buf)
			require.NoError(t, err)
			require.NotEmpty(t, buf)

			_, err = localFile.Write(buf)
			require.NoError(t, err)
			err = localFile.Close()
			require.NoError(t, err)

			returned, err := fromSRT(localFile.Name())
			require.NoError(t, err)
			require.NotZero(t, returned.Latitude)
			require.NotZero(t, returned.Longitude)
			require.NotEqual(t, returned.Latitude, returned.Longitude)
			fmt.Printf("\n\treturned: %f %f\n", returned.Latitude, returned.Longitude)
		}
	}
}
