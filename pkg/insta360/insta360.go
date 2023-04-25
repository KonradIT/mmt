package insta360

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/karrick/godirwalk"
	"github.com/konradit/mmt/pkg/media"
	"github.com/konradit/mmt/pkg/utils"
	"github.com/minio/minio/pkg/disk"
	"github.com/vbauerster/mpb/v8"
)

func getDeviceName(manifest string) string {
	name := "Insta360 Camera"
	file, err := os.ReadFile(manifest)
	if err != nil {
		return name
	}

	endBytes := []byte{0x1A, 0x0F}

	res := bytes.Split(file, append([]byte{0x12, 0x0B}, []byte("Insta360")...))
	if len(res) == 1 {
		return name
	}
	modelName := bytes.Split(res[1], endBytes)
	if len(modelName) == 1 {
		return name
	}
	return fmt.Sprintf("Insta360%s", modelName[0])
}

type Entrypoint struct{}

func (Entrypoint) Import(params utils.ImportParams) (*utils.Result, error) {
	if params.CameraName == "" {
		params.CameraName = getDeviceName(filepath.Join(params.Input, "DCIM", "fileinfo_list.list"))
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

	mediaFolder := `Camera\d+`
	mediaFolderRegex := regexp.MustCompile(mediaFolder)

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

		err = godirwalk.Walk(filepath.Join(root, f.Name()), &godirwalk.Options{
			Unsorted: true,
			Callback: func(osPathname string, de *godirwalk.Dirent) error {
				for _, ftype := range fileTypes {
					if !ftype.Regex.MatchString(de.Name()) {
						continue
					}

					d := media.GetFileTime(osPathname, true)
					mediaDate := media.GetMediaDate(d, params.DateFormat)

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
					dayFolder := utils.GetOrder(params.Sort, nil, osPathname, params.Output, mediaDate, params.CameraName)

					x := de.Name()

					switch ftype.Type {
					case Photo, RawPhoto:
						id := x[3+8+2 : 3+8+6+2]
						if _, err := os.Stat(filepath.Join(dayFolder, "photos", id)); os.IsNotExist(err) {
							mkdirerr := os.MkdirAll(filepath.Join(dayFolder, "photos", id), 0o755)
							if mkdirerr != nil {
								log.Fatal(mkdirerr.Error())
							}
						}

						go func(id, filename, osPathname string, bar *mpb.Bar) {
							defer wg.Done()

							err = utils.CopyFile(osPathname, filepath.Join(dayFolder, "photos", id, x), params.BufferSize, bar, d)
							if err != nil {
								bar.EwmaSetCurrent(info.Size(), 1*time.Millisecond)
								bar.EwmaIncrInt64(info.Size(), 1*time.Millisecond)
								inlineCounter.SetFailure(err, filename)
							} else {
								inlineCounter.SetSuccess()
							}
						}(id, x, osPathname, bar)
					case Video, LowResolutionVideo:
						if params.SkipAuxiliaryFiles && ftype.Type == LowResolutionVideo {
							wg.Done()
							bar.Abort(true)
							break
						}
						slug := ""
						if ftype.SteadyCamMode {
							slug = "videos/flat"
							if ftype.ProMode {
								slug = "videos/flat/pro_mode"
							}
						} else {
							slug = "videos/360"
						}
						id := x[3+8+2 : 3+8+6+2]
						if ftype.ProMode {
							id = x[3+3+8+2+1 : 3+3+8+6+2+1]
						}
						if _, err := os.Stat(filepath.Join(dayFolder, slug, id)); os.IsNotExist(err) {
							mkdirerr := os.MkdirAll(filepath.Join(dayFolder, slug, id), 0o755)
							if mkdirerr != nil {
								log.Fatal(mkdirerr.Error())
							}
						}

						go func(id, filename, osPathname string, bar *mpb.Bar) {
							defer wg.Done()

							err = utils.CopyFile(osPathname, filepath.Join(dayFolder, slug, id, x), params.BufferSize, bar, d)
							if err != nil {
								bar.EwmaSetCurrent(info.Size(), 1*time.Millisecond)
								bar.EwmaIncrInt64(info.Size(), 1*time.Millisecond)
								inlineCounter.SetFailure(err, filename)
							} else {
								inlineCounter.SetSuccess()
							}
						}(id, x, osPathname, bar)
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
