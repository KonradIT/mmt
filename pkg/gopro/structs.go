package gopro

import (
	"regexp"
)

type Info struct {
	InfoVersion        string `json:"info version"`
	FirmwareVersion    string `json:"firmware version"`
	WifiMac            string `json:"wifi mac"`
	CameraType         string `json:"camera type"`
	CameraSerialNumber string `json:"camera serial number"`
}

type Directory string

const (
	DCIM Directory = "DCIM"
	MISC Directory = "MISC"
)

type Type string

const (
	V2 Type = "v2"
	V1 Type = "v1"

	UNKNOWN Type = "unknown"
)

type FileOfInterest string

const (
	GetStarted FileOfInterest = "Get_started_with_GoPro.url"
	Version    FileOfInterest = "version.txt"
)

type FileType string

const (
	Video              FileType = "video"
	Photo              FileType = "photo"
	PowerPano          FileType = "powerpano"
	ChapteredVideo     FileType = "chapteredvideo"
	Multishot          FileType = "multishot"
	LowResolutionVideo FileType = "lrv"
	Thumbnail          FileType = "thm"
	RawPhoto           FileType = "gpr"
)

type FileTypeMatch struct {
	Regex    *regexp.Regexp
	Type     FileType
	HeroMode bool
}

type Camera struct {
	Model                  int      `json:"model"`
	ModelString            string   `json:"model_string"`
	Name                   string   `json:"name"`
	Version                string   `json:"version"`
	URL                    string   `json:"url"`
	ReleaseDate            string   `json:"release_date"`
	Sha1                   string   `json:"sha1"`
	RequiredFreeLocalSpace int      `json:"required_free_local_space"`
	CriticalUpdate         bool     `json:"criticalUpdate"`
	Languages              []string `json:"languages"`
	ReleaseHTML            string   `json:"release_html"`
	LicenseHTML            string   `json:"license_html"`
}

type FirmwareCatalog struct {
	Version string   `json:"version"`
	Magic   string   `json:"magic"`
	Group   string   `json:"group"`
	Cameras []Camera `json:"cameras"`
}

type cameraInfo struct {
	Info struct {
		ModelNumber             int      `json:"model_number"`
		ModelName               string   `json:"model_name"`
		FirmwareVersion         string   `json:"firmware_version"`
		SerialNumber            string   `json:"serial_number"`
		BoardType               string   `json:"board_type"`
		ApMac                   string   `json:"ap_mac"`
		ApSsid                  string   `json:"ap_ssid"`
		ApHasDefaultCredentials string   `json:"ap_has_default_credentials"`
		Capabilities            string   `json:"capabilities"`
		LensCount               string   `json:"lens_count"`
		UpdateRequired          string   `json:"update_required"`
		RequiredNetworks        []string `json:"required_networks"`
		HTTPKeepalive           int      `json:"http_keepalive"`
	} `json:"info"`
}

type goProMediaList struct {
	ID    string `json:"id"`
	Media []struct {
		D  string `json:"d"`
		Fs []struct {
			N    string        `json:"n"`
			Cre  string        `json:"cre"`
			Mod  string        `json:"mod"`
			Glrv int           `json:"glrv,string,omitempty"`
			Ls   string        `json:"ls,omitempty"`
			S    int64         `json:"s,string"`
			G    string        `json:"g,omitempty"`
			B    int           `json:"b,string,omitempty"`
			L    int           `json:"l,string,omitempty"`
			T    string        `json:"t,omitempty"`
			M    []interface{} `json:"m,omitempty"`
			Raw  string        `json:"raw,omitempty"`
		} `json:"fs"`
	} `json:"media"`
}

type ConnectDevice struct {
	IP   string
	Info cameraInfo
}

type goProMediaMetadata struct {
	Cre       string        `json:"cre"`
	S         int64         `json:"s,string"`
	Us        string        `json:"us"`
	Mos       []interface{} `json:"mos"`
	Eis       string        `json:"eis"`
	Pta       string        `json:"pta"`
	Ao        string        `json:"ao"`
	Tr        string        `json:"tr"`
	Mp        string        `json:"mp"`
	Gumi      string        `json:"gumi"`
	Ls        string        `json:"ls"`
	Cl        string        `json:"cl"`
	Hc        string        `json:"hc"`
	Hi        []int         `json:"hi"`
	Dur       int           `json:"dur,string"`
	W         string        `json:"w"`
	H         string        `json:"h"`
	Fps       int           `json:"fps,string"`
	FpsDenom  int           `json:"fps_denom,string"`
	Prog      string        `json:"prog"`
	Subsample string        `json:"subsample"`
}

type goProStatus struct {
        batteryPresent          int `json:"1"`
        batteryLevel            int `json:"2"`
        currentMode             int `json:"43"`
        currentSubMode          int `json:"44"`
        currentRecVidDuration   int `json:"13"`
        multiShotNumber         int `json:"39"`
        clients                 int `json:"31"`
        streaming               int `json:"32"`
        sdCard                  int `json:"33"`
        photosRemaining         string `json:"34,string"`
        videoRemaining          string `json:"35,string"`
        batchNumber             int `json:"36"`
        videoCount              int `json:"37"`
        photoCount              int `json:"38"`
        videoAllCount           int `json:"39"`
        processing              int `json:"8"`
        cardSpace               int `json:"54"`
}

type goProSettings struct {
        subModeVideo            int `json:"68"`
        vidRes                  int `json:"2"`
        frameRes                int `json:"3"`
        fovVid                  int `json:"4"`
        timeLapseInt            int `json:"5"`
        loopVidInt              int `json:"6"`
        interval                int `json:"7"`
        lowLight                bool `json:"8"`
        spotMeter               bool `json:"9"`
        proTune                 bool `json:"10"`
        whiteBalance            int `json:"11"`
        color                   int `json:"12"`
        exposure                int `json:"73"`
        isoMode                 int `json:"74"`
        isoLimet                int `json:"13"`
        sharpness               int `json:"14"`
        evComp                  int `json:"15"`
}

type goProCameraStatus struct {
        Status []goProStatus `json:"status"`
        Settings []goProSettings `json:"settings"`
}
