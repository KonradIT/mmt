package insta360

import "regexp"

var fileTypes = []FileTypeMatch{
	{
		Regex:         regexp.MustCompile(`IMG_\d+_\d+_\d\d_\d+.jpg`),
		Type:          Photo,
		SteadyCamMode: false,
		OSCMode:       true,
		ProMode:       false,
	},
	{
		Regex:         regexp.MustCompile(`IMG_\d+_\d+_\d\d_\d+.insp`),
		Type:          Photo,
		SteadyCamMode: false,
		OSCMode:       false,
		ProMode:       false,
	},
	{
		Regex:         regexp.MustCompile(`IMG_\d+_\d+_\d\d_\d+.dng`),
		Type:          RawPhoto,
		SteadyCamMode: false,
		OSCMode:       false,
		ProMode:       false,
	},
	{
		Regex:         regexp.MustCompile(`LRV_\d+_\d+_\d\d_\d+.mp4`),
		Type:          LowResolutionVideo,
		SteadyCamMode: true,
		OSCMode:       false,
		ProMode:       false,
	},
	{
		Regex:         regexp.MustCompile(`PRO_LRV_\d+_\d+_\d\d_\d+.mp4`),
		Type:          LowResolutionVideo,
		SteadyCamMode: true,
		OSCMode:       false,
		ProMode:       true,
	},
	{
		Regex:         regexp.MustCompile(`PRO_VID_\d+_\d+_\d\d_\d+.mp4`),
		Type:          Video,
		SteadyCamMode: true,
		OSCMode:       false,
		ProMode:       true,
	},
	{
		Regex:         regexp.MustCompile(`VID_\d+_\d+_\d\d_\d+.mp4`),
		Type:          Video,
		SteadyCamMode: true,
		OSCMode:       false,
		ProMode:       false,
	},
	{
		Regex:         regexp.MustCompile(`VID_\d+_\d+_\d\d_\d+.insv`),
		Type:          Video,
		SteadyCamMode: false,
		OSCMode:       false,
		ProMode:       false,
	},
	{
		Regex:         regexp.MustCompile(`LRV_\d+_\d+_\d\d_\d+.insv`),
		Type:          LowResolutionVideo,
		SteadyCamMode: false,
		OSCMode:       false,
		ProMode:       false,
	},
}
