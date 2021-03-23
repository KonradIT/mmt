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
	"gopkg.in/djherbis/times.v1"
)

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

								dayFolder := filepath.Join(out, mediaDate)
								if _, err := os.Stat(dayFolder); os.IsNotExist(err) {
									os.Mkdir(dayFolder, 0755)
								}

								if _, err := os.Stat(filepath.Join(dayFolder, "DJI Drone")); os.IsNotExist(err) {
									os.Mkdir(filepath.Join(dayFolder, "DJI Drone"), 0755)
								}
								dayFolder = filepath.Join(dayFolder, "DJI Drone")

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

									err = utils.CopyFile(osPathname, filepath.Join(dayFolder, "photos", x), 1000)
									if err != nil {
										result.Errors = append(result.Errors, err)
										result.FilesNotImported = append(result.FilesNotImported, osPathname)
									} else {
										result.FilesImported += 1
									}
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
									err = utils.CopyFile(osPathname, filepath.Join(dayFolder, "videos", x), 1000)
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

									err = utils.CopyFile(osPathname, filepath.Join(dayFolder, "photos/raw", x), 1000)
									if err != nil {
										result.Errors = append(result.Errors, err)
										result.FilesNotImported = append(result.FilesNotImported, osPathname)
									} else {
										result.FilesImported += 1
									}
									break
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
