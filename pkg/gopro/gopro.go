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
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/karrick/godirwalk"
	mErrors "github.com/konradit/mmt/pkg/errors"
	"github.com/konradit/mmt/pkg/media"
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
	fpsAsFloat := strconv.Itoa(framerate.(int))
	return fmt.Sprintf("%dx%d %s", s.Streams[0].Width, s.Streams[0].Height, fpsAsFloat), nil
}

type Entrypoint struct{}

func (Entrypoint) Import(params utils.ImportParams) (*utils.Result, error) {
	/* Import method using SD card bay or SD card reader */

	switch params.Connection {
	case utils.Connect:
		return ImportConnect(params)
	case utils.SDCard:
		break
	default:
		return nil, mErrors.ErrUnsupportedConnection
	}

	versionFile := filepath.Join(params.Input, "MISC", fmt.Sprint(Version))

	_, err := os.Stat(versionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, mErrors.ErrNoCameraDetected
		}
		return nil, mErrors.ErrNotFound(versionFile)
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
	c.Printf("\t📹 FW: %s ", gpVersion.FirmwareVersion)
	y.Printf("SN: %s\n", gpVersion.CameraSerialNumber)
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
	case "HD6", "HD7", "HD8", "H19", "HD9", "H21", "H22":
		result := importFromGoProV2(params)
		return &result, nil
	case "HD2", "HD3", "HD4", "HX", "HD5":
		result := importFromGoProV1(params)
		return &result, nil
	default:
		return nil, mErrors.ErrUnsupportedCamera(gpVersion.CameraType)
	}
}

func importFromGoProV2(params utils.ImportParams) utils.Result {
	fileTypes := FileTypeMatches[V2]
	var result utils.Result

	folders, err := ioutil.ReadDir(params.Input)
	if err != nil {
		result.Errors = append(result.Errors, err)
		return result
	}

	var wg sync.WaitGroup
	progressBar := mpb.New(mpb.WithWaitGroup(&wg),
		mpb.WithWidth(60),
		mpb.WithRefreshRate(180*time.Millisecond))

	inlineCounter := utils.ResultCounter{}

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

					d := media.GetFileTime(osPathname, true)
					mediaDate := media.GetMediaDate(getFileTime(osPathname, true), params.DateFormat)

					if d.Before(params.DateRange[0]) || d.After(params.DateRange[1]) {
						return godirwalk.SkipThis
					}

					info, err := os.Stat(osPathname)
					if err != nil {
						return godirwalk.SkipThis
					}

					dayFolder := utils.GetOrder(params.Sort, locationService, osPathname, params.Output, mediaDate, params.CameraName)

					wg.Add(1)
					bar := utils.GetNewBar(progressBar, info.Size(), de.Name(), utils.IoTX)

					switch ftype.Type {
					case Video:
						x := de.Name()
						filename := fmt.Sprintf("%s%s-%s%s", x[:2], x[4:][:4], x[2:][:2], filepath.Ext(x))
						rfpsFolder, err := getRfpsFolder(osPathname)
						if err != nil {
							return godirwalk.SkipThis
						}
						additionalDir := ""
						if !ftype.HeroMode {
							additionalDir = "360"
						}

						if hilights, err := GetHiLights(osPathname); err == nil {
							if durationResp, err := ffprobe.Duration(osPathname); err == nil {
								additionalDir = filepath.Join(additionalDir, getImportanceName(hilights.Timestamps, int(durationResp.Streams[0].Duration), params.TagNames))
							}
						}
						folder := filepath.Join(dayFolder, "videos", additionalDir, rfpsFolder)
						go func(folder, filename, osPathname string, bar *mpb.Bar) {
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
							return godirwalk.SkipThis
						}
						proxyVideoBar := utils.GetNewBar(progressBar, lrvStat.Size(), lrvReplacer.Replace(de.Name()), utils.IoTX)

						go func(folder, filename, osPathname string, bar *mpb.Bar) {
							defer wg.Done()
							_ = parse(folder, filename, osPathname, params.BufferSize, bar, d)
						}(folder, filename, lrvFullpath, proxyVideoBar)
					case Photo:
						additionalDir := ""
						if !ftype.HeroMode {
							additionalDir = "360"
						}
						folder := filepath.Join(dayFolder, "photos", additionalDir)
						go func(folder, filename, osPathname string, bar *mpb.Bar) {
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
						go func(folder, filename, osPathname string, bar *mpb.Bar) {
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
						go func(folder, filename, osPathname string, bar *mpb.Bar) {
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
						go func(folder, filename, osPathname string, bar *mpb.Bar) {
							defer wg.Done()
							err := parse(folder, filename, osPathname, params.BufferSize, bar, d)
							if err != nil {
								inlineCounter.SetFailure(err, filename)
							} else {
								inlineCounter.SetSuccess()
							}
						}(folder, de.Name(), osPathname, bar)

					default:
						inlineCounter.SetFailure(errors.New("Unsupported file"), de.Name())
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

func importFromGoProV1(params utils.ImportParams) utils.Result {
	fileTypes := FileTypeMatches[V1]
	var result utils.Result

	folders, err := ioutil.ReadDir(params.Input)
	if err != nil {
		result.Errors = append(result.Errors, err)
		return result
	}

	var wg sync.WaitGroup
	progressBar := mpb.New(mpb.WithWaitGroup(&wg),
		mpb.WithWidth(60),
		mpb.WithRefreshRate(180*time.Millisecond))

	inlineCounter := utils.ResultCounter{}

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

					d := media.GetFileTime(osPathname, true)
					mediaDate := media.GetMediaDate(d, params.DateFormat)

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

						additionalDir := ""
						if hilights, err := GetHiLights(osPathname); err == nil {
							if durationResp, err := ffprobe.Duration(osPathname); err == nil {
								additionalDir = filepath.Join(additionalDir, getImportanceName(hilights.Timestamps, int(durationResp.Streams[0].Duration), params.TagNames))
							}
						}

						folder := filepath.Join(dayFolder, "videos", additionalDir, rfpsFolder)
						go func(folder, filename, osPathname string, bar *mpb.Bar) {
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
						lrvFullpath := filepath.Join(filepath.Dir(osPathname), strings.Replace(de.Name(), ".MP4", ".LRV", -1))
						lrvStat, err := os.Stat(lrvFullpath)
						if err != nil {
							return godirwalk.SkipThis
						}
						proxyVideoBar := utils.GetNewBar(progressBar, lrvStat.Size(), strings.Replace(de.Name(), ".MP4", ".LRV", -1), utils.IoTX)

						go func(folder, filename, osPathname string, bar *mpb.Bar) {
							defer wg.Done()
							_ = parse(folder, filename, osPathname, params.BufferSize, bar, d)
						}(folder, x, lrvFullpath, proxyVideoBar)

					case ChapteredVideo:
						x := de.Name()
						name := fmt.Sprintf("GOPR%s%s.%s", x[4:][:4], x[2:][:2], strings.Split(x, ".")[1])
						s, err := ffprobe.VideoSize(osPathname)
						if err != nil {
							log.Fatal(err.Error())
							return godirwalk.SkipThis
						}
						framerate := strings.ReplaceAll(s.Streams[0].RFrameRate, "/1", "")
						rfpsFolder := fmt.Sprintf("%dx%d %s", s.Streams[0].Width, s.Streams[0].Height, framerate)

						additionalDir := ""
						if hilights, err := GetHiLights(osPathname); err == nil {
							if durationResp, err := ffprobe.Duration(osPathname); err == nil {
								additionalDir = filepath.Join(additionalDir, getImportanceName(hilights.Timestamps, int(durationResp.Streams[0].Duration), params.TagNames))
							}
						}

						folder := filepath.Join(dayFolder, "videos", additionalDir, rfpsFolder)
						go func(folder, filename, osPathname string, bar *mpb.Bar) {
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
						lrvFullpath := filepath.Join(filepath.Dir(osPathname), strings.Replace(de.Name(), ".MP4", ".LRV", -1))
						lrvStat, err := os.Stat(lrvFullpath)
						if err != nil {
							return godirwalk.SkipThis
						}
						proxyVideoBar := utils.GetNewBar(progressBar, lrvStat.Size(), strings.Replace(de.Name(), ".MP4", ".LRV", -1), utils.IoTX)

						go func(folder, filename, osPathname string, bar *mpb.Bar) {
							defer wg.Done()
							_ = parse(folder, filename, osPathname, params.BufferSize, bar, d)
						}(folder, x, lrvFullpath, proxyVideoBar)
					case Photo:
						folder := filepath.Join(dayFolder, "photos")
						go func(folder, filename, osPathname string, bar *mpb.Bar) {
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
						go func(folder, filename, osPathname string, bar *mpb.Bar) {
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
						go func(folder, filename, osPathname string, bar *mpb.Bar) {
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
						go func(folder, filename, osPathname string, bar *mpb.Bar) {
							defer wg.Done()
							err := parse(folder, filename, osPathname, params.BufferSize, bar, d)
							if err != nil {
								inlineCounter.SetFailure(err, filename)
							} else {
								inlineCounter.SetSuccess()
							}
						}(folder, de.Name(), osPathname, bar)

					default:
						inlineCounter.SetFailure(errors.New("Unsupported file"), de.Name())
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

	if strings.Contains(s, "HERO10") || strings.Contains(s, "HERO11") {
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

func getMediaDate(d time.Time, dateFormat string) string {
	mediaDate := d.Format("02-01-2006")
	if strings.Contains(dateFormat, "yyyy") && strings.Contains(dateFormat, "mm") && strings.Contains(dateFormat, "dd") {
		mediaDate = d.Format(utils.DateFormatReplacer.Replace(dateFormat))
	}
	return mediaDate
}

func parse(folder string, name string, osPathname string, bufferSize int, bar *mpb.Bar, modTime time.Time) error {
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		mkdirerr := os.MkdirAll(folder, 0o755)
		if mkdirerr != nil {
			return mkdirerr
		}
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
