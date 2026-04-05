package dji

import (
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/karrick/godirwalk"
	"github.com/konradit/mmt/pkg/camera"
	mErrors "github.com/konradit/mmt/pkg/errors"
	"github.com/konradit/mmt/pkg/utils"
	diskinfo "github.com/minio/minio/pkg/disk"
	gopsutil "github.com/shirou/gopsutil/disk"
	"github.com/vbauerster/mpb/v8"
	"gopkg.in/djherbis/times.v1"
)

var locationService = LocationService{}

type Entrypoint struct{}

func init() {
	camera.Register(Entrypoint{})
}

func (Entrypoint) Name() string { return "dji" }

func (Entrypoint) GuessFromPath(root string) bool {
	_, err := os.Stat(filepath.Join(root, "MISC", "GIS", "dji.gis"))

	return err == nil
}

func (Entrypoint) Detect() (string, camera.ConnectionType, error) {
	partitions, err := gopsutil.Partitions(false)
	if err != nil {
		return "", "", err
	}

	for _, partition := range partitions {
		if (Entrypoint{}).GuessFromPath(partition.Mountpoint) {
			return partition.Mountpoint, camera.SDCard, nil
		}
	}

	return "", "", mErrors.ErrNoCameraDetected
}

func (Entrypoint) Import(params camera.ImportParams) (*camera.Result, error) {
	if params.CameraName == "" {
		params.CameraName = "DJI Device"
	}

	di, err := diskinfo.GetInfo(params.Input)
	if err != nil {
		return nil, err
	}

	percentage := (float64(di.Total-di.Free) / float64(di.Total)) * 100

	color.Cyan("\t💾 %s/%s (%0.2f%%)\n",
		humanize.Bytes(di.Total-di.Free),
		humanize.Bytes(di.Total),
		percentage,
	)

	mediaFolderRegex := regexp.MustCompile(`\d+MEDIA|DJI_\d+`)

	root := filepath.Join(params.Input, "DCIM")

	var result camera.Result

	folders, err := os.ReadDir(root)
	if err != nil {
		result.Errors = append(result.Errors, err)

		return &result, nil
	}

	var wg sync.WaitGroup

	sem := camera.NewSemaphore(params.MaxConcurrent)
	progressBar := mpb.New(mpb.WithWaitGroup(&wg),
		mpb.WithWidth(60),
		mpb.WithRefreshRate(180*time.Millisecond))

	inlineCounter := camera.ResultCounter{}

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

					mediaDate := camera.FormatMediaDate(d, params.DateFormat)

					if d.Before(params.DateRange[0]) || d.After(params.DateRange[1]) {
						return godirwalk.SkipThis
					}

					info, err := os.Stat(osPathname)
					if err != nil {
						return godirwalk.SkipThis
					}

					wg.Add(1)

					bar := camera.GetNewBar(progressBar, info.Size(), de.Name(), camera.IoTX)

					dayFolder := camera.GetOrder(params.Sort, locationService, osPathname, params.Output, mediaDate, params.CameraName)

					switch ftype.Type {
					case Photo:
						err := os.MkdirAll(filepath.Join(dayFolder, "photos"), 0o755)
						if err != nil {
							return godirwalk.SkipThis
						}

						sem <- struct{}{}

						go func(filename, osPathname string, bar *mpb.Bar) {
							defer func() { <-sem }()
							defer wg.Done()

							err = utils.CopyFile(osPathname, filepath.Join(dayFolder, "photos", filename), params.BufferSize, bar, d)
							if err != nil {
								bar.EwmaSetCurrent(info.Size(), 1*time.Millisecond)
								bar.EwmaIncrInt64(info.Size(), 1*time.Millisecond)
								inlineCounter.SetFailure(err, filename)
							} else {
								inlineCounter.SetSuccess()
							}
						}(de.Name(), osPathname, bar)

					case Video:
						err := os.MkdirAll(filepath.Join(dayFolder, "videos"), 0o755)
						if err != nil {
							return godirwalk.SkipThis
						}

						sem <- struct{}{}

						go func(filename, osPathname string, bar *mpb.Bar) {
							defer func() { <-sem }()
							defer wg.Done()

							err = utils.CopyFile(osPathname, filepath.Join(dayFolder, "videos", filename), params.BufferSize, bar, d)
							if err != nil {
								bar.EwmaSetCurrent(info.Size(), 1*time.Millisecond)
								bar.EwmaIncrInt64(info.Size(), 1*time.Millisecond)
								inlineCounter.SetFailure(err, filename)
							} else {
								inlineCounter.SetSuccess()
							}
						}(de.Name(), osPathname, bar)
					case Subtitle:
						extraPath := srtFolderFromConfig()

						if params.SkipAuxiliaryFiles {
							wg.Done()
							bar.Abort(true)

							break
						}

						err := os.MkdirAll(filepath.Join(dayFolder, "videos", extraPath), 0o755)
						if err != nil {
							return godirwalk.SkipThis
						}

						sem <- struct{}{}

						go func(filename, osPathname string, bar *mpb.Bar) {
							defer func() { <-sem }()
							defer wg.Done()

							err = utils.CopyFile(osPathname, filepath.Join(dayFolder, "videos", extraPath, filename), params.BufferSize, bar, d)
							if err != nil {
								bar.EwmaSetCurrent(info.Size(), 1*time.Millisecond)
								bar.EwmaIncrInt64(info.Size(), 1*time.Millisecond)
								inlineCounter.SetFailure(err, filename)
							} else {
								inlineCounter.SetSuccess()
							}
						}(de.Name(), osPathname, bar)
					case RawPhoto:
						err := os.MkdirAll(filepath.Join(dayFolder, "photos/raw"), 0o755)
						if err != nil {
							return godirwalk.SkipThis
						}

						sem <- struct{}{}

						go func(filename, osPathname string, bar *mpb.Bar) {
							defer func() { <-sem }()
							defer wg.Done()

							err = utils.CopyFile(osPathname, filepath.Join(dayFolder, "photos/raw", filename), params.BufferSize, bar, d)
							if err != nil {
								bar.EwmaSetCurrent(info.Size(), 1*time.Millisecond)
								bar.EwmaIncrInt64(info.Size(), 1*time.Millisecond)
								inlineCounter.SetFailure(err, filename)
							} else {
								inlineCounter.SetSuccess()
							}
						}(de.Name(), osPathname, bar)
					case PanoramaIndex:
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
