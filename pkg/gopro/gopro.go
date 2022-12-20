package gopro

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/karrick/godirwalk"
	"github.com/konradit/mmt/pkg/utils"
	"github.com/maja42/goval"
	"github.com/minio/minio/pkg/disk"
	"gopkg.in/djherbis/times.v1"
)

/*
Uses data from:
https://community.gopro.com/t5/en/GoPro-Camera-File-Naming-Convention/ta-p/390220#
*/

var replacer = strings.NewReplacer("dd", "02", "mm", "01", "yyyy", "2006")

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
			Regex:    regexp.MustCompile(`GH\d+.MP4`),
			Type:     Video,
			HeroMode: true,
		},
		{
			Regex:    regexp.MustCompile(`GL\d+.LRV`),
			Type:     LowResolutionVideo,
			HeroMode: true,
		},
		{
			Regex:    regexp.MustCompile(`GH\d+.THM`),
			Type:     Thumbnail,
			HeroMode: true,
		},
		{
			Regex:    regexp.MustCompile(`GX\d+.THM`),
			Type:     Thumbnail,
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
	},
	MAX: {
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
			Regex:    regexp.MustCompile(`GH\d+.MP4`),
			Type:     Video,
			HeroMode: true,
		},
		{
			Regex:    regexp.MustCompile(`GX\d+.MP4`),
			Type:     Video,
			HeroMode: true,
		},
		{
			Regex:    regexp.MustCompile(`GPAA\d+.JPG`),
			Type:     Multishot,
			HeroMode: true,
		},
		{
			Regex:    regexp.MustCompile(`GH\d+.LRV`),
			Type:     LowResolutionVideo,
			HeroMode: true,
		},
		{
			Regex:    regexp.MustCompile(`GH\d+.THM`),
			Type:     Thumbnail,
			HeroMode: true,
		},
		{
			Regex:    regexp.MustCompile(`GS\d+.LRV`),
			Type:     LowResolutionVideo,
			HeroMode: false,
		},
		{
			Regex:    regexp.MustCompile(`GS\d+.THM`),
			Type:     Thumbnail,
			HeroMode: false,
		},
		{
			Regex:    regexp.MustCompile(`GSAA\d+.JPG`),
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
			Regex:    regexp.MustCompile(`GOPR\d+.LRV`),
			Type:     LowResolutionVideo,
			HeroMode: true,
		},
		{
			Regex:    regexp.MustCompile(`GOPR\d+.THM`),
			Type:     Thumbnail,
			HeroMode: true,
		},
		{
			Regex:    regexp.MustCompile(`.GPR`),
			Type:     RawPhoto,
			HeroMode: true,
		},
	},
}

var MediaFolderRegex = regexp.MustCompile(`\d\d\dGOPRO`)

var ffprobe = utils.NewFFprobe(nil)

func Import(in, out, dateFormat string, bufferSize int, prefix string, dateRange []string, cameraOptions map[string]interface{}) (*utils.Result, error) {
	/* Import method using SD card bay or SD card reader */

	dateStart := time.Date(0000, time.Month(1), 1, 0, 0, 0, 0, time.UTC)

	dateEnd := time.Now()

	byCamera := false

	sortByOptions, found := cameraOptions["sort_by"]
	if found {
		for _, sortop := range sortByOptions.([]string) {
			if sortop == "camera" {
				byCamera = true
			}

			if sortop != "camera" && sortop != "days" {
				return nil, errors.New("Unrecognized option for sort_by: " + sortop)
			}
		}
	}
	if len(dateRange) == 1 {
		switch dateRange[0] {
		case "today":
			dateStart = time.Date(dateEnd.Year(), dateEnd.Month(), dateEnd.Day(), 0, 0, 0, 0, dateEnd.Location())
		case "yesterday":
			dateStart = time.Date(dateEnd.Year(), dateEnd.Month(), dateEnd.Day(), 0, 0, 0, 0, dateEnd.Location()).Add(-24 * time.Hour)
		case "week":
			dateStart = time.Date(dateEnd.Year(), dateEnd.Month(), dateEnd.Day(), 0, 0, 0, 0, dateEnd.Location()).Add(-24 * time.Duration((int(dateEnd.Weekday()) - 1)) * time.Hour)
		}
	}

	if len(dateRange) == 2 {
		start, err := time.Parse(replacer.Replace(dateFormat), dateRange[0])
		if err != nil {
			log.Fatal(err.Error())
		}
		if err == nil {
			dateStart = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
		}
		end, err := time.Parse(replacer.Replace(dateFormat), dateRange[1])
		if err != nil {
			log.Fatal(err.Error())
		}
		if err == nil {
			dateEnd = time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())
		}
	}

	skipAux := false
	skipAuxOption, found := cameraOptions["skip_aux"]
	if found {
		skipAux = skipAuxOption.(bool)
	}
	sortOptions := SortOptions{
		SkipAuxiliaryFiles: skipAux,
		AddHiLightTags:     true,
		ByCamera:           byCamera,
		DateFormat:         dateFormat,
		BufferSize:         bufferSize,
		Prefix:             prefix,
		DateRange:          []time.Time{dateStart, dateEnd},
	}

	connectionType, found := cameraOptions["connection"]
	if found {
		switch connectionType.(string) {
		case string(utils.Connect):
			return ImportConnect(in, out, sortOptions)
		case string(utils.SDCard):
			break
		default:
			return nil, errors.New("Unsupported connection")
		}
	}

	gpVersion, err := readInfo(in)
	if err != nil {
		return nil, err
	}
	if prefix == "cameraname" {
		prefix = gpVersion.CameraType
		sortOptions.Prefix = prefix
	}

	di, err := disk.GetInfo(in)
	if err != nil {
		return nil, err
	}
	percentage := (float64(di.Total-di.Free) / float64(di.Total)) * 100

	c := color.New(color.FgCyan)
	y := color.New(color.FgHiBlue)
	color.Cyan("🎥 [%s]:", gpVersion.CameraType)
	c.Printf("\t📹 FW: %s ", gpVersion.FirmwareVersion)
	y.Printf("SN: %s\n", gpVersion.CameraSerialNumber)
	color.Cyan("\t💾 %s/%s (%0.2f%%)\n",
		humanize.Bytes(di.Total-di.Free),
		humanize.Bytes(di.Total),
		percentage,
	)

	root := strings.Split(gpVersion.FirmwareVersion, ".")[0]

	switch root {
	case "HD6", "HD7", "HD8", "HD9", "H21", "H22":
		result := importFromGoProV2(filepath.Join(in, fmt.Sprint(DCIM)), out, sortOptions, gpVersion.CameraType)
		return &result, nil
	case "HD2,", "HD3", "HD4", "HX", "HD5":
		result := importFromGoProV1(filepath.Join(in, fmt.Sprint(DCIM)), out, sortOptions, gpVersion.CameraType)
		return &result, nil
	case "H19":
		result := importFromMAX(filepath.Join(in, fmt.Sprint(DCIM)), out, sortOptions)
		return &result, nil
	default:
		return nil, fmt.Errorf("Camera `%s` is not supported", gpVersion.CameraType)
	}
}

func importFromMAX(root string, output string, sortoptions SortOptions) utils.Result {
	fileTypes := FileTypeMatches[MAX]

	var result utils.Result
	/*
		The idea is to have a result like:

		20-02-2021/MAX/photos/
							normal/
						   			GS_0001.JPG
							powerpano/
						   			GP_0001.JPG
					   /videos
						   /heromode
									/GH012345.MP4
						   /360
								/GS012345.MP4

	*/
	folders, err := ioutil.ReadDir(root)
	if err != nil {
		result.Errors = append(result.Errors, err)
		return result
	}

	for _, f := range folders {
		r := MediaFolderRegex.MatchString(f.Name())

		if !r {
			continue
		}
		color.Green("Looking at %s", f.Name())

		err = godirwalk.Walk(filepath.Join(root, f.Name()), &godirwalk.Options{
			Callback: func(osPathname string, de *godirwalk.Dirent) error {
				for _, ftype := range fileTypes {
					if !ftype.Regex.MatchString(de.Name()) {
						continue
					}

					d := getFileTime(osPathname, false)
					mediaDate := getMediaDate(d, sortoptions)

					if len(sortoptions.DateRange) == 2 {
						start := sortoptions.DateRange[0]
						end := sortoptions.DateRange[1]
						if d.Before(start) {
							return godirwalk.SkipThis
						}
						if d.After(end) {
							return godirwalk.SkipThis
						}
					}

					dayFolder := getDayFolder("MAX", output, sortoptions, mediaDate)

					switch ftype.Type {
					case Video:
						x := de.Name()
						filename := fmt.Sprintf("%s%s-%s.%s", x[:2], x[4:][:4], x[2:][:2], strings.Split(x, ".")[1])
						foldersNeeded := []string{"videos/360", "videos/heromode"}
						dest := foldersNeeded[1]
						if !ftype.HeroMode {
							dest = foldersNeeded[0]
						}
						folder := filepath.Join(dayFolder, dest)
						result = parse(folder, filename, osPathname, sortoptions, result)
					case Photo:
						foldersNeeded := []string{"photos/360", "photos/heromode"}
						dest := foldersNeeded[1]
						if !ftype.HeroMode {
							dest = foldersNeeded[0]
						}
						folder := filepath.Join(dayFolder, dest)
						result = parse(folder, de.Name(), osPathname, sortoptions, result)
					case PowerPano:
						folder := filepath.Join(dayFolder, "photos/powerpano")
						result = parse(folder, de.Name(), osPathname, sortoptions, result)
					case LowResolutionVideo:
						if sortoptions.SkipAuxiliaryFiles {
							continue
						}
						foldersNeeded := []string{"videos/proxy/heromode", "videos/proxy/360"}
						dest := foldersNeeded[1]
						if ftype.HeroMode {
							dest = foldersNeeded[0]
						}
						x := de.Name()
						filename := fmt.Sprintf("%s%s-%s.%s", x[:2], x[4:][:4], x[2:][:2], strings.Split(x, ".")[1])
						folder := filepath.Join(dayFolder, dest)
						result = parse(folder, filename, osPathname, sortoptions, result)
					case Thumbnail:
						if sortoptions.SkipAuxiliaryFiles {
							continue
						}
						foldersNeeded := []string{"videos/thumbnails/heromode", "videos/thumbnails/360"}
						dest := foldersNeeded[1]
						if ftype.HeroMode {
							dest = foldersNeeded[0]
						}
						x := de.Name()
						filename := fmt.Sprintf("%s%s-%s.%s", x[:2], x[4:][:4], x[2:][:2], strings.Split(x, ".")[1])
						folder := filepath.Join(dayFolder, dest)
						result = parse(folder, filename, osPathname, sortoptions, result)
					default:
						result = unsupported(de, osPathname, result)
					}
				}
				return nil
			},
			Unsorted: true,
		})

		if err != nil {
			result.Errors = append(result.Errors, err)
		}
	}
	return result
}

func importFromGoProV2(root string, output string, sortoptions SortOptions, cameraName string) utils.Result {
	fileTypes := FileTypeMatches[V2]
	var result utils.Result
	/*
		The idea is to have a result like:

		20-02-2021/HERO9_Black/photos/
							GOPR00001.JPG
					   /videos
							single/
								GH010001.MP4
							proxy/
								GL010001.LRV
							thumbnails/
								GL010001.THM


	*/
	folders, err := ioutil.ReadDir(root)
	if err != nil {
		result.Errors = append(result.Errors, err)
		return result
	}

	for _, f := range folders {
		r := MediaFolderRegex.MatchString(f.Name())

		if !r {
			continue
		}
		color.Green("Looking at %s", f.Name())

		err = godirwalk.Walk(filepath.Join(root, f.Name()), &godirwalk.Options{
			Callback: func(osPathname string, de *godirwalk.Dirent) error {
				for _, ftype := range fileTypes {
					if !ftype.Regex.MatchString(de.Name()) {
						continue
					}

					d := getFileTime(osPathname, false)

					mediaDate := getMediaDate(d, sortoptions)

					if len(sortoptions.DateRange) == 2 {
						start := sortoptions.DateRange[0]
						end := sortoptions.DateRange[1]
						if d.Before(start) {
							return godirwalk.SkipThis
						}
						if d.After(end) {
							return godirwalk.SkipThis
						}
					}

					dayFolder := getDayFolder(cameraName, output, sortoptions, mediaDate)

					switch ftype.Type {
					case Video:
						x := de.Name()
						filename := fmt.Sprintf("%s%s-%s.%s", x[:2], x[4:][:4], x[2:][:2], strings.Split(x, ".")[1])
						color.Green(">>> %s", filename)
						s, err := ffprobe.VideoSize(osPathname)
						if err != nil {
							log.Fatal(err.Error())
							return godirwalk.SkipThis
						}
						eval := goval.NewEvaluator()
						framerate, err := eval.Evaluate(s.Streams[0].RFrameRate, nil, nil)
						if err != nil {
							log.Fatal(err.Error())
							return godirwalk.SkipThis
						}
						fpsAsFloat := strconv.Itoa(framerate.(int))

						if err != nil {
							log.Fatal(err.Error())
							return godirwalk.SkipThis
						}
						rfpsFolder := fmt.Sprintf("%dx%d %s", s.Streams[0].Width, s.Streams[0].Height, fpsAsFloat)
						folder := filepath.Join(dayFolder, "videos", rfpsFolder)
						result = parse(folder, filename, osPathname, sortoptions, result)
					case Photo:
						folder := filepath.Join(dayFolder, "photos")
						result = parse(folder, de.Name(), osPathname, sortoptions, result)
					case LowResolutionVideo:
						if sortoptions.SkipAuxiliaryFiles {
							continue
						}
						x := de.Name()
						filename := fmt.Sprintf("%s%s-%s.%s", x[:2], x[4:][:4], x[2:][:2], strings.Split(x, ".")[1])
						folder := filepath.Join(dayFolder, "videos/proxy")
						result = parse(folder, filename, osPathname, sortoptions, result)

					case Thumbnail:
						if sortoptions.SkipAuxiliaryFiles {
							continue
						}
						x := de.Name()
						filename := fmt.Sprintf("%s%s-%s.%s", x[:2], x[4:][:4], x[2:][:2], strings.Split(x, ".")[1])
						folder := filepath.Join(dayFolder, "videos/proxy")
						result = parse(folder, filename, osPathname, sortoptions, result)

					case Multishot:
						folder := filepath.Join(dayFolder, "multishot", de.Name()[:4])
						result = parse(folder, de.Name(), osPathname, sortoptions, result)

					case RawPhoto:
						folder := filepath.Join(dayFolder, "photos/raw")
						result = parse(folder, de.Name(), osPathname, sortoptions, result)

					default:
						result = unsupported(de, osPathname, result)
					}
				}
				return nil
			},
			Unsorted: true,
		})

		if err != nil {
			result.Errors = append(result.Errors, err)
		}
	}
	return result
}

func importFromGoProV1(root string, output string, sortoptions SortOptions, cameraName string) utils.Result {
	fileTypes := FileTypeMatches[V1]
	var result utils.Result

	folders, err := ioutil.ReadDir(root)
	if err != nil {
		result.Errors = append(result.Errors, err)
		return result
	}

	for _, f := range folders {
		r := MediaFolderRegex.MatchString(f.Name())

		if !r {
			continue
		}
		color.Green("Looking at %s", f.Name())

		err = godirwalk.Walk(filepath.Join(root, f.Name()), &godirwalk.Options{
			Callback: func(osPathname string, de *godirwalk.Dirent) error {
				for _, ftype := range fileTypes {
					if !ftype.Regex.MatchString(de.Name()) {
						continue
					}

					d := getFileTime(osPathname, true)

					mediaDate := getMediaDate(d, sortoptions)

					if len(sortoptions.DateRange) == 2 {
						start := sortoptions.DateRange[0]
						end := sortoptions.DateRange[1]
						if d.Before(start) {
							return godirwalk.SkipThis
						}
						if d.After(end) {
							return godirwalk.SkipThis
						}
					}

					dayFolder := getDayFolder(cameraName, output, sortoptions, mediaDate)

					switch ftype.Type {
					case Video:
						x := de.Name()

						chaptered := regexp.MustCompile(`GP\d+.MP4`)
						if chaptered.MatchString(de.Name()) {
							x = fmt.Sprintf("GOPR%s%s.%s", x[4:][:4], x[2:][:2], strings.Split(x, ".")[1])
						}
						s, err := ffprobe.VideoSize(osPathname)
						if err != nil {
							log.Fatal(err.Error())
							return godirwalk.SkipThis
						}
						framerate := strings.ReplaceAll(s.Streams[0].RFrameRate, "/1", "")
						rfpsFolder := fmt.Sprintf("%dx%d %s", s.Streams[0].Width, s.Streams[0].Height, framerate)
						folder := filepath.Join(dayFolder, "videos", rfpsFolder)
						result = parse(folder, x, osPathname, sortoptions, result)

					case ChapteredVideo:
						x := de.Name()
						name := fmt.Sprintf("GOPR%s%s.%s", x[4:][:4], x[2:][:2], strings.Split(x, ".")[1])
						folder := filepath.Join(dayFolder, "videos")
						result = parse(folder, name, osPathname, sortoptions, result)

					case Photo:
						folder := filepath.Join(dayFolder, "photos")
						result = parse(folder, de.Name(), osPathname, sortoptions, result)

					case LowResolutionVideo:
						if sortoptions.SkipAuxiliaryFiles {
							continue
						}
						folder := filepath.Join(dayFolder, "videos/proxy")
						result = parse(folder, de.Name(), osPathname, sortoptions, result)

					case Thumbnail:
						if sortoptions.SkipAuxiliaryFiles {
							continue
						}
						folder := filepath.Join(dayFolder, "videos/proxy")
						result = parse(folder, de.Name(), osPathname, sortoptions, result)

					case Multishot:
						folder := filepath.Join(dayFolder, "multishot", de.Name()[:4])
						result = parse(folder, de.Name(), osPathname, sortoptions, result)

					case RawPhoto:
						folder := filepath.Join(dayFolder, "photos/raw")
						result = parse(folder, de.Name(), osPathname, sortoptions, result)

					default:
						result = unsupported(de, osPathname, result)
					}
				}
				return nil
			},
			Unsorted: true,
		})

		if err != nil {
			result.Errors = append(result.Errors, err)
		}
	}
	return result
}

/*
GoPro adds a trailing comma to their version.txt file... this removes it.
*/
func cleanVersion(s string) string {
	i := strings.LastIndex(s, ",")
	excludingLast := s[:i] + strings.Replace(s[i:], ",", "", 1)

	if strings.Contains(s, "HERO10") || strings.Contains(s, "HERO11") {
		return strings.ReplaceAll(s, "\n", "")
	}
	return excludingLast
}

func readInfo(in string) (*Info, error) {
	jsonFile, err := os.Open(filepath.Join(in, "MISC", fmt.Sprint(Version)))
	if err != nil {
		return nil, err
	}
	inBytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	text := string(inBytes)
	clean := cleanVersion(text)
	var gpVersion Info
	err = json.Unmarshal([]byte(clean), &gpVersion)
	if err != nil {
		return nil, err
	}
	return &gpVersion, nil
}

func getFileTime(osPathname string, utcFix bool) time.Time {
	var d time.Time
	t, err := times.Stat(osPathname)
	if err != nil {
		log.Fatal(err.Error())
	}
	d = t.ModTime()
	if utcFix {
		zoneName, _ := d.Zone()
		newTime := strings.Replace(d.Format(time.UnixDate), zoneName, "UTC", -1)
		d, _ = time.Parse(time.UnixDate, newTime)
	}
	return d
}

func getDayFolder(cameraName string, output string, sortoptions SortOptions, mediaDate string) string {
	dayFolder := filepath.Join(output, mediaDate)
	if _, err := os.Stat(dayFolder); os.IsNotExist(err) {
		_ = os.Mkdir(dayFolder, 0755)
	}

	if sortoptions.ByCamera {
		if _, err := os.Stat(filepath.Join(dayFolder, cameraName)); os.IsNotExist(err) {
			_ = os.Mkdir(filepath.Join(dayFolder, cameraName), 0755)
		}
		dayFolder = filepath.Join(dayFolder, cameraName)
	}
	return dayFolder
}

func getMediaDate(d time.Time, sortoptions SortOptions) string {
	mediaDate := d.Format("02-01-2006")
	if strings.Contains(sortoptions.DateFormat, "yyyy") && strings.Contains(sortoptions.DateFormat, "mm") && strings.Contains(sortoptions.DateFormat, "dd") {
		mediaDate = d.Format(replacer.Replace(sortoptions.DateFormat))
	}
	return mediaDate
}

func parse(folder string, name string, osPathname string, sortoptions SortOptions, result utils.Result) utils.Result {
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		err = os.MkdirAll(folder, 0755)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	color.Green(">>> %s", name)

	err := utils.CopyFile(osPathname, filepath.Join(folder, name), sortoptions.BufferSize)
	if err != nil {
		result.Errors = append(result.Errors, err)
		result.FilesNotImported = append(result.FilesNotImported, osPathname)
	} else {
		result.FilesImported++
	}
	return result
}

func unsupported(de *godirwalk.Dirent, osPathname string, result utils.Result) utils.Result {
	color.Red("Unsupported file %s", de.Name())
	result.Errors = append(result.Errors, errors.New("Unsupported file "+de.Name()))
	result.FilesNotImported = append(result.FilesNotImported, osPathname)
	return result
}
