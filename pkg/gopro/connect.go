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

func getNewBar(progressBar *mpb.Progress, total int64, filename string) *mpb.Bar {
	return progressBar.AddBar(total,
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

func forceGetFolder(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}

type ResultCounter struct {
	mu               sync.Mutex
	Errors           []error
	FilesNotImported []string
	FilesImported    int
}

func (rc *ResultCounter) SetFailure(err error, file string) {
	rc.mu.Lock()
	rc.Errors = append(rc.Errors, err)
	rc.FilesNotImported = append(rc.FilesNotImported, file)
	rc.mu.Unlock()
}

func (rc *ResultCounter) SetSuccess() {
	rc.mu.Lock()
	rc.FilesImported++
	rc.mu.Unlock()
}

func (rc *ResultCounter) Get() utils.Result {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	return utils.Result{
		FilesImported:    rc.FilesImported,
		FilesNotImported: rc.FilesNotImported,
		Errors:           rc.Errors,
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

	inlineCounter := ResultCounter{}

	unsorted := filepath.Join(out, "unsorted")
	if _, err := os.Stat(unsorted); os.IsNotExist(err) {
		_ = os.Mkdir(unsorted, 0755)
	}
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

				if tm.Before(start) {
					continue
				}
				if tm.After(end) {
					continue
				}

				total, err := head(fmt.Sprintf("http://%s:8080/videos/DCIM/%s/%s", in, folder.D, goprofile.N))
				if err != nil {
					log.Fatal(err.Error())
				}

				wg.Add(1)
				bar := getNewBar(progressBar, int64(total), goprofile.N)

				switch fileTypeMatch.Type {
				case Video, ChapteredVideo:

					go func(in, folder, origFilename, unsorted string, result utils.Result) {
						defer wg.Done()
						x := origFilename
						filename := fmt.Sprintf("%s%s-%s.%s", x[:2], x[4:][:4], x[2:][:2], strings.Split(x, ".")[1])

						err := utils.DownloadFile(
							filepath.Join(unsorted, origFilename),
							fmt.Sprintf("http://%s:8080/videos/DCIM/%s/%s", in, folder, origFilename),
							bar)
						if err != nil {
							inlineCounter.SetFailure(err, origFilename)
						} else {
							inlineCounter.SetSuccess()

							// Move to actual folder

							finalPath := utils.GetOrder(sortOptions, locationService, filepath.Join(unsorted, origFilename), out, mediaDate, cameraName)

							var gpFileInfo = &goProMediaMetadata{}
							err = caller(in, fmt.Sprintf("gp/gpMediaMetadata?p=%s/%s&t=v4info", folder, origFilename), gpFileInfo)
							if err != nil {
								log.Fatal(err.Error())
							}

							framerate := gpFileInfo.Fps / gpFileInfo.FpsDenom
							if framerate == 0 {
								framerate = (gpFileInfo.FpsDenom / gpFileInfo.Fps)
							}

							rfpsFolder := fmt.Sprintf("%sx%s %d", gpFileInfo.W, gpFileInfo.H, framerate)

							forceGetFolder(filepath.Join(finalPath, "videos", rfpsFolder))

							err = os.Rename(
								filepath.Join(unsorted, origFilename),
								filepath.Join(finalPath, "videos", rfpsFolder, filename),
							)
							if err != nil {
								log.Fatal(err.Error())
							}
						}
					}(in, folder.D, goprofile.N, unsorted, result)

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
							log.Fatal(err.Error())
						}

						rawPhotoBar := getNewBar(progressBar, int64(rawPhotoTotal), rawPhotoName)
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

								err = os.Rename(
									filepath.Join(unsorted, nowPhoto.Name),
									filepath.Join(photoPath, nowPhoto.Name),
								)
								if err != nil {
									log.Fatal(err.Error())
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

						multiShotTotal, err := head(fmt.Sprintf("http://%s:8080/videos/DCIM/%s/%s", in, folder.D, filename))
						if err != nil {
							log.Fatal(err.Error())
						}
						multiShotBar := getNewBar(progressBar, int64(multiShotTotal), filename)

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

								err = os.Rename(
									filepath.Join(unsorted, origFilename),
									filepath.Join(finalPath, "multishot", filebaseroot, origFilename),
								)
								if err != nil {
									log.Fatal(err.Error())
								}
							}
						}(in, folder.D, filename, unsorted, result)
					}

				default:
					color.Red("Unsupported file %s", goprofile.N)
					result.Errors = append(result.Errors, errors.New("Media format unrecognized"))
					result.FilesNotImported = append(result.FilesNotImported, goprofile.N)
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
	result.Errors = append(result.Errors, inlineCounter.Get().Errors...)
	result.FilesImported += inlineCounter.Get().FilesImported
	result.FilesNotImported = append(result.FilesNotImported, inlineCounter.Get().FilesNotImported...)

	// cleanup
	os.Remove(unsorted)
	return &result, nil
}
