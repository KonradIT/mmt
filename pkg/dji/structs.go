package dji

import "regexp"

type FileType string

const (
	Video    FileType = "video"
	Photo    FileType = "photo"
	Subtitle FileType = "srt"
	RawPhoto FileType = "dng"
)

type FileTypeMatch struct {
	Regex *regexp.Regexp
	Type  FileType
}
