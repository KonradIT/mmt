package insta360

type Insta360Metadata struct {
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
