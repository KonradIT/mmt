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

func getDeviceNameFromPhoto(path string) (string, error) { //nolint:unused
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

type Entrypoint struct{}

func (Entrypoint) Import(params utils.ImportParams) (*utils.Result, error) {
	// Tested on Mavic Air 2. Osmo Pocket v1 and Spark specific changes to follow.

	if params.CameraName == "" {
		params.CameraName = "DJI Device"
	}
	di, err := disk.GetInfo(params.Input)
	if err != nil {
		return nil, err
	}
	percentage := (float64(di.Total-di.Free) / float64(di.Total)) * 100

	color.Cyan("\t💾 %s/%s (%0.2f%%)\n",
		humanize.Bytes(di.Total-di.Free),
		humanize.Bytes(di.Total),
		percentage,
	)

	mediaFolderRegex := regexp.MustCompile(`\d+MEDIA`)

	root := filepath.Join(params.Input, "DCIM")
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

	inlineCounter := utils.ResultCounter{}

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

					mediaDate := d.Format("02-01-2006")
					if strings.Contains(params.DateFormat, "yyyy") && strings.Contains(params.DateFormat, "mm") && strings.Contains(params.DateFormat, "dd") {
						mediaDate = d.Format(utils.DateFormatReplacer.Replace(params.DateFormat))
					}

					// check if is in date range

					if d.Before(params.DateRange[0]) || d.After(params.DateRange[1]) {
						return godirwalk.SkipThis
					}

					info, err := os.Stat(osPathname)
					if err != nil {
						return godirwalk.SkipThis
					}

					wg.Add(1)
					bar := utils.GetNewBar(progressBar, info.Size(), de.Name(), utils.IoTX)

					dayFolder := utils.GetOrder(params.Sort, locationService, osPathname, params.Output, mediaDate, params.CameraName)
					switch ftype.Type {
					case Photo:
						if _, err := os.Stat(filepath.Join(dayFolder, "photos")); os.IsNotExist(err) {
							mkdirerr := os.MkdirAll(filepath.Join(dayFolder, "photos"), 0o755)
							if mkdirerr != nil {
								return godirwalk.SkipThis
							}
						}

						go func(folder, filename, osPathname string, bar *mpb.Bar) {
							defer wg.Done()
							err = utils.CopyFile(osPathname, filepath.Join(dayFolder, "photos", filename), params.BufferSize, bar)
							if err != nil {
								bar.EwmaSetCurrent(info.Size(), 1*time.Millisecond)
								bar.EwmaIncrInt64(info.Size(), 1*time.Millisecond)
								inlineCounter.SetFailure(err, filename)
							} else {
								inlineCounter.SetSuccess()
							}
						}(f.Name(), de.Name(), osPathname, bar)

					case Video:
						if _, err := os.Stat(filepath.Join(dayFolder, "videos")); os.IsNotExist(err) {
							mkdirerr := os.MkdirAll(filepath.Join(dayFolder, "videos"), 0o755)
							if mkdirerr != nil {
								log.Fatal(mkdirerr.Error())
							}
						}

						go func(folder, filename, osPathname string, bar *mpb.Bar) {
							defer wg.Done()
							err = utils.CopyFile(osPathname, filepath.Join(dayFolder, "videos", filename), params.BufferSize, bar)
							if err != nil {
								bar.EwmaSetCurrent(info.Size(), 1*time.Millisecond)
								bar.EwmaIncrInt64(info.Size(), 1*time.Millisecond)
								inlineCounter.SetFailure(err, filename)
							} else {
								inlineCounter.SetSuccess()
							}
						}(f.Name(), de.Name(), osPathname, bar)
					case Subtitle:
						extraPath := srtFolderFromConfig()
						if params.SkipAuxiliaryFiles {
							wg.Done()
							bar.Abort(true)
							break
						}

						if _, err := os.Stat(filepath.Join(dayFolder, "videos", extraPath)); os.IsNotExist(err) {
							mkdirerr := os.MkdirAll(filepath.Join(dayFolder, "videos", extraPath), 0o755)
							if mkdirerr != nil {
								log.Fatal(mkdirerr.Error())
							}
						}

						go func(folder, filename, osPathname string, bar *mpb.Bar) {
							defer wg.Done()
							err = utils.CopyFile(osPathname, filepath.Join(dayFolder, "videos", extraPath, filename), params.BufferSize, bar)
							if err != nil {
								bar.EwmaSetCurrent(info.Size(), 1*time.Millisecond)
								bar.EwmaIncrInt64(info.Size(), 1*time.Millisecond)
								inlineCounter.SetFailure(err, filename)
							} else {
								inlineCounter.SetSuccess()
							}
						}(f.Name(), de.Name(), osPathname, bar)
					case RawPhoto:
						if _, err := os.Stat(filepath.Join(dayFolder, "photos/raw")); os.IsNotExist(err) {
							mkdirerr := os.MkdirAll(filepath.Join(dayFolder, "photos/raw"), 0o755)
							if mkdirerr != nil {
								log.Fatal(mkdirerr.Error())
							}
						}

						go func(folder, filename, osPathname string, bar *mpb.Bar) {
							defer wg.Done()
							err = utils.CopyFile(osPathname, filepath.Join(dayFolder, "photos/raw", filename), params.BufferSize, bar)
							if err != nil {
								bar.EwmaSetCurrent(info.Size(), 1*time.Millisecond)
								bar.EwmaIncrInt64(info.Size(), 1*time.Millisecond)
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

	result.Errors = append(result.Errors, inlineCounter.Get().Errors...)
	result.FilesImported += inlineCounter.Get().FilesImported
	result.FilesNotImported = append(result.FilesNotImported, inlineCounter.Get().FilesNotImported...)

	return &result, nil
}
