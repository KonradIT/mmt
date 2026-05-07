package gopro

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/karrick/godirwalk"
	"github.com/konradit/mmt/pkg/camera"
	mErrors "github.com/konradit/mmt/pkg/errors"
	"github.com/konradit/mmt/pkg/utils"
	"github.com/maja42/goval"
	"github.com/minio/minio/pkg/disk"
	"github.com/vbauerster/mpb/v8"
	"gopkg.in/djherbis/times.v1"
)

/*
Uses data from:
https://community.gopro.com/t5/en/GoPro-Camera-File-Naming-Convention/ta-p/390220#
*/

var MediaFolderRegex = regexp.MustCompile(`\d\d\dGOPRO`)

var ffprobe = utils.NewFFprobe(nil)

var locationService = LocationService{}

func init() {
	camera.Register(Entrypoint{})
}

func getRfpsFolder(pathName string) (string, error) {
	if filepath.Ext(pathName) == ".360" {
		return "", nil
	}

	s, err := ffprobe.VideoSize(pathName)
	if err != nil {
		return "", err
	}

	eval := goval.NewEvaluator()

	framerate, err := eval.Evaluate(s.Streams[0].RFrameRate, nil, nil)
	if err != nil {
		return "", err
	}

	framerateInt, ok := framerate.(int)
	if !ok {
		return "", fmt.Errorf("unexpected framerate type: %T", framerate)
	}

	fpsAsFloat := strconv.Itoa(framerateInt)

	return fmt.Sprintf("%dx%d %s", s.Streams[0].Width, s.Streams[0].Height, fpsAsFloat), nil
}

type Entrypoint struct{}

func (Entrypoint) Name() string { return "gopro" }

func (Entrypoint) GuessFromPath(root string) bool {
	_, err := os.Stat(filepath.Join(root, "MISC", "version.txt"))

	return err == nil
}

func (e Entrypoint) UpdateFirmware(input string, _ camera.UpdateOptions) error {
	return UpdateCamera(input)
}

func (Entrypoint) Import(params camera.ImportParams) (*camera.Result, error) {
	/* Import method using SD card bay or SD card reader */
	switch params.Connection {
	case camera.Connect:
		return ImportConnect(params)
	case camera.SDCard:
		break
	default:
		return nil, mErrors.ErrUnsupportedConnection
	}

	versionFile := filepath.Join(params.Input, "MISC", fmt.Sprint(Version))

	_, err := os.Stat(versionFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, mErrors.ErrNoCameraDetected
		}

		return nil, fmt.Errorf("unable to find %s", versionFile)
	}

	versionContent, err := os.ReadFile(versionFile)
	if err != nil {
		return nil, err
	}

	gpVersion, err := readInfo(versionContent)
	if err != nil {
		return nil, err
	}

	di, err := disk.GetInfo(params.Input)
	if err != nil {
		return nil, err
	}

	percentage := (float64(di.Total-di.Free) / float64(di.Total)) * 100

	c := color.New(color.FgCyan)
	y := color.New(color.FgHiBlue)

	color.Cyan("🎥 [%s]:", gpVersion.CameraType)
	_, _ = c.Printf("\t📹 FW: %s ", gpVersion.FirmwareVersion)
	_, _ = y.Printf("SN: %s\n", gpVersion.CameraSerialNumber)

	color.Cyan("\t💾 %s/%s (%0.2f%%)\n",
		humanize.Bytes(di.Total-di.Free),
		humanize.Bytes(di.Total),
		percentage,
	)

	root := strings.Split(gpVersion.FirmwareVersion, ".")[0]

	if params.CameraName == "" {
		params.CameraName = gpVersion.CameraType
	}

	if params.Prefix != "" {
		params.CameraName = fmt.Sprintf("%s %s", params.Prefix, params.CameraName)
	}

	params.Input = filepath.Join(params.Input, fmt.Sprint(DCIM))

	switch root {
	case "HD6", "HD7", "HD8", "H19", "HD9", "H21", "H22", "H23", "H24", "H26":
		result := importFromGoProV2(params)

		return &result, nil
	case "HD2", "HD3", "HD4", "HX", "HD5":
		result := importFromGoProV1(params)

		return &result, nil
	default:
		return nil, fmt.Errorf("camera %s is not supported", gpVersion.CameraType)
	}
}

func importFromGoProV2(params camera.ImportParams) camera.Result {
	fileTypes := FileTypeMatches[V2]

	var result camera.Result

	folders, err := os.ReadDir(params.Input)
	if err != nil {
		result.Errors = append(result.Errors, err)

		return result
	}

	var wg sync.WaitGroup

	sem := camera.NewSemaphore(params.MaxConcurrent)
	progressBar := mpb.New(mpb.WithWaitGroup(&wg),
		mpb.WithWidth(60),
		mpb.WithRefreshRate(180*time.Millisecond))

	inlineCounter := camera.ResultCounter{}

folderLoop:
	for _, f := range folders {
		r := MediaFolderRegex.MatchString(f.Name())

		if !r {
			continue folderLoop
		}

		color.Green("Looking at %s", f.Name())

		err = godirwalk.Walk(filepath.Join(params.Input, f.Name()), &godirwalk.Options{
			Callback: func(osPathname string, de *godirwalk.Dirent) error {
			fileTypeLoop:
				for _, ftype := range fileTypes {
					if !ftype.Regex.MatchString(de.Name()) {
						continue fileTypeLoop
					}

					d := getFileTime(osPathname)
					mediaDate := camera.FormatMediaDate(d, params.DateFormat)

					if d.Before(params.DateRange[0]) || d.After(params.DateRange[1]) {
						return godirwalk.SkipThis
					}

					info, err := os.Stat(osPathname)
					if err != nil {
						return godirwalk.SkipThis
					}

					dayFolder := camera.GetOrder(params.Sort, locationService, osPathname, params.Output, mediaDate, params.CameraName)

					wg.Add(1)

					bar := camera.GetNewBar(progressBar, info.Size(), de.Name(), camera.IoTX)

					switch ftype.Type {
					case Video:
						x := de.Name()
						filename := fmt.Sprintf("%s%s-%s%s", x[:2], x[4:][:4], x[2:][:2], filepath.Ext(x))

						rfpsFolder, err := getRfpsFolder(osPathname)
						if err != nil {
							inlineCounter.SetFailure(err, filename)
							bar.Abort(true)
							wg.Done()

							return godirwalk.SkipThis
						}

						additionalDir := ""
						if !ftype.HeroMode {
							additionalDir = "360"
						}

						hilights, hilightErr := GetHiLights(osPathname)
						if hilightErr == nil {
							durationResp, durationErr := ffprobe.Duration(osPathname)
							if durationErr == nil {
								additionalDir = filepath.Join(additionalDir, getImportanceName(hilights.Timestamps, int(durationResp.Streams[0].Duration), params.TagNames))
							}
						}

						folder := filepath.Join(dayFolder, "videos", additionalDir, rfpsFolder)

						sem <- struct{}{}

						go func(folder, filename, osPathname string, bar *mpb.Bar) {
							defer func() { <-sem }()
							defer wg.Done()

							err := parse(folder, filename, osPathname, params.BufferSize, bar, d)
							if err != nil {
								inlineCounter.SetFailure(err, filename)
							} else {
								inlineCounter.SetSuccess()
							}
						}(folder, filename, osPathname, bar)

						// Get LRV
						if params.SkipAuxiliaryFiles {
							return godirwalk.SkipThis
						}

						wg.Add(1)

						folder = filepath.Join(dayFolder, "videos/proxy", rfpsFolder)
						lrvReplacer := strings.NewReplacer("GX", "GL", "GH", "GL", "GM", "GL", "MP4", "LRV")
						lrvFullpath := filepath.Join(filepath.Dir(osPathname), lrvReplacer.Replace(de.Name()))

						lrvStat, err := os.Stat(lrvFullpath)
						if err != nil {
							wg.Done()

							return godirwalk.SkipThis
						}

						proxyVideoBar := camera.GetNewBar(progressBar, lrvStat.Size(), lrvReplacer.Replace(de.Name()), camera.IoTX)

						sem <- struct{}{}

						go func(folder, filename, osPathname string, bar *mpb.Bar) {
							defer func() { <-sem }()
							defer wg.Done()

							_ = parse(folder, filename, osPathname, params.BufferSize, bar, d)
						}(folder, filename, lrvFullpath, proxyVideoBar)
					case Photo:
						additionalDir := ""
						if !ftype.HeroMode {
							additionalDir = "360"
						}

						folder := filepath.Join(dayFolder, "photos", additionalDir)

						sem <- struct{}{}

						go func(folder, filename, osPathname string, bar *mpb.Bar) {
							defer func() { <-sem }()
							defer wg.Done()

							err := parse(folder, filename, osPathname, params.BufferSize, bar, d)
							if err != nil {
								inlineCounter.SetFailure(err, filename)
							} else {
								inlineCounter.SetSuccess()
							}
						}(folder, de.Name(), osPathname, bar)

					case Multishot:
						additionalDir := ""
						if !ftype.HeroMode {
							additionalDir = "360"
						}

						folder := filepath.Join(dayFolder, "multishot", additionalDir, de.Name()[:4])

						sem <- struct{}{}

						go func(folder, filename, osPathname string, bar *mpb.Bar) {
							defer func() { <-sem }()
							defer wg.Done()

							err := parse(folder, filename, osPathname, params.BufferSize, bar, d)
							if err != nil {
								inlineCounter.SetFailure(err, filename)
							} else {
								inlineCounter.SetSuccess()
							}
						}(folder, de.Name(), osPathname, bar)

					case RawPhoto:
						folder := filepath.Join(dayFolder, "photos/raw")

						sem <- struct{}{}

						go func(folder, filename, osPathname string, bar *mpb.Bar) {
							defer func() { <-sem }()
							defer wg.Done()

							err := parse(folder, filename, osPathname, params.BufferSize, bar, d)
							if err != nil {
								inlineCounter.SetFailure(err, filename)
							} else {
								inlineCounter.SetSuccess()
							}
						}(folder, de.Name(), osPathname, bar)

					case Audio:
						folder := filepath.Join(dayFolder, "audios")

						sem <- struct{}{}

						go func(folder, filename, osPathname string, bar *mpb.Bar) {
							defer func() { <-sem }()
							defer wg.Done()

							err := parse(folder, filename, osPathname, params.BufferSize, bar, d)
							if err != nil {
								inlineCounter.SetFailure(err, filename)
							} else {
								inlineCounter.SetSuccess()
							}
						}(folder, de.Name(), osPathname, bar)

					default:
						inlineCounter.SetFailure(errors.New("unsupported file"), de.Name())
					}
				}

				return nil
			},
			Unsorted: true,
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

	return result
}

func importFromGoProV1(params camera.ImportParams) camera.Result {
	fileTypes := FileTypeMatches[V1]

	var result camera.Result

	folders, err := os.ReadDir(params.Input)
	if err != nil {
		result.Errors = append(result.Errors, err)

		return result
	}

	var wg sync.WaitGroup

	sem := camera.NewSemaphore(params.MaxConcurrent)
	progressBar := mpb.New(mpb.WithWaitGroup(&wg),
		mpb.WithWidth(60),
		mpb.WithRefreshRate(180*time.Millisecond))

	inlineCounter := camera.ResultCounter{}

	for _, f := range folders {
		r := MediaFolderRegex.MatchString(f.Name())

		if !r {
			continue
		}

		color.Green("Looking at %s", f.Name())

		err = godirwalk.Walk(filepath.Join(params.Input, f.Name()), &godirwalk.Options{
			Callback: func(osPathname string, de *godirwalk.Dirent) error {
				for _, ftype := range fileTypes {
					if !ftype.Regex.MatchString(de.Name()) {
						continue
					}

					d := getFileTime(osPathname)
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
					case Video:
						x := de.Name()

						chaptered := regexp.MustCompile(`GP\d+.MP4`)
						if chaptered.MatchString(de.Name()) {
							x = fmt.Sprintf("GOPR%s%s.%s", x[4:][:4], x[2:][:2], strings.Split(x, ".")[1])
						}

						s, err := ffprobe.VideoSize(osPathname)
						if err != nil {
							inlineCounter.SetFailure(err, de.Name())
							bar.Abort(true)
							wg.Done()

							return godirwalk.SkipThis
						}

						framerate := strings.ReplaceAll(s.Streams[0].RFrameRate, "/1", "")
						rfpsFolder := fmt.Sprintf("%dx%d %s", s.Streams[0].Width, s.Streams[0].Height, framerate)

						additionalDir := ""

						hilights, hilightErr := GetHiLights(osPathname)
						if hilightErr == nil {
							durationResp, durationErr := ffprobe.Duration(osPathname)
							if durationErr == nil {
								additionalDir = filepath.Join(additionalDir, getImportanceName(hilights.Timestamps, int(durationResp.Streams[0].Duration), params.TagNames))
							}
						}

						folder := filepath.Join(dayFolder, "videos", additionalDir, rfpsFolder)

						sem <- struct{}{}

						go func(folder, filename, osPathname string, bar *mpb.Bar) {
							defer func() { <-sem }()
							defer wg.Done()

							err := parse(folder, filename, osPathname, params.BufferSize, bar, d)
							if err != nil {
								inlineCounter.SetFailure(err, filename)
							} else {
								inlineCounter.SetSuccess()
							}
						}(folder, x, osPathname, bar)

						if params.SkipAuxiliaryFiles {
							return godirwalk.SkipThis
						}

						wg.Add(1)

						folder = filepath.Join(dayFolder, "videos/proxy", rfpsFolder)
						lrvFullpath := filepath.Join(filepath.Dir(osPathname), strings.ReplaceAll(de.Name(), ".MP4", ".LRV"))

						lrvStat, err := os.Stat(lrvFullpath)
						if err != nil {
							wg.Done()

							return godirwalk.SkipThis
						}

						proxyVideoBar := camera.GetNewBar(progressBar, lrvStat.Size(), strings.ReplaceAll(de.Name(), ".MP4", ".LRV"), camera.IoTX)

						sem <- struct{}{}

						go func(folder, filename, osPathname string, bar *mpb.Bar) {
							defer func() { <-sem }()
							defer wg.Done()

							_ = parse(folder, filename, osPathname, params.BufferSize, bar, d)
						}(folder, x, lrvFullpath, proxyVideoBar)

					case ChapteredVideo:
						x := de.Name()
						name := fmt.Sprintf("GOPR%s%s.%s", x[4:][:4], x[2:][:2], strings.Split(x, ".")[1])

						s, err := ffprobe.VideoSize(osPathname)
						if err != nil {
							inlineCounter.SetFailure(err, de.Name())
							bar.Abort(true)
							wg.Done()

							return godirwalk.SkipThis
						}

						framerate := strings.ReplaceAll(s.Streams[0].RFrameRate, "/1", "")
						rfpsFolder := fmt.Sprintf("%dx%d %s", s.Streams[0].Width, s.Streams[0].Height, framerate)

						additionalDir := ""

						hilights, hilightErr := GetHiLights(osPathname)
						if hilightErr == nil {
							durationResp, durationErr := ffprobe.Duration(osPathname)
							if durationErr == nil {
								additionalDir = filepath.Join(additionalDir, getImportanceName(hilights.Timestamps, int(durationResp.Streams[0].Duration), params.TagNames))
							}
						}

						folder := filepath.Join(dayFolder, "videos", additionalDir, rfpsFolder)

						sem <- struct{}{}

						go func(folder, filename, osPathname string, bar *mpb.Bar) {
							defer func() { <-sem }()
							defer wg.Done()

							err := parse(folder, filename, osPathname, params.BufferSize, bar, d)
							if err != nil {
								inlineCounter.SetFailure(err, filename)
							} else {
								inlineCounter.SetSuccess()
							}
						}(folder, name, osPathname, bar)

						if params.SkipAuxiliaryFiles {
							return godirwalk.SkipThis
						}

						wg.Add(1)

						folder = filepath.Join(dayFolder, "videos/proxy", rfpsFolder)
						lrvFullpath := filepath.Join(filepath.Dir(osPathname), strings.ReplaceAll(de.Name(), ".MP4", ".LRV"))

						lrvStat, err := os.Stat(lrvFullpath)
						if err != nil {
							wg.Done()

							return godirwalk.SkipThis
						}

						proxyVideoBar := camera.GetNewBar(progressBar, lrvStat.Size(), strings.ReplaceAll(de.Name(), ".MP4", ".LRV"), camera.IoTX)

						sem <- struct{}{}

						go func(folder, filename, osPathname string, bar *mpb.Bar) {
							defer func() { <-sem }()
							defer wg.Done()

							_ = parse(folder, filename, osPathname, params.BufferSize, bar, d)
						}(folder, x, lrvFullpath, proxyVideoBar)
					case Photo:
						folder := filepath.Join(dayFolder, "photos")

						sem <- struct{}{}

						go func(folder, filename, osPathname string, bar *mpb.Bar) {
							defer func() { <-sem }()
							defer wg.Done()

							err := parse(folder, filename, osPathname, params.BufferSize, bar, d)
							if err != nil {
								inlineCounter.SetFailure(err, filename)
							} else {
								inlineCounter.SetSuccess()
							}
						}(folder, de.Name(), osPathname, bar)

					case LowResolutionVideo:
						if params.SkipAuxiliaryFiles {
							return godirwalk.SkipThis
						}

						folder := filepath.Join(dayFolder, "videos/proxy")

						sem <- struct{}{}

						go func(folder, filename, osPathname string, bar *mpb.Bar) {
							defer func() { <-sem }()
							defer wg.Done()

							err := parse(folder, filename, osPathname, params.BufferSize, bar, d)
							if err != nil {
								inlineCounter.SetFailure(err, filename)
							} else {
								inlineCounter.SetSuccess()
							}
						}(folder, de.Name(), osPathname, bar)

					case Multishot:
						folder := filepath.Join(dayFolder, "multishot", de.Name()[:4])

						sem <- struct{}{}

						go func(folder, filename, osPathname string, bar *mpb.Bar) {
							defer func() { <-sem }()
							defer wg.Done()

							err := parse(folder, filename, osPathname, params.BufferSize, bar, d)
							if err != nil {
								inlineCounter.SetFailure(err, filename)
							} else {
								inlineCounter.SetSuccess()
							}
						}(folder, de.Name(), osPathname, bar)

					case RawPhoto:
						folder := filepath.Join(dayFolder, "photos/raw")

						sem <- struct{}{}

						go func(folder, filename, osPathname string, bar *mpb.Bar) {
							defer func() { <-sem }()
							defer wg.Done()

							err := parse(folder, filename, osPathname, params.BufferSize, bar, d)
							if err != nil {
								inlineCounter.SetFailure(err, filename)
							} else {
								inlineCounter.SetSuccess()
							}
						}(folder, de.Name(), osPathname, bar)

					default:
						inlineCounter.SetFailure(errors.New("unsupported file"), de.Name())
					}
				}

				return nil
			},
			Unsorted: true,
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

	return result
}

/*
GoPro adds a trailing comma to their version.txt file... this removes it.
*/
func cleanVersion(s string) string {
	i := strings.LastIndex(s, ",")
	excludingLast := s[:i] + strings.Replace(s[i:], ",", "", 1)

	if strings.Contains(s, `,"firmware version"`) {
		return strings.ReplaceAll(s, "\n", "")
	}

	return excludingLast
}

func readInfo(inBytes []byte) (*Info, error) {
	text := string(inBytes)
	clean := cleanVersion(text)

	var gpVersion Info

	err := json.Unmarshal([]byte(clean), &gpVersion)
	if err != nil {
		return nil, err
	}

	return &gpVersion, nil
}

func getFileTime(osPathname string) time.Time {
	t, err := times.Stat(osPathname)
	if err != nil {
		return time.Time{}
	}

	return camera.ModTimeAsUTC(t.ModTime())
}

func parse(folder string, name string, osPathname string, bufferSize int, bar *mpb.Bar, modTime time.Time) error {
	err := os.MkdirAll(folder, 0o755)
	if err != nil {
		return err
	}

	sourceFileStat, err := os.Stat(osPathname)
	if err != nil {
		return err
	}

	err = utils.CopyFile(osPathname, filepath.Join(folder, name), bufferSize, bar, modTime)
	if err != nil {
		bar.EwmaSetCurrent(sourceFileStat.Size(), 1*time.Millisecond)
		bar.EwmaIncrInt64(sourceFileStat.Size(), 1*time.Millisecond)

		return err
	}

	return nil
}

// getModDates recursively collects unique modification dates from a directory tree.
func getModDates(input string) ([]time.Time, error) {
	modificationDates := []time.Time{}

	items, err := os.ReadDir(input)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if item.IsDir() {
			m, err := getModDates(filepath.Join(input, item.Name()))
			if err != nil {
				return nil, err
			}

			modificationDates = append(modificationDates, m...)
		} else {
			fileInfo, err := item.Info()
			if err != nil {
				return nil, err
			}

			fileDate := fileInfo.ModTime()
			parsedDate := time.Date(fileDate.Year(), fileDate.Month(), fileDate.Day(), 0, 0, 0, 0, fileDate.Location())
			found := false

			for _, d := range modificationDates {
				if d.Equal(parsedDate) {
					found = true

					break
				}
			}

			if !found {
				modificationDates = append(modificationDates, parsedDate)
			}
		}
	}

	return modificationDates, nil
}

// CaptureDates reports the dates on which media was captured.
func (Entrypoint) CaptureDates(input string, conn camera.ConnectionType) ([]time.Time, error) {
	switch conn {
	case camera.Connect:
		mediaList, err := GetMediaList(input)
		if err != nil {
			return nil, err
		}

		var dates []time.Time

		for _, folder := range mediaList.Media {
			for _, file := range folder.Fs {
				fileDate := time.Unix(file.Cre, 0)
				parsedDate := time.Date(fileDate.Year(), fileDate.Month(), fileDate.Day(), 0, 0, 0, 0, fileDate.Location())
				found := false

				for _, d := range dates {
					if d.Equal(parsedDate) {
						found = true

						break
					}
				}

				if !found {
					dates = append(dates, parsedDate)
				}
			}
		}

		return dates, nil
	case camera.SDCard:
		return getModDates(filepath.Join(input, string(DCIM)))
	}

	return nil, fmt.Errorf("unsupported connection type %q", conn)
}
