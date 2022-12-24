package dji

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/karrick/godirwalk"
	"github.com/konradit/mmt/pkg/utils"
	"github.com/minio/minio/pkg/disk"
	"github.com/rwcarlsen/goexif/exif"
	"gopkg.in/djherbis/times.v1"
)

var deviceName = "DJI Device"

func getDeviceName() string {
	return deviceName
}

var locationService = LocationService{}

func Import(in, out, dateFormat string, bufferSize int, prefix string, dateRange []string, cameraOptions map[string]interface{}) (*utils.Result, error) {
	// Tested on Mavic Air 2. Osmo Pocket v1 and Spark specific changes to follow.
	byCamera := false
	byLocation := false

	sortByOptions, found := cameraOptions["sort_by"]
	if found {
		for _, sortop := range sortByOptions.([]string) {
			if sortop == "camera" {
				byCamera = true
			}
			if sortop == "location" {
				byLocation = true
			}

			if sortop != "camera" && sortop != "days" && sortop != "location" {
				return nil, errors.New("Unrecognized option for sort_by: " + sortop)
			}
		}
	}

	sortOptions := utils.SortOptions{
		ByCamera:   byCamera,
		ByLocation: byLocation,
		DateFormat: dateFormat,
		BufferSize: bufferSize,
		Prefix:     prefix,
	}

	di, err := disk.GetInfo(in)
	if err != nil {
		return nil, err
	}
	percentage := (float64(di.Total-di.Free) / float64(di.Total)) * 100

	color.Cyan("\t💾 %s/%s (%0.2f%%)\n",
		humanize.Bytes(di.Total-di.Free),
		humanize.Bytes(di.Total),
		percentage,
	)

	mediaFolder := `\d+MEDIA`
	mediaFolderRegex, err := regexp.Compile(mediaFolder)
	if err != nil {
		return nil, err
	}

	fileTypes := []FileTypeMatch{
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

	root := filepath.Join(in, "DCIM")
	var result utils.Result

	folders, err := ioutil.ReadDir(root)
	if err != nil {
		result.Errors = append(result.Errors, err)
		return &result, nil
	}

	for _, f := range folders {
		r := mediaFolderRegex.MatchString(f.Name())
		if !r {
			continue
		}

		color.Green("Looking at %s", f.Name())

		err = godirwalk.Walk(filepath.Join(root, f.Name()), &godirwalk.Options{
			Unsorted: true,
			Callback: func(osPathname string, de *godirwalk.Dirent) error {
				for _, ftype := range fileTypes {
					if !ftype.Regex.MatchString(de.Name()) {
						continue
					}
					t, err := times.Stat(osPathname)
					if err != nil {
						log.Fatal(err.Error())
					}
					d := t.ModTime()
					replacer := strings.NewReplacer("dd", "02", "mm", "01", "yyyy", "2006")

					mediaDate := d.Format("02-01-2006")
					if strings.Contains(dateFormat, "yyyy") && strings.Contains(dateFormat, "mm") && strings.Contains(dateFormat, "dd") {
						mediaDate = d.Format(replacer.Replace(dateFormat))
					}

					// check if is in date range
					if len(dateRange) == 1 {
						dateStart := time.Date(0000, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
						dateEnd := time.Now()
						switch dateRange[0] {
						case "today":
							dateStart = time.Date(dateEnd.Year(), dateEnd.Month(), dateEnd.Day(), 0, 0, 0, 0, dateEnd.Location())
						case "yesterday":
							dateStart = time.Date(dateEnd.Year(), dateEnd.Month(), dateEnd.Day(), 0, 0, 0, 0, dateEnd.Location()).Add(-24 * time.Hour)
						case "week":
							dateStart = time.Date(dateEnd.Year(), dateEnd.Month(), dateEnd.Day(), 0, 0, 0, 0, dateEnd.Location()).Add(-24 * time.Duration((int(dateEnd.Weekday()) - 1)) * time.Hour)
						}

						if d.Before(dateStart) {
							return godirwalk.SkipThis
						}
						if d.After(dateEnd) {
							return godirwalk.SkipThis
						}
					}

					if len(dateRange) == 2 { //nolint:nestif
						layout := replacer.Replace(dateFormat)

						start, err1 := time.Parse(layout, dateRange[0])
						end, err2 := time.Parse(layout, dateRange[1])
						if err1 == nil && err2 == nil {
							if d.Before(start) {
								return godirwalk.SkipThis
							}
							if d.After(end) {
								return godirwalk.SkipThis
							}
						}
					}

					dayFolder := utils.GetOrder(sortOptions, locationService, osPathname, out, mediaDate, getDeviceName())
					switch ftype.Type {
					case Photo:

						x := de.Name()

						color.Green(">>> %s", x)

						if _, err := os.Stat(filepath.Join(dayFolder, "photos")); os.IsNotExist(err) {
							err = os.MkdirAll(filepath.Join(dayFolder, "photos"), 0755)
							if err != nil {
								log.Fatal(err.Error())
							}
						}

						err = utils.CopyFile(osPathname, filepath.Join(dayFolder, "photos", x), bufferSize)
						if err != nil {
							result.Errors = append(result.Errors, err)
							result.FilesNotImported = append(result.FilesNotImported, osPathname)
						} else {
							result.FilesImported++
						}

						// Get Device Name

						f, err := os.Open(osPathname)
						if err != nil {
							log.Fatal(err.Error())
							return godirwalk.SkipThis
						}
						exifData, err := exif.Decode(f)
						if err != nil {
							log.Fatal(err.Error())
							return godirwalk.SkipThis
						}

						camModel, err := exifData.Get(exif.Model)
						if err != nil {
							log.Fatal(err.Error())
							return godirwalk.SkipThis
						}
						s, err := camModel.StringVal()
						if err != nil {
							log.Fatal(err.Error())
							return godirwalk.SkipThis
						}

						// Rename directory
						matchDeviceName, is := DeviceNames[s]
						if is {
							s = matchDeviceName
						}
						_ = os.Rename(dayFolder, strings.Replace(dayFolder, deviceName, s, 1)) // Could be a folder already exists... time to move the content to that folder.

						deviceName = s
					case Video, Subtitle:

						x := de.Name()

						color.Green(">>> %s", x)

						if _, err := os.Stat(filepath.Join(dayFolder, "videos")); os.IsNotExist(err) {
							err = os.MkdirAll(filepath.Join(dayFolder, "videos"), 0755)
							if err != nil {
								log.Fatal(err.Error())
							}
						}
						err = utils.CopyFile(osPathname, filepath.Join(dayFolder, "videos", x), bufferSize)
						if err != nil {
							result.Errors = append(result.Errors, err)
							result.FilesNotImported = append(result.FilesNotImported, osPathname)
						} else {
							result.FilesImported++
						}
					case RawPhoto:
						x := de.Name()

						color.Green(">>> %s", x)

						if _, err := os.Stat(filepath.Join(dayFolder, "photos/raw")); os.IsNotExist(err) {
							err = os.MkdirAll(filepath.Join(dayFolder, "photos/raw"), 0755)
							if err != nil {
								log.Fatal(err.Error())
							}
						}

						err = utils.CopyFile(osPathname, filepath.Join(dayFolder, "photos/raw", x), bufferSize)
						if err != nil {
							result.Errors = append(result.Errors, err)
							result.FilesNotImported = append(result.FilesNotImported, osPathname)
						} else {
							result.FilesImported++
						}
					case PanoramaIndex:
					case Audio:
						// TODO get audio files
					}
				}

				return nil
			},
		})

		if err != nil {
			result.Errors = append(result.Errors, err)
		}
	}
	return &result, nil
}
