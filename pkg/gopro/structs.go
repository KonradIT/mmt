package gopro

import (
	"regexp"
	"time"
)

type GoProVersion struct {
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

type FileOfInterest string

const (
	GetStarted FileOfInterest = "Get_started_with_GoPro.url"
	Version    FileOfInterest = "version.txt"
)

type SortOptions struct {
	ByDays             bool
	ByLocation         bool
	SkipAuxiliaryFiles bool
	AddHiLightTags     bool
	ByCamera           bool
	DateFormat         string
	BufferSize         int
	Prefix             string
	DateRange          []time.Time
}

type FileType string

const (
	Video              FileType = "video"
	Photo              FileType = "photo"
	PowerPano          FileType = "powerpano"
	ChapteredVideo     FileType = "chapteredvideo"
	Multishot          FileType = "multishot"
	LowResolutionVideo FileType = "lrv"
	Thumbnail          FileType = "thm"
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
