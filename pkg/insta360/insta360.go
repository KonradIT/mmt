package insta360

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
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

func GetTag(st interface{}, fieldName string, tagName string) (string, bool) {
	t := reflect.TypeOf(st)
	f, _ := t.FieldByName(fieldName)
	return f.Tag.Lookup(tagName)
}
func Import(in, out, dateFormat string, bufferSize int, prefix string, dateRange []string) (*utils.Result, error) {

	// Tested on X2

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

	model := ""

	mediaFolder := `Camera\d+`

	fileTypes := []FileTypeMatch{
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
	}

	root := filepath.Join(in, "DCIM")
	var result utils.Result

	folders, err := ioutil.ReadDir(root)
	if err != nil {
		result.Errors = append(result.Errors, err)
		return &result, nil
	}

	for _, f := range folders {
		r, err := regexp.MatchString(mediaFolder, f.Name())
		if err != nil {
			result.Errors = append(result.Errors, err)
		}
		if r {
			if model != "" {
				color.Green("Looking at %s", f.Name())
			}
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
								if strings.Contains(dateFormat, "year") && strings.Contains(dateFormat, "month") && strings.Contains(dateFormat, "day") {
									mediaDate = d.Format(replacer.Replace(dateFormat))
								}

								// check if is in date range

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

								if len(dateRange) == 1 {
									dateEnd := time.Now()
									dateStart := dateEnd
									switch dateRange[0] {
									case "today":
										dateStart = time.Date(dateEnd.Year(), dateEnd.Month(), dateEnd.Day(), 0, 0, 0, 0, dateEnd.Location())
									case "yesterday":
										dateStart = time.Date(dateEnd.Year(), dateEnd.Month(), dateEnd.Day(), 0, 0, 0, 0, dateEnd.Location()).Add(-24 * time.Hour)
									case "week":
										dateStart = time.Date(dateEnd.Year(), dateEnd.Month(), dateEnd.Day(), 0, 0, 0, 0, dateEnd.Location()).Add(-24 * time.Duration((int(dateEnd.Weekday()) - 1)) * time.Hour)
									}

									if dateStart != dateEnd {
										if d.Before(dateStart) {
											return godirwalk.SkipThis
										}
										if d.After(dateEnd) {
											return godirwalk.SkipThis
										}
									}
								}

								dayFolder := filepath.Join(out, mediaDate)
								if _, err := os.Stat(dayFolder); os.IsNotExist(err) {
									os.Mkdir(dayFolder, 0755)
								}

								if _, err := os.Stat(filepath.Join(dayFolder, "Insta360 Camera")); os.IsNotExist(err) {
									os.Mkdir(filepath.Join(dayFolder, "Insta360 Camera"), 0755)
								}
								dayFolder = filepath.Join(dayFolder, "Insta360 Camera")

								switch ftype.Type {
								case Photo:

									// get model first
									if model == "" {

										f, err := os.Open(osPathname)
										if err != nil {
											log.Fatal(err)
										}

										exifObj, err := exif.Decode(f)
										if err != nil {
											log.Fatal(err)
										}

										camModel, err := exifObj.Get(exif.Model)
										if err == nil {
											if m, err := camModel.StringVal(); err == nil {
												model = m
											}
										}

										color.Cyan("\t🎥 [%s]:", model)
									}
									if ftype.OSCMode {
										x := de.Name()

										color.Green(">>> %s", x)

										// 3 = IMG
										// 8 = date
										// 2 = jump to next + "_"
										// 6 = id
										id := x[3+8+2 : 3+8+6+2]
										if _, err := os.Stat(filepath.Join(dayFolder, "photos/osc_mode", id)); os.IsNotExist(err) {
											err = os.MkdirAll(filepath.Join(dayFolder, "photos/osc_mode", id), 0755)
											if err != nil {
												log.Fatal(err.Error())
											}
										}

										err = utils.CopyFile(osPathname, filepath.Join(dayFolder, "photos/osc_mode", id, x), bufferSize)
										if err != nil {
											result.Errors = append(result.Errors, err)
											result.FilesNotImported = append(result.FilesNotImported, osPathname)
										} else {
											result.FilesImported += 1
										}

									} else {
										x := de.Name()

										color.Green(">>> %s", x)

										if _, err := os.Stat(filepath.Join(dayFolder, "photos")); os.IsNotExist(err) {
											err = os.MkdirAll(filepath.Join(dayFolder, "photos"), 0755)
											if err != nil {
												log.Fatal(err.Error())
											}
										}

										// 3 = IMG
										// 8 = date
										// 2 = jump to next + "_"
										// 6 = id
										id := x[3+8+2 : 3+8+6+2]
										if _, err := os.Stat(filepath.Join(dayFolder, "photos", id)); os.IsNotExist(err) {
											err = os.MkdirAll(filepath.Join(dayFolder, "photos", id), 0755)
											if err != nil {
												log.Fatal(err.Error())
											}
										}

										err = utils.CopyFile(osPathname, filepath.Join(dayFolder, "photos", id, x), bufferSize)
										if err != nil {
											result.Errors = append(result.Errors, err)
											result.FilesNotImported = append(result.FilesNotImported, osPathname)
										} else {
											result.FilesImported += 1
										}

									}
								case Video, LowResolutionVideo:
									slug := ""
									if ftype.SteadyCamMode {
										slug = "videos/steady_cam"
									} else if ftype.ProMode {
										slug = "videos/pro_mode"
									} else {
										slug = "videos/360"
									}

									if slug == "" {
										result.Errors = append(result.Errors, errors.New("Video file "+de.Name()+" not recognized"))
										result.FilesNotImported = append(result.FilesNotImported, osPathname)

									} else {
										x := de.Name()

										color.Green(">>> %s", x)

										// 3 = IMG
										// 8 = date
										// 2 = jump to next + "_"
										// 6 = id
										id := x[3+8+2 : 3+8+6+2]
										if _, err := os.Stat(filepath.Join(dayFolder, slug, id)); os.IsNotExist(err) {
											err = os.MkdirAll(filepath.Join(dayFolder, slug, id), 0755)
											if err != nil {
												log.Fatal(err.Error())
											}
										}
										err = utils.CopyFile(osPathname, filepath.Join(dayFolder, slug, id, x), bufferSize)
										if err != nil {
											result.Errors = append(result.Errors, err)
											result.FilesNotImported = append(result.FilesNotImported, osPathname)
										} else {
											result.FilesImported += 1
										}

									}
								case RawPhoto:
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
	}
	return &result, nil
}
