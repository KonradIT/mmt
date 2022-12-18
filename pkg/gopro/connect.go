package gopro

/* GoPro Connect - API exposed over USB Ethernet */

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/konradit/mmt/pkg/utils"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

var ipAddress = ""
var gpTurbo = true

func handleKill() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM) //nolint:govet // todo
	go func() {
		<-c
		color.Red("\nKilling program, exiting Turbo mode.")
		if gpTurbo {
			if err := caller(ipAddress, "gp/gpTurbo?p=0", nil); err != nil {
				color.Red("Could not exit turbo mode")
			}
		}
		os.Exit(0)
	}()
}
func caller(ip, path string, object interface{}) error {

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/%s", ip, path), nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if object != nil {
		err = json.NewDecoder(resp.Body).Decode(object)
		if err != nil {
			return err
		}
	}
	return nil
}

func head(path string) (int, error) {

	client := &http.Client{}
	req, err := http.NewRequest("HEAD", path, nil)
	if err != nil {
		return 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	length, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return 0, err
	}
	return length, nil
}

func GetGoProNetworkAddresses() ([]GoProConnectDevice, error) {
	ipsFound := []GoProConnectDevice{}
	ifaces, err := net.Interfaces()
	if err != nil {
		return ipsFound, err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			r := regexp.MustCompile(`172.2\d.\d\d\d.5\d`)
			ipv4Addr := a.(*net.IPNet).IP.To4()
			if r.MatchString(ipv4Addr.String()) {
				correctIP := ipv4Addr.String()[:len(ipv4Addr.String())-1] + "1"
				var gpInfo = &cameraInfo{}
				err := caller(correctIP, "gp/gpControl/info", gpInfo)
				if err != nil {
					continue
				}
				ipsFound = append(ipsFound, GoProConnectDevice{
					IP:   correctIP,
					Info: *gpInfo,
				})
			}
		}
	}
	return ipsFound, nil
}

func getThumbnailFilename(filename string) string {
	replacer := strings.NewReplacer("H", "L", "X", "L", "MP4", "LRV")
	return replacer.Replace(filename)
}
func ImportConnect(in, out string, sortOptions SortOptions) (*utils.Result, error) {
	var verType GoProType
	var result utils.Result
	ipAddress = in

	// handle ctrl-c
	handleKill()

	var gpInfo = &cameraInfo{}
	err := caller(in, "gp/gpControl/info", gpInfo)
	if err != nil {
		return nil, err
	}
	cameraName := gpInfo.Info.ModelName

	root := strings.Split(gpInfo.Info.FirmwareVersion, ".")[0]

	switch root {
	case "HD9", "H21", "H22":
		verType = V2
		gpTurbo = true
	case "HD6", "HD7", "HD8":
		verType = V2
		gpTurbo = false
	default:
		verType = V1
		gpTurbo = false
	}
	// activate turbo

	if verType == V2 {
		err = caller(in, "gp/gpTurbo?p=1", nil)
		if err != nil {
			color.Red("Error activating Turbo! Download speeds will be much slower")
		}
	}

	var gpMediaList = &goProMediaList{}
	err = caller(in, "gp/gpMediaList", gpMediaList)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	progressBar := mpb.New(mpb.WithWaitGroup(&wg),
		mpb.WithWidth(60),
		mpb.WithRefreshRate(180*time.Millisecond))
	for _, folder := range gpMediaList.Media {
		for _, goprofile := range folder.Fs {
			for _, fileTypeMatch := range FileTypeMatches[verType] {

				if fileTypeMatch.Regex.MatchString(goprofile.N) {

					i, err := strconv.ParseInt(goprofile.Mod, 10, 64)
					if err != nil {
						continue
					}
					tm := time.Unix(i, 0)
					mediaDate := tm.Format("02-01-2006")

					if strings.Contains(sortOptions.DateFormat, "yyyy") && strings.Contains(sortOptions.DateFormat, "mm") && strings.Contains(sortOptions.DateFormat, "dd") {
						mediaDate = tm.Format(replacer.Replace(sortOptions.DateFormat))
					}

					start := sortOptions.DateRange[0]
					end := sortOptions.DateRange[1]
					if tm.Before(start) {
						continue
					}
					if tm.After(end) {
						continue
					}

					dayFolder := filepath.Join(out, mediaDate)
					if _, err := os.Stat(dayFolder); os.IsNotExist(err) {
						_ = os.Mkdir(dayFolder, 0755)
					}

					if sortOptions.ByCamera {
						if _, err := os.Stat(filepath.Join(dayFolder, cameraName)); os.IsNotExist(err) {
							_ = os.Mkdir(filepath.Join(dayFolder, cameraName), 0755)
						}
						dayFolder = filepath.Join(dayFolder, cameraName)
					}

					total, err := head(fmt.Sprintf("http://%s:8080/videos/DCIM/%s/%s", in, folder.D, goprofile.N))
					if err != nil {
						log.Fatal(err.Error())
					}

					wg.Add(1)
					bar := progressBar.AddBar(int64(total),
						mpb.PrependDecorators(
							decor.Name(color.CyanString(fmt.Sprintf("%s: ", goprofile.N))),
							decor.CountersKiloByte("% .2f / % .2f"),
						),
						mpb.AppendDecorators(
							decor.OnComplete(
								decor.EwmaETA(decor.ET_STYLE_GO, 60, decor.WCSyncWidth), "✔️",
							),
						),
					)

					switch fileTypeMatch.Type {
					case Video, ChapteredVideo:
						x := goprofile.N
						filename := fmt.Sprintf("%s%s-%s.%s", x[:2], x[4:][:4], x[2:][:2], strings.Split(x, ".")[1])

						var gpFileInfo = &goProMediaMetadata{}
						err = caller(in, fmt.Sprintf("gp/gpMediaMetadata?p=%s/%s&t=v4info", folder.D, goprofile.N), gpFileInfo)
						if err != nil {
							return nil, err
						}
						framerate := gpFileInfo.Fps / gpFileInfo.FpsDenom
						if framerate == 0 {
							framerate = (gpFileInfo.FpsDenom / gpFileInfo.Fps)

						}
						rfpsFolder := fmt.Sprintf("%sx%s %d", gpFileInfo.W, gpFileInfo.H, framerate)

						if _, err := os.Stat(filepath.Join(dayFolder, "videos", rfpsFolder)); os.IsNotExist(err) {
							err = os.MkdirAll(filepath.Join(dayFolder, "videos", rfpsFolder), 0755)
							if err != nil {
								log.Fatal(err.Error())
							}
						}

						var werr = make(chan error)
						go func(outfile string, path string, result utils.Result) {
							defer wg.Done()

							werr <- utils.DownloadFile(
								outfile,
								path,
								bar)

						}(filepath.Join(dayFolder, "videos", rfpsFolder, filename), fmt.Sprintf("http://%s:8080/videos/DCIM/%s/%s", in, folder.D, goprofile.N), result)

						err = <-werr
						if err != nil {
							result.Errors = append(result.Errors, err)
							result.FilesNotImported = append(result.FilesNotImported, filename)
						} else {
							result.FilesImported += 1
						}

					case Photo:
						if _, err := os.Stat(filepath.Join(dayFolder, "photos")); os.IsNotExist(err) {
							err = os.MkdirAll(filepath.Join(dayFolder, "photos"), 0755)
							if err != nil {
								log.Fatal(err.Error())
							}
						}

						var werr = make(chan error)
						go func(outfile, path string, result utils.Result) {
							defer wg.Done()

							werr <- utils.DownloadFile(
								outfile,
								path,
								bar,
							)

						}(filepath.Join(dayFolder, "photos", goprofile.N), fmt.Sprintf("http://%s:8080/videos/DCIM/%s/%s", in, folder.D, goprofile.N), result)
						err = <-werr
						if err != nil {
							result.Errors = append(result.Errors, err)
							result.FilesNotImported = append(result.FilesNotImported, goprofile.N)
						} else {
							result.FilesImported += 1
						}

					case Multishot:
						filebaseroot := goprofile.N[:4]
						if _, err := os.Stat(filepath.Join(dayFolder, "multishot", filebaseroot)); os.IsNotExist(err) {
							err = os.MkdirAll(filepath.Join(dayFolder, "multishot", filebaseroot), 0755)
							if err != nil {
								log.Fatal(err.Error())
							}
						}

						for i := goprofile.B; i <= goprofile.L; i++ {
							if i > goprofile.B {
								wg.Add(1)
							}
							filename := fmt.Sprintf("%s%04d.JPG", filebaseroot, i)

							multiShotTotal, err := head(fmt.Sprintf("http://%s:8080/videos/DCIM/%s/%s", in, folder.D, filename))
							if err != nil {
								log.Fatal(err.Error())
							}
							multiShotBar := progressBar.AddBar(int64(multiShotTotal),
								mpb.PrependDecorators(
									decor.Name(color.CyanString(fmt.Sprintf("%s: ", filename))),
									decor.CountersKiloByte("% .2f / % .2f"),
								),
								mpb.AppendDecorators(
									decor.OnComplete(
										decor.EwmaETA(decor.ET_STYLE_GO, 60, decor.WCSyncWidth), "✔️",
									),
								),
							)
							var werr = make(chan error)
							go func(outfile, path string, result utils.Result) {
								defer wg.Done()

								werr <- utils.DownloadFile(
									outfile, path,
									multiShotBar,
								)

							}(filepath.Join(dayFolder, "multishot", filebaseroot, filename),
								fmt.Sprintf("http://%s:8080/videos/DCIM/%s/%s", in, folder.D, filename), result)

							err = <-werr
							if err != nil {
								result.Errors = append(result.Errors, err)
								result.FilesNotImported = append(result.FilesNotImported, filename)
							} else {
								result.FilesImported += 1
							}
						}
					case RawPhoto:
						if _, err := os.Stat(filepath.Join(dayFolder, "photos/raw")); os.IsNotExist(err) {
							err = os.MkdirAll(filepath.Join(dayFolder, "photos/raw"), 0755)
							if err != nil {
								log.Fatal(err.Error())
							}
						}

						// convert to DNG here
						err := utils.DownloadFile(
							filepath.Join(dayFolder, "photos/raw", goprofile.N),
							fmt.Sprintf("http://%s:8080/videos/DCIM/%s/%s", in, folder.D, goprofile.N),
							bar,
						)
						if err != nil {
							result.Errors = append(result.Errors, err)
							result.FilesNotImported = append(result.FilesNotImported, goprofile.N)
						} else {
							result.FilesImported += 1
						}
					default:
						color.Red("Unsupported file %s", goprofile.N)
						result.Errors = append(result.Errors, errors.New("Media format unrecognized"))
						result.FilesNotImported = append(result.FilesNotImported, goprofile.N)
					}
				}
			}
		}
	}

	wg.Wait()
	progressBar.Shutdown()
	if verType == V2 {
		if err := caller(in, "gp/gpTurbo?p=0", nil); err != nil {
			color.Red("Could not exit turbo mode")
		}
	}

	return &result, nil
}
