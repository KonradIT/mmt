package dji

import (
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

var DeviceName = "DJI Device"

func getDeviceName() string {
	return DeviceName
}

func Import(in, out, dateFormat string, bufferSize int, prefix string, dateRange []string) (*utils.Result, error) {

	// Tested on Mavic Air 2. Osmo Pocket v1 and Spark specific changes to follow.

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
	panoramaFolder := "PANORAMA"

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
		if f.Name() == panoramaFolder {

			panoramaBatches, err := ioutil.ReadDir(filepath.Join(root, panoramaFolder))
			if err != nil {
				result.Errors = append(result.Errors, err)
				return &result, nil
			}

			for _, panoramaId := range panoramaBatches {
				t, err := times.Stat(filepath.Join(root, panoramaFolder, panoramaId.Name()))
				if err != nil {
					log.Fatal(err.Error())
					continue
				}
				if t.HasBirthTime() {
					d := t.BirthTime()
					// check if is in date range
					replacer := strings.NewReplacer("dd", "02", "mm", "01", "yyyy", "2006")

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
							continue
						}
						if d.After(dateEnd) {
							continue
						}
					}

					if len(dateRange) == 2 {
						layout := replacer.Replace(dateFormat)
						start, err1 := time.Parse(layout, dateRange[0])
						end, err2 := time.Parse(layout, dateRange[1])
						if err1 == nil && err2 == nil {
							if d.Before(start) {
								continue
							}
							if d.After(end) {
								continue
							}
						}

					}
					mediaDate := d.Format("02-01-2006")
					if strings.Contains(dateFormat, "yyyy") && strings.Contains(dateFormat, "mm") && strings.Contains(dateFormat, "dd") {
						mediaDate = d.Format(replacer.Replace(dateFormat))
					}

					dayFolder := filepath.Join(out, mediaDate, getDeviceName(), "photos/panoramas")
					if _, err := os.Stat(dayFolder); os.IsNotExist(err) {
						os.Mkdir(dayFolder, 0755)
					}
					err = utils.CopyDir(filepath.Join(root, panoramaFolder, panoramaId.Name()), filepath.Join(dayFolder, panoramaId.Name()), bufferSize)
					if err != nil {
						result.Errors = append(result.Errors, err)
						result.FilesNotImported = append(result.FilesNotImported, panoramaId.Name())
					} else {
						result.FilesImported += 1
					}
				}
			}

		}

		r, err := regexp.MatchString(mediaFolder, f.Name())
		if err != nil {
			result.Errors = append(result.Errors, err)
		}
		if !r {
			continue
		}

		color.Green("Looking at %s", f.Name())

		err = godirwalk.Walk(filepath.Join(root, f.Name()), &godirwalk.Options{
			Unsorted: true,
			Callback: func(osPathname string, de *godirwalk.Dirent) error {

				for _, ftype := range fileTypes {
					if ftype.Regex.MatchString(de.Name()) {
						t, err := times.Stat(osPathname)
						if err != nil {
							log.Fatal(err.Error())
						}
						if t.HasBirthTime() {
							d := t.BirthTime()
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

							if len(dateRange) == 2 {

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

							dayFolder := filepath.Join(out, mediaDate)
							if _, err := os.Stat(dayFolder); os.IsNotExist(err) {
								os.Mkdir(dayFolder, 0755)
							}

							if _, err := os.Stat(filepath.Join(dayFolder, getDeviceName())); os.IsNotExist(err) {
								os.Mkdir(filepath.Join(dayFolder, getDeviceName()), 0755)
							}
							dayFolder = filepath.Join(dayFolder, getDeviceName())

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
									result.FilesImported += 1
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
								err = os.Rename(dayFolder, strings.Replace(dayFolder, DeviceName, s, 1))
								if err != nil {

									// Could be a folder allready exists... time to move the content to that folder.

								}
								DeviceName = s

								break
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
									result.FilesImported += 1
								}
								break
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
									result.FilesImported += 1
								}
								break
							case PanoramaIndex:

							}
						}
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
