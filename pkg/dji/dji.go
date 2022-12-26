package dji

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/karrick/godirwalk"
	"github.com/konradit/mmt/pkg/utils"
	"github.com/minio/minio/pkg/disk"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/vbauerster/mpb/v8"
	"gopkg.in/djherbis/times.v1"
)

func getDeviceNameFromPhoto(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	exifData, err := exif.Decode(f)
	if err != nil {
		return "", err
	}

	camModel, err := exifData.Get(exif.Model)
	if err != nil {
		return "", err
	}
	s, err := camModel.StringVal()
	if err != nil {
		return "", err
	}
	return s, nil
}

var locationService = LocationService{}

func Import(in, out, dateFormat string, bufferSize int, prefix string, dateRange []string, cameraOptions map[string]interface{}) (*utils.Result, error) {
	// Tested on Mavic Air 2. Osmo Pocket v1 and Spark specific changes to follow.
	sortOptions := utils.ParseCliOptions(cameraOptions)

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

	root := filepath.Join(in, "DCIM")
	var result utils.Result

	folders, err := ioutil.ReadDir(root)
	if err != nil {
		result.Errors = append(result.Errors, err)
		return &result, nil
	}

	var wg sync.WaitGroup
	progressBar := mpb.New(mpb.WithWaitGroup(&wg),
		mpb.WithWidth(60),
		mpb.WithRefreshRate(180*time.Millisecond))

	inlineCounter := utils.ResultCounter{
		CameraName: "DJI Device",
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
						return godirwalk.SkipThis
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

					info, err := os.Stat(osPathname)
					if err != nil {
						return godirwalk.SkipThis
					}

					wg.Add(1)
					bar := utils.GetNewBar(progressBar, int64(info.Size()), de.Name())

					dayFolder := utils.GetOrder(sortOptions, locationService, osPathname, out, mediaDate, "DJI Device")
					switch ftype.Type {
					case Photo:
						if _, err := os.Stat(filepath.Join(dayFolder, "photos")); os.IsNotExist(err) {
							err = os.MkdirAll(filepath.Join(dayFolder, "photos"), 0755)
							if err != nil {
								return godirwalk.SkipThis
							}
						}

						go func(folder, filename, osPathname string, bar *mpb.Bar) {
							defer wg.Done()
							err = utils.CopyFile(osPathname, filepath.Join(dayFolder, "photos", filename), bufferSize, bar)
							if err != nil {
								inlineCounter.SetFailure(err, filename)
							} else {
								inlineCounter.SetSuccess()

								// Get Device Name

								devName, err := getDeviceNameFromPhoto(osPathname)
								if err != nil {
									inlineCounter.SetFailure(err, filename)
									return
								}
								// Rename directory
								matchDeviceName, is := DeviceNames[devName]
								if is {
									devName = matchDeviceName
								}

								if err != nil {
									inlineCounter.SetFailure(err, filename)
									return
								}
								inlineCounter.SetCameraName(devName)
							}
						}(f.Name(), de.Name(), osPathname, bar)

					case Video, Subtitle:
						if _, err := os.Stat(filepath.Join(dayFolder, "videos")); os.IsNotExist(err) {
							err = os.MkdirAll(filepath.Join(dayFolder, "videos"), 0755)
							if err != nil {
								log.Fatal(err.Error())
							}
						}

						go func(folder, filename, osPathname string, bar *mpb.Bar) {
							defer wg.Done()
							err = utils.CopyFile(osPathname, filepath.Join(dayFolder, "videos", filename), bufferSize, bar)
							if err != nil {
								inlineCounter.SetFailure(err, filename)
							} else {
								inlineCounter.SetSuccess()
							}
						}(f.Name(), de.Name(), osPathname, bar)
					case RawPhoto:
						if _, err := os.Stat(filepath.Join(dayFolder, "photos/raw")); os.IsNotExist(err) {
							err = os.MkdirAll(filepath.Join(dayFolder, "photos/raw"), 0755)
							if err != nil {
								log.Fatal(err.Error())
							}
						}

						go func(folder, filename, osPathname string, bar *mpb.Bar) {
							defer wg.Done()
							err = utils.CopyFile(osPathname, filepath.Join(dayFolder, "photos/raw", filename), bufferSize, bar)
							if err != nil {
								inlineCounter.SetFailure(err, filename)
							} else {
								inlineCounter.SetSuccess()
							}
						}(f.Name(), de.Name(), osPathname, bar)
					case PanoramaIndex:
					case Audio:
						// TODO get audio files
					}
				}

				return nil
			},
		})

		if err != nil {
			inlineCounter.SetFailure(err, "")
		}
	}

	wg.Wait()
	progressBar.Shutdown()

	// Rename each folder

	_ = godirwalk.Walk(out, &godirwalk.Options{
		Unsorted: true,
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if !de.ModeType().IsDir() {
				return godirwalk.SkipThis
			}

			modified, err := utils.FindFolderInPath(osPathname, "DJI Device")
			if err == nil {
				_ = os.Rename(modified, strings.Replace(modified, "DJI Device", inlineCounter.CameraName, -1)) // Could be a folder already exists... time to move the content to that folder.
			}
			return nil
		},
	})

	result.Errors = append(result.Errors, inlineCounter.Get().Errors...)
	result.FilesImported += inlineCounter.Get().FilesImported
	result.FilesNotImported = append(result.FilesNotImported, inlineCounter.Get().FilesNotImported...)

	return &result, nil
}
