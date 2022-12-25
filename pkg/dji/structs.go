package dji

import "regexp"

type FileType string

const (
	Video         FileType = "video"
	Photo         FileType = "photo"
	Subtitle      FileType = "srt"
	RawPhoto      FileType = "dng"
	PanoramaIndex FileType = "panoramaindex"
	Audio         FileType = "audio"
)

type FileTypeMatch struct {
	Regex *regexp.Regexp
	Type  FileType
}

var DeviceNames = map[string]string{
	"FC1102":  "Spark",
	"FC220":   "Mavic Pro",
	"FC300C":  "Phantom 3",
	"FC300S":  "Phantom 3 Pro",
	"FC300SE": "Phantom 3 Pro",
	"FC300X":  "Phantom 3 Pro",
	"FC300XW": "Phantom 3 Adv",
	"FC3170":  "Mavic Air 2",
	"FC330":   "Phantom 4",
	"FC3411":  "Air 2S",
	"FC350":   "X3",
	"FC550":   "X5",
	"FC6310":  "Phantom 4 Pro",
	"FC6510":  "X4S",
	"FC6520":  "X5S",
	"FC6540":  "X7",
	"FC7203":  "Mavic Mini",
	"HG310":   "OSMO",
	"OT110":   "Osmo Pocket",
	"L1D-20":  "Mavic 2 Pro",
	"L2D-20c": "Mavic 3",
	"FC7303":  "Mini 2",
	"FC3582":  "Mini 3 Pro",
}
