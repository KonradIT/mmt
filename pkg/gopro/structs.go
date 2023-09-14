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
	Audio              FileType = "audio"
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

type MediaList struct {
	ID    string `json:"id"`
	Media []struct {
		D  string `json:"d"`
		Fs []struct {
			N    string        `json:"n"`
			Cre  int64         `json:"cre,string"`
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
//CameraStatus current status and settings of the camera
type CameraStatus struct {
	Status   CurrentStatus   `json:"status"`
	Settings CurrentSettings `json:"settings"`
}

// CurrentStatus ...
type CurrentStatus struct {
	// 1 - Internal Battery is available:
	// 	0 = No Battery
	// 	1 = Battery is available
	InternalBattery int `json:"1"`

	// 2 - Internal Battery Level:
	// 	4 = Charging
	// 	3 = Full
	// 	2 = Halfway
	// 	1 = Low
	// 	0 = Empty
	InternalBatteryLevel int `json:"2"`

	Num3  int `json:"3"`
	Num4  int `json:"4"`
	Num6  int `json:"6"`
	Num8  int `json:"8"`
	Num9  int `json:"9"`
	Num10 int `json:"10"`
	Num11 int `json:"11"`

	// 13 - Current Recording Video Duration
	CurrentRecordingVideoDuration int `json:"13"`

	Num14 int    `json:"14"`
	Num15 int    `json:"15"`
	Num16 int    `json:"16"`
	Num17 int    `json:"17"`
	Num19 int    `json:"19"`
	Num20 int    `json:"20"`
	Num21 int    `json:"21"`
	Num22 int    `json:"22"`
	Num23 int    `json:"23"`
	Num24 int    `json:"24"`
	Num26 int    `json:"26"`
	Num27 int    `json:"27"`
	Num28 int    `json:"28"`
	Num29 string `json:"29"`

	// 30 - WiFi SSID
	WiFiSSID string `json:"30"`

	// 31 - Number of clients connected to the camera
	NumberOfClients int `json:"31"`

	// 32 - Streaming feed status:
	// 	0 = Not Streaming
	// 	1 = Streaming
	StreamingFeed int `json:"32"`

	// 33 - SD card inserted:
	// 	0 = SD card inserted
	// 	2 = No SD Card present
	SDcardInserted int `json:"33"`

	// 34 - Remaining Photos
	RemainingPhotos int `json:"34"`

	// 35 - Remaining Video Time
	RemainingVideoTime int `json:"35"`

	// 36 - Number of Batch Photos taken (Example: TimeLapse batches, burst batches, continouous photo batches...)
	NumberOfBatchPhotos int `json:"36"`

	// 37 - Number of videos shot
	NumberOfVideos int `json:"37"`

	// 38 - Number of ALL photos taken
	Num38 int `json:"38"`

	// 39 - Number of MultiShot pictures taken
	// 39 - Number of ALL videos taken
	// 	8 = Recording/Processing status:
	// 	0 = Not recording/Processing
	// 	1 = Recording/processing
	Num39 int    `json:"39"`
	Num40 string `json:"40"`
	Num41 int    `json:"41"`
	Num42 int    `json:"42"`

	// 43 - Current Mode:
	// 	Video - 0
	// 	Photo - 1
	// 	MultiShot - 2
	CurrentMode int `json:"43"`

	// 44 - Current SubMode
	// 	0 = Video/Single Pic/Burst
	// 	1 = TL Video/Continuous/TimeLapse
	// 	2 = Video+Photo/NightPhoto/NightLapse
	CurrentSubMode int `json:"44"`

	Num45 int `json:"45"`
	Num46 int `json:"46"`
	Num47 int `json:"47"`
	Num48 int `json:"48"`
	Num49 int `json:"49"`

	// 54 - Remaning free space on memorycard in bytes
	RemaningFreeSpace int `json:"54"`

	Num55 int `json:"55"`
	Num56 int `json:"56"`
	Num57 int `json:"57"`
	Num58 int `json:"58"`
	Num59 int `json:"59"`
	Num60 int `json:"60"`
	Num61 int `json:"61"`
	Num62 int `json:"62"`
	Num63 int `json:"63"`
	Num64 int `json:"64"`
	Num65 int `json:"65"`
	Num66 int `json:"66"`
	Num67 int `json:"67"`
	Num68 int `json:"68"`
	Num69 int `json:"69"`

	// 70 - Battery Percentage
	BatteryPercentage int `json:"70"`

	Num71 int `json:"71"`
	Num72 int `json:"72"`
	Num73 int `json:"73"`
	Num74 int `json:"74"`
}

// CurrentSettings ...
type CurrentSettings struct {
	Num1 int `json:"1"`

	// 2 - Video Resolutions
	// 	1 = 4K
	// 	4 = 2.7K: http://10.5.5.9/gp/gpControl/setting/2/4
	// 	6 = 2.7K 4:3: http://10.5.5.9/gp/gpControl/setting/2/6
	// 	7 = 1440p: http://10.5.5.9/gp/gpControl/setting/2/7
	// 	9 = 1080p: http://10.5.5.9/gp/gpControl/setting/2/9
	// 	10 = 960p: http://10.5.5.9/gp/gpControl/setting/2/10
	// 	12 = 720p: http://10.5.5.9/gp/gpControl/setting/2/12
	// 	17 = WVGA: http://10.5.5.9/gp/gpControl/setting/2/17
	VideoResolutions int `json:"2"`

	// 3 - Frame Rate
	// 	0 = 240fps:	http://10.5.5.9/gp/gpControl/setting/3/0
	// 	1 = 120fps:	http://10.5.5.9/gp/gpControl/setting/3/1
	// 	2 = 100fps:	http://10.5.5.9/gp/gpControl/setting/3/2
	// 	3 = 90fps:	http://10.5.5.9/gp/gpControl/setting/3/3
	// 	4 = 80fps:	http://10.5.5.9/gp/gpControl/setting/3/4
	// 	5 = 60fps:	http://10.5.5.9/gp/gpControl/setting/3/5
	// 	6 = 50fps:	http://10.5.5.9/gp/gpControl/setting/3/6
	// 	7 = 48fps:	http://10.5.5.9/gp/gpControl/setting/3/7
	// 	8 = 30fps:	http://10.5.5.9/gp/gpControl/setting/3/8
	// 	9 = 25fps:	http://10.5.5.9/gp/gpControl/setting/3/9
	FrameRate int `json:"3"`

	// 4 - Field of View
	// 	0 = Wide: http://10.5.5.9/gp/gpControl/setting/4/0
	// 	1 = Medium: http://10.5.5.9/gp/gpControl/setting/4/1
	// 	2 = Narrow: http://10.5.5.9/gp/gpControl/setting/4/2
	// 	3 = SuperView: http://10.5.5.9/gp/gpControl/4/3
	// 	4 = Linear: http://10.5.5.9/gp/gpControl/setting/4/4
	FOV int `json:"4"`

	// 5 - Video Timelapse Interval:
	// 	0 = 0.5: http://10.5.5.9/gp/gpControl/setting/5/0
	// 	1 = 1: http://10.5.5.9/gp/gpControl/setting/5/1
	// 	2 = 2: http://10.5.5.9/gp/gpControl/setting/5/2
	// 	3 = 5: http://10.5.5.9/gp/gpControl/setting/5/3
	// 	4 = 10: http://10.5.5.9/gp/gpControl/setting/5/4
	// 	5 = 30: http://10.5.5.9/gp/gpControl/setting/5/5
	// 	6 = 60: http://10.5.5.9/gp/gpControl/setting/5/6
	VideoTimelapseInterval int `json:"5"`

	// 6 - Video Looping Duration:
	// 	0 = Max: http://10.5.5.9/gp/gpControl/setting/6/0
	// 	1 = 5Min: http://10.5.5.9/gp/gpControl/setting/6/1
	// 	2 = 20Min: http://10.5.5.9/gp/gpControl/setting/6/2
	// 	3 = 60Min: http://10.5.5.9/gp/gpControl/setting/6/3
	// 	4 = 120Min: http://10.5.5.9/gp/gpControl/setting/6/4
	VideoLoopingDuration int `json:"6"`

	// 7 - Video+Photo Interval:
	// 	1 = 5: http://10.5.5.9/gp/gpControl/setting/7/1
	// 	2 = 10: http://10.5.5.9/gp/gpControl/setting/7/2
	// 	3 = 30: http://10.5.5.9/gp/gpControl/setting/7/3
	// 	4 = 60Min: http://10.5.5.9/gp/gpControl/setting/7/4
	VideoPhotoInterval int `json:"7"`

	// 8 - Low Light
	// 	0 = OFF: http://10.5.5.9/gp/gpControl/setting/8/0
	// 	1 = ON: http://10.5.5.9/gp/gpControl/setting/8/1
	LowLight int `json:"8"`

	// 9 - Spot Meter:
	// 	0 = off: http://10.5.5.9/gp/gpControl/setting/9/0
	// 	1 = on: http://10.5.5.9/gp/gpControl/setting/9/1
	SpotMeter int `json:"9"`

	// https://github.com/KonradIT/goprowifihack/blob/master/HERO5/HERO5-Commands.md
	Num10 int `json:"10"`
	Num11 int `json:"11"`
	Num12 int `json:"12"`
	Num13 int `json:"13"`
	Num14 int `json:"14"`
	Num15 int `json:"15"`
	Num16 int `json:"16"`
	Num17 int `json:"17"`
	Num18 int `json:"18"`
	Num19 int `json:"19"`
	Num20 int `json:"20"`
	Num21 int `json:"21"`
	Num22 int `json:"22"`
	Num23 int `json:"23"`
	Num24 int `json:"24"`
	Num25 int `json:"25"`
	Num26 int `json:"26"`
	Num27 int `json:"27"`
	Num28 int `json:"28"`
	Num29 int `json:"29"`
	Num30 int `json:"30"`
	Num31 int `json:"31"`
	Num32 int `json:"32"`
	Num33 int `json:"33"`
	Num34 int `json:"34"`
	Num35 int `json:"35"`
	Num36 int `json:"36"`
	Num37 int `json:"37"`
	Num38 int `json:"38"`
	Num39 int `json:"39"`
	Num40 int `json:"40"`
	Num41 int `json:"41"`
	Num42 int `json:"42"`

	// 43 - Primary modes
	//	0 = Video: http://10.5.5.9/gp/gpControl/command/mode?p=0
	//	1 = Photo: http://10.5.5.9/gp/gpControl/command/mode?p=1
	//	2 = MultiShot: http://10.5.5.9/gp/gpControl/command/mode?p=2
	PrimaryMode int `json:"43"`

	// 44 - Secondary modes
	// 	0 = Video (VIDEO): http://10.5.5.9/gp/gpControl/command/sub_mode?mode=0&sub_mode=0
	// 	1 = TimeLapse Video (VIDEO): http://10.5.5.9/gp/gpControl/command/sub_mode?mode=0&sub_mode=1
	// 	2 = Video + Photo (VIDEO): http://10.5.5.9/gp/gpControl/command/sub_mode?mode=0&sub_mode=2
	// 	3 = Looping (VIDEO): http://10.5.5.9/gp/gpControl/command/sub_mode?mode=0&sub_mode=3

	// 	1 = Single (PHOTO): http://10.5.5.9/gp/gpControl/command/sub_mode?mode=1&sub_mode=1
	// 	2 = Night (PHOTO): http://10.5.5.9/gp/gpControl/command/sub_mode?mode=1&sub_mode=2

	// 	0 = Burst (MultiShot): http://10.5.5.9/gp/gpControl/command/sub_mode?mode=2&sub_mode=0
	// 	1 = Timelapse (MultiShot): http://10.5.5.9/gp/gpControl/command/sub_mode?mode=2&sub_mode=1
	// 	2 = NightLapse (MultiShot): http://10.5.5.9/gp/gpControl/command/sub_mode?mode=2&sub_mode=2
	SecondaryModes int `json:"44"`

	Num45 int `json:"45"`
	Num46 int `json:"46"`
	Num47 int `json:"47"`
	Num48 int `json:"48"`
	Num50 int `json:"50"`
	Num51 int `json:"51"`
	Num52 int `json:"52"`
	Num54 int `json:"54"`
	Num57 int `json:"57"`
	Num58 int `json:"58"`
	Num59 int `json:"59"`
	Num60 int `json:"60"`
	Num61 int `json:"61"`
	Num62 int `json:"62"`
	Num63 int `json:"63"`
	Num64 int `json:"64"`
	Num65 int `json:"65"`
	Num66 int `json:"66"`
	Num67 int `json:"67"`
	Num68 int `json:"68"`
	Num69 int `json:"69"`
	Num70 int `json:"70"`
	Num71 int `json:"71"`
	Num72 int `json:"72"`
	Num73 int `json:"73"`
	Num74 int `json:"74"`
	Num75 int `json:"75"`
	Num76 int `json:"76"`
	Num77 int `json:"77"`
	Num78 int `json:"78"`
	Num79 int `json:"79"`
	Num80 int `json:"80"`
	Num81 int `json:"81"`
	Num82 int `json:"82"`
	Num83 int `json:"83"`
	Num84 int `json:"84"`
	Num85 int `json:"85"`
	Num86 int `json:"86"`
	Num87 int `json:"87"`
	Num88 int `json:"88"`
	Num89 int `json:"89"`
	Num91 int `json:"91"`
	Num92 int `json:"92"`
	Num93 int `json:"93"`
	Num94 int `json:"94"`
	Num95 int `json:"95"`
	Num96 int `json:"96"`
	Num97 int `json:"97"`
	Num98 int `json:"98"`
	Num99 int `json:"99"`
}
