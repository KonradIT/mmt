package insta360

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/konradit/mmt/pkg/utils"
)

/* HTML parser for Insta360 (rather, AMBA-wide) media site
Why:
OSC does not serve all files under the `camera.listFiles` option https://developers.google.com/streetview/open-spherical-camera/reference/camera/listfiles
just files for the current session
http://192.168.42.1/DCIM/Camera01/
http://192.168.42.1/DCIM/fileinfo_list.list
*/

type ConnectionType int

const (
	WI_Fi ConnectionType = iota
	USB_API
)

var magicSeparatorStart = []byte{
	0x44,
	0x43,
	0x49,
	0x4D,
}

func importViaWifi(in, out, dateFormat string, bufferSize int, prefix string, dateRange []string) (*utils.Result, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/DCIM/fileinfo_list.list", in), nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	res := bytes.Split(body, magicSeparatorStart)

	photoEnding := byte(0xDA)
	videoEnding := byte(0x12)

	insv := append([]byte("insv"), videoEnding)
	mp4 := append([]byte("mp4"), videoEnding)
	insp := append([]byte("insp"), photoEnding)

	for _, x := range res {
		if strings.HasPrefix(string(x), "/Camera") {

			if bytes.Contains(x, insv) {
				res2 := bytes.Split(x, []byte{videoEnding})
				fmt.Printf(">>> 360: %s\n", string(res2[0]))
			}
			if bytes.Contains(x, mp4) {
				res2 := bytes.Split(x, []byte{videoEnding})
				fmt.Printf(">>> MP4: %s\n", string(res2[0]))
			}
			if bytes.Contains(x, insp) {
				res2 := bytes.Split(x, []byte{photoEnding})
				fmt.Printf(">>> photo: %s\n", string(res2[0]))
			}
		}
	}
	return nil, nil
}
