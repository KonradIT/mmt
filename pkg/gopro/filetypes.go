package gopro

import "regexp"

var FileTypeMatches = map[Type][]FileTypeMatch{
	V2: {
		{
			Regex:    regexp.MustCompile(`GOPR\d+.JPG`),
			Type:     Photo,
			HeroMode: true,
		},
		{
			Regex:    regexp.MustCompile(`GP\d+.JPG`),
			Type:     Photo,
			HeroMode: true,
		},
		{
			Regex:    regexp.MustCompile(`GX\d+.MP4`),
			Type:     Video,
			HeroMode: true,
		},
		{
			Regex:    regexp.MustCompile(`GX\d+.WAV`),
			Type:     Audio,
			HeroMode: true,
		},
		{
			Regex:    regexp.MustCompile(`GH\d+.MP4`),
			Type:     Video,
			HeroMode: true,
		},
		{
			Regex:    regexp.MustCompile(`GG\d+.MP4`), // Live Bursts...
			Type:     Video,
			HeroMode: true,
		},
		{
			Regex:    regexp.MustCompile(`G\d+.JPG`),
			Type:     Multishot,
			HeroMode: true,
		},
		{
			Regex:    regexp.MustCompile(`.GPR`),
			Type:     RawPhoto,
			HeroMode: true,
		},
		// 360 formats, just MAX for now
		{
			Regex:    regexp.MustCompile(`GS\d+.360`),
			Type:     Video,
			HeroMode: false,
		},
		{
			Regex:    regexp.MustCompile(`GS_+\d+.JPG`),
			Type:     Photo,
			HeroMode: false,
		},
		{
			Regex:    regexp.MustCompile(`GP_+\d+.JPG`),
			Type:     Photo,
			HeroMode: true,
		},
		{
			Regex:    regexp.MustCompile(`GPA[A-Z]\d+.JPG`),
			Type:     Multishot,
			HeroMode: true,
		},
		{
			Regex:    regexp.MustCompile(`GSA[A-Z]\d+.JPG`),
			Type:     Multishot,
			HeroMode: false,
		},
	},
	V1: {
		{
			Regex:    regexp.MustCompile(`GOPR\d+.JPG`),
			Type:     Photo,
			HeroMode: true,
		},
		{
			Regex:    regexp.MustCompile(`G\d+.JPG`),
			Type:     Multishot,
			HeroMode: true,
		},
		{
			Regex:    regexp.MustCompile(`GOPR\d+.MP4`),
			Type:     Video,
			HeroMode: true,
		},
		{
			Regex:    regexp.MustCompile(`GP\d+.MP4`),
			Type:     ChapteredVideo,
			HeroMode: true,
		},
		{
			Regex:    regexp.MustCompile(`.GPR`),
			Type:     RawPhoto,
			HeroMode: true,
		},
	},
}
