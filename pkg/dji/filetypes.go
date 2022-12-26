package dji

import "regexp"

var fileTypes = []FileTypeMatch{
	{
		Regex: regexp.MustCompile(`.JPG`),
		Type:  Photo,
	},
	{
		Regex: regexp.MustCompile(`.MP4`),
		Type:  Video,
	},
	{
		Regex: regexp.MustCompile(`.SRT`),
		Type:  Subtitle,
	},
	{
		Regex: regexp.MustCompile(`.DNG`),
		Type:  RawPhoto,
	},
	{
		Regex: regexp.MustCompile(`.html`),
		Type:  PanoramaIndex,
	},
	{
		Regex: regexp.MustCompile(`.AAC`),
		Type:  Audio,
	},
}
