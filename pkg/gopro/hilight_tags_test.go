package gopro

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

var importanceNames = []string{"Marked 1", "Lit AF", "Important"}

func TestLengthTooShort(t *testing.T) {
	payload := `{"cre":"1672490791","s":"56337675","us":"0","mos":[],"eis":"0","pta":"1","ao":"stereo","tr":"0","mp":"0","gumi":"edc695738198c0e25b5d439c036dbfd1","ls":"4149813","cl":"0","hc":"3","hi":[1360,3400,7400],"dur":"10","w":"1920","h":"1080","fps":"3600","fps_denom":"90000","prog":"1","subsample":"0"}`
	gpFileInfo := goProMediaMetadata{}
	err := json.Unmarshal([]byte(payload), &gpFileInfo)
	require.NoError(t, err)

	importanceName := getImportanceName(gpFileInfo.Hi, gpFileInfo.Dur, importanceNames)
	require.Empty(t, importanceName)
}

func TestAllInBounds(t *testing.T) {
	payload := `{"cre":"1672494840","s":"116068980","us":"0","mos":[],"eis":"0","pta":"1","ao":"stereo","tr":"0","mp":"0","gumi":"d3d262f1e7c4772ef00281333e482074","ls":"9323550","cl":"0","hc":"2","hi":[18880,20440],"dur":"24","w":"1920","h":"1080","fps":"3600","fps_denom":"90000","prog":"1","subsample":"0"}`
	gpFileInfo := goProMediaMetadata{}
	err := json.Unmarshal([]byte(payload), &gpFileInfo)
	require.NoError(t, err)

	importanceName := getImportanceName(gpFileInfo.Hi, gpFileInfo.Dur, importanceNames)
	require.Equal(t, importanceName, "Lit AF")
}

func TestMarkerOverflow(t *testing.T) {
	payload := `{"cre":"1672496212","s":"127341393","us":"0","mos":[],"eis":"0","pta":"1","ao":"stereo","tr":"0","mp":"0","gumi":"2de66cc9ef62b52c326b20b7ad15c098","ls":"11858166","cl":"0","hc":"4","hi":[22600,24320,25920,27640],"dur":"30","w":"1920","h":"1080","fps":"3600","fps_denom":"90000","prog":"1","subsample":"0"}`
	gpFileInfo := goProMediaMetadata{}
	err := json.Unmarshal([]byte(payload), &gpFileInfo)
	require.NoError(t, err)

	importanceName := getImportanceName(gpFileInfo.Hi, gpFileInfo.Dur, importanceNames)
	require.Equal(t, importanceName, "Important")
}

func TestOneMarker(t *testing.T) {
	payload := `{"cre":"1672496247","s":"171050196","us":"0","mos":[],"eis":"0","pta":"1","ao":"stereo","tr":"0","mp":"0","gumi":"1ff6fdfbc8c3dbe6ca064bb00283e7c0","ls":"17821498","cl":"0","hc":"1","hi":[42720],"dur":"46","w":"1920","h":"1080","fps":"3600","fps_denom":"90000","prog":"1","subsample":"0"}`
	gpFileInfo := goProMediaMetadata{}
	err := json.Unmarshal([]byte(payload), &gpFileInfo)
	require.NoError(t, err)

	importanceName := getImportanceName(gpFileInfo.Hi, gpFileInfo.Dur, importanceNames)
	require.Equal(t, importanceName, "Marked 1")
}

func TestNoneAtAll(t *testing.T) {
	payload := `{"cre":"1672497199","s":"301389825","us":"0","mos":[],"eis":"0","pta":"1","ao":"stereo","tr":"0","mp":"0","gumi":"012912787a0bdffaab89f940285de16c","ls":"31581398","cl":"0","hc":"0","hi":[],"dur":"80","w":"1920","h":"1080","fps":"3600","fps_denom":"90000","prog":"1","subsample":"0"}`
	gpFileInfo := goProMediaMetadata{}
	err := json.Unmarshal([]byte(payload), &gpFileInfo)
	require.NoError(t, err)

	importanceName := getImportanceName(gpFileInfo.Hi, gpFileInfo.Dur, importanceNames)
	require.Empty(t, importanceName)
}
