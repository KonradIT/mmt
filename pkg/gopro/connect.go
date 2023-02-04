package gopro

/* GoPro Connect - API exposed over USB Ethernet */

import (
	"encoding/json"
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
	mErrors "github.com/konradit/mmt/pkg/errors"
	"github.com/konradit/mmt/pkg/utils"
	"github.com/vbauerster/mpb/v8"
)

var ipAddress = ""
var gpTurbo = true

func handleKill() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM) //nolint
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
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
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

func GetGoProNetworkAddresses() ([]ConnectDevice, error) {
	ipsFound := []ConnectDevice{}
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
				ipsFound = append(ipsFound, ConnectDevice{
					IP:   correctIP,
					Info: *gpInfo,
				})
			}
		}
	}
	return ipsFound, nil
}

func GetMediaList(in string) (*MediaList, error) {
	var gpMediaList = &MediaList{}
	err := caller(in, "gp/gpMediaList", gpMediaList)
	if err != nil {
		return nil, err
	}
	return gpMediaList, nil
}
func forceGetFolder(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}

func ImportConnect(in, out string, sortOptions utils.SortOptions) (*utils.Result, error) {
	var verType Type
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

	if gpTurbo {
		err = caller(in, "gp/gpTurbo?p=1", nil)
		if err != nil {
			color.Red("Error activating Turbo! Download speeds will be much slower")
		}
	}

	gpMediaList, err := GetMediaList(in)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	progressBar := mpb.New(mpb.WithWaitGroup(&wg),
		mpb.WithWidth(60),
		mpb.WithRefreshRate(180*time.Millisecond))

	inlineCounter := utils.ResultCounter{}

	unsorted := filepath.Join(out, "unsorted")
	if _, err := os.Stat(unsorted); os.IsNotExist(err) {
		_ = os.Mkdir(unsorted, 0755)
	}

	chaptered := regexp.MustCompile(`GP\d+.MP4`)

	for _, folder := range gpMediaList.Media {
		for _, goprofile := range folder.Fs {
			for _, fileTypeMatch := range FileTypeMatches[verType] {
				if !fileTypeMatch.Regex.MatchString(goprofile.N) {
					continue
				}
				i, err := strconv.ParseInt(goprofile.Mod, 10, 64)
				if err != nil {
					continue
				}
				tm := time.Unix(i, 0).UTC()
				start := sortOptions.DateRange[0]
				end := sortOptions.DateRange[1]
				zoneName, _ := end.Zone()
				newTime := strings.Replace(tm.Format(time.UnixDate), "UTC", zoneName, -1)
				tm, _ = time.Parse(time.UnixDate, newTime)
				mediaDate := tm.Format("02-01-2006")

				if strings.Contains(sortOptions.DateFormat, "yyyy") && strings.Contains(sortOptions.DateFormat, "mm") && strings.Contains(sortOptions.DateFormat, "dd") {
					mediaDate = tm.Format(replacer.Replace(sortOptions.DateFormat))
				}

				if tm.Before(start) || tm.After(end) {
					continue
				}

				wg.Add(1)
				bar := utils.GetNewBar(progressBar, goprofile.S, goprofile.N, utils.IoTX)

				switch fileTypeMatch.Type {
				case Video, ChapteredVideo:

					go func(in, folder, origFilename, unsorted string, lrvSize int, bar *mpb.Bar, result utils.Result) {
						defer wg.Done()
						x := origFilename
						filename := origFilename

						if verType == V2 {
							filename = fmt.Sprintf("%s%s-%s.%s", x[:2], x[4:][:4], x[2:][:2], "MP4")
						}
						if verType == V1 && chaptered.MatchString(x) {
							filename = fmt.Sprintf("GOPR%s%s.%s", x[4:][:4], x[2:][:2], "MP4")
						}

						err := utils.DownloadFile(
							filepath.Join(unsorted, origFilename),
							fmt.Sprintf("http://%s:8080/videos/DCIM/%s/%s", in, folder, origFilename),
							bar)
						if err != nil {
							inlineCounter.SetFailure(err, origFilename)
							return
						}

						inlineCounter.SetSuccess()

						// Move to actual folder

						finalPath := utils.GetOrder(sortOptions, locationService, filepath.Join(unsorted, origFilename), out, mediaDate, cameraName)
						var gpFileInfo = &goProMediaMetadata{}
						err = caller(in, fmt.Sprintf("gp/gpMediaMetadata?p=%s/%s&t=v4info", folder, origFilename), gpFileInfo)
						if err != nil {
							inlineCounter.SetFailure(err, origFilename)
							return
						}

						importanceName := getImportanceName(gpFileInfo.Hi, gpFileInfo.Dur, sortOptions.TagNames)

						denom := gpFileInfo.FpsDenom
						if denom == 0 {
							denom = 1
						}
						framerate := gpFileInfo.Fps / denom
						if framerate == 0 {
							framerate = (denom / gpFileInfo.Fps)
						}

						rfpsFolder := fmt.Sprintf("%sx%s %d", gpFileInfo.W, gpFileInfo.H, framerate)

						forceGetFolder(filepath.Join(finalPath, "videos", importanceName, rfpsFolder))

						err = os.Rename(
							filepath.Join(unsorted, origFilename),
							filepath.Join(finalPath, "videos", importanceName, rfpsFolder, filename),
						)
						if err != nil {
							inlineCounter.SetFailure(err, origFilename)
							return
						}

						// download proxy
						if lrvSize > 0 && !sortOptions.SkipAuxiliaryFiles {
							proxyVideoName := "GL" + strings.Replace(origFilename[2:], ".MP4", ".LRV", -1)
							if verType == V1 {
								proxyVideoName = strings.Replace(origFilename, ".MP4", ".LRV", -1)
							}

							proxyVideoBar := utils.GetNewBar(progressBar, int64(lrvSize), proxyVideoName, utils.IoTX)
							err := utils.DownloadFile(
								filepath.Join(unsorted, proxyVideoName),
								fmt.Sprintf("http://%s:8080/videos/DCIM/%s/%s", in, folder, proxyVideoName),
								proxyVideoBar)
							if err != nil {
								inlineCounter.SetFailure(err, origFilename)
								return
							}
							forceGetFolder(filepath.Join(finalPath, "videos", "proxy", rfpsFolder))

							err = os.Rename(
								filepath.Join(unsorted, proxyVideoName),
								filepath.Join(finalPath, "videos", "proxy", rfpsFolder, filename),
							)
							if err != nil {
								inlineCounter.SetFailure(err, origFilename)
								return
							}
							inlineCounter.SetSuccess()
						}
					}(in, folder.D, goprofile.N, unsorted, goprofile.Glrv, bar, result)

				case Photo:
					type photo struct {
						Name   string
						Folder string
						Size   int
						IsRaw  bool
						Bar    *mpb.Bar
					}
					totalPhotos := []photo{
						{
							Folder: folder.D,
							Name:   goprofile.N,
							IsRaw:  false,
							Bar:    bar},
					}

					hasRawPhoto := goprofile.Raw == "1"
					if hasRawPhoto {
						wg.Add(1)
						rawPhotoName := strings.Replace(goprofile.N, ".JPG", ".GPR", -1)

						rawPhotoTotal, err := head(fmt.Sprintf("http://%s:8080/videos/DCIM/%s/%s", in, folder.D, rawPhotoName))
						if err != nil {
							continue
						}

						rawPhotoBar := utils.GetNewBar(progressBar, int64(rawPhotoTotal), rawPhotoName, utils.IoTX)
						totalPhotos = append(totalPhotos, photo{
							Name:   rawPhotoName,
							Folder: folder.D,
							IsRaw:  true,
							Bar:    rawPhotoBar,
						})
					}

					for _, item := range totalPhotos {
						go func(in string, nowPhoto photo, unsorted string, result utils.Result) {
							defer wg.Done()

							err := utils.DownloadFile(
								filepath.Join(unsorted, nowPhoto.Name),
								fmt.Sprintf("http://%s:8080/videos/DCIM/%s/%s", in, nowPhoto.Folder, nowPhoto.Name),
								nowPhoto.Bar,
							)
							if err != nil {
								inlineCounter.SetFailure(err, nowPhoto.Name)
							} else {
								inlineCounter.SetSuccess()
								// Move to actual folder

								finalPath := utils.GetOrder(sortOptions, locationService, filepath.Join(unsorted, nowPhoto.Name), out, mediaDate, cameraName)

								photoPath := filepath.Join(finalPath, "photos")
								if nowPhoto.IsRaw {
									photoPath = filepath.Join(photoPath, "raw")
								}
								forceGetFolder(photoPath)

								err := os.Rename(
									filepath.Join(unsorted, nowPhoto.Name),
									filepath.Join(photoPath, nowPhoto.Name),
								)
								if err != nil {
									inlineCounter.SetFailure(err, nowPhoto.Name)
									return
								}
							}
						}(in, item, unsorted, result)
					}

				case Multishot:
					filebaseroot := goprofile.N[:4]

					for i := goprofile.B; i <= goprofile.L; i++ {
						if i > goprofile.B {
							wg.Add(1)
						}
						filename := fmt.Sprintf("%s%04d.JPG", filebaseroot, i)

						var gpFileInfo = &goProMediaMetadata{}
						err = caller(in, fmt.Sprintf("gp/gpMediaMetadata?p=%s/%s&t=v4info", folder.D, filename), gpFileInfo)
						if err != nil {
							log.Fatal(err.Error())
						}
						multiShotBar := utils.GetNewBar(progressBar, int64(gpFileInfo.S), filename, utils.IoTX)

						go func(in, folder, origFilename, unsorted string, result utils.Result) {
							defer wg.Done()

							err := utils.DownloadFile(
								filepath.Join(unsorted, origFilename),
								fmt.Sprintf("http://%s:8080/videos/DCIM/%s/%s", in, folder, origFilename),
								multiShotBar,
							)
							if err != nil {
								inlineCounter.SetFailure(err, origFilename)
							} else {
								inlineCounter.SetSuccess()
								// Move to actual folder
								finalPath := utils.GetOrder(sortOptions, locationService, filepath.Join(unsorted, origFilename), out, mediaDate, cameraName)
								forceGetFolder(filepath.Join(finalPath, "multishot", filebaseroot))

								err := os.Rename(
									filepath.Join(unsorted, origFilename),
									filepath.Join(finalPath, "multishot", filebaseroot, origFilename),
								)
								if err != nil {
									inlineCounter.SetFailure(err, origFilename)
									return
								}
							}
						}(in, folder.D, filename, unsorted, result)
					}

				default:
					color.Red("Unsupported file %s", goprofile.N)
					result.Errors = append(result.Errors, mErrors.ErrUnrecognizedMediaFormat)
					result.FilesNotImported = append(result.FilesNotImported, goprofile.N)
				}
			}
		}
	}

	wg.Wait()
	progressBar.Shutdown()
	if gpTurbo {
		if err := caller(in, "gp/gpTurbo?p=0", nil); err != nil {
			color.Red("Could not exit turbo mode")
		}
	}
	result.Errors = append(result.Errors, inlineCounter.Get().Errors...)
	result.FilesImported += inlineCounter.Get().FilesImported
	result.FilesNotImported = append(result.FilesNotImported, inlineCounter.Get().FilesNotImported...)

	// cleanup
	os.Remove(unsorted)
	return &result, nil
}
