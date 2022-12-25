package insta360

import (
	"regexp"

	mErrors "github.com/konradit/mmt/pkg/errors"
)

type Metadata struct {
	Model        string
	SerialNumber string
}
type Item struct {
	IsMininumVersion bool   `json:"is_mininum_version"`
	WebsiteVisible   bool   `json:"website_visible"`
	Forced           bool   `json:"forced"`
	Description      string `json:"description"`
	Language         string `json:"language"`
	Version          string `json:"version"`
	Platform         string `json:"platform"`
	AppVisible       bool   `json:"app_visible"`
	ItemID           int    `json:"itemId"`
	UpdateTime       string `json:"update_time"`
	VersionName      string `json:"version_name,omitempty"`
	Channels         []struct {
		Channel     string `json:"channel"`
		DownloadURL string `json:"download_url"`
		ItemID      int    `json:"item_id"`
		OrderIndex  int    `json:"order_index"`
		Visible     bool   `json:"visible"`
	} `json:"channels"`
	ImportantTag bool `json:"important_tag"`
	ID           int  `json:"id"`
	IsTest       bool `json:"is_test"`
	AppID        int  `json:"app_id"`
}
type App struct {
	MainName            string `json:"main_name"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	TitleImage          string `json:"title_image"`
	Language            string `json:"language"`
	ID                  int    `json:"id"`
	LogoImage           string `json:"logo_image"`
	AppID               int    `json:"app_id"`
	Items               []Item `json:"items"`
	Key                 string `json:"key"`
	NameLinkURL         string `json:"name_link_url,omitempty"`
	DescriptionLinkText string `json:"description_link_text,omitempty"`
	DescriptionLinkURL  string `json:"description_link_url,omitempty"`
	NameLinkText        string `json:"name_link_text,omitempty"`
	UnsupportText       string `json:"unsupport_text,omitempty"`
}
type FirmwareDownloadList struct {
	Code int `json:"code"`
	Data struct {
		Apps []App `json:"apps"`
	} `json:"data"`
}
type FileType string

const (
	Video              FileType = "video"
	Photo              FileType = "photo"
	LowResolutionVideo FileType = "lrv"
	RawPhoto           FileType = "dng"
)

type FileTypeMatch struct {
	Regex         *regexp.Regexp
	Type          FileType
	SteadyCamMode bool
	OSCMode       bool
	ProMode       bool
}

type File struct {
	Type           FileType `len:"3"`
	Date           string   `len:"8"`
	ID             int      `len:"6"`
	Part           string   `len:"2"`
	SequenceNumber int      `len:"3"`
}

type Camera string

const (
	OneR  Camera = "insta360-oner"
	OneX  Camera = "insta360-onex"
	OneX2 Camera = "insta360-onex2"
	Go2   Camera = "insta360-go2"
)

func (e Camera) String() string {
	extensions := [...]string{"insta360-oner", "insta360-onex", "insta360-onex2", "insta360-go2"}

	x := string(e)
	for _, v := range extensions {
		if v == x {
			return x
		}
	}

	return ""
}

func CameraGet(s string) (Camera, error) {
	switch s {
	case OneR.String():
		return OneR, nil
	case OneX2.String():
		return OneX2, nil
	case OneX.String():
		return OneX, nil
	case Go2.String():
		return Go2, nil
	}
	return OneX, mErrors.ErrUnsupportedCamera(s)
}
