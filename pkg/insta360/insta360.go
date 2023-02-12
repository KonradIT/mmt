package insta360

import (
	"bytes"
	"fmt"
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
	"github.com/vbauerster/mpb/v8"
	"gopkg.in/djherbis/times.v1"
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

func Import(in, out, dateFormat string, bufferSize int, prefix string, dateRange []string, cameraName string, cameraOptions map[string]interface{}) (*utils.Result, error) {
	sortOptions := utils.ParseCliOptions(cameraOptions)

	if cameraName == "" {
		cameraName = getDeviceName(filepath.Join(in, "DCIM", "fileinfo_list.list"))
	}
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

	mediaFolder := `Camera\d+`
	mediaFolderRegex := regexp.MustCompile(mediaFolder)

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

					switch len(dateRange) {
					case 1:
						dateStart := time.Date(0o000, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
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
					case 2:
						layout := replacer.Replace(dateFormat)

						start, err := time.Parse(layout, dateRange[0])
						if err != nil {
							log.Fatal(err.Error())
						}
						end, err := time.Parse(layout, dateRange[1])
						if err != nil {
							log.Fatal(err.Error())
						}
						if d.Before(start) || d.After(end) {
							return godirwalk.SkipThis
						}
					}

					info, err := os.Stat(osPathname)
					if err != nil {
						return godirwalk.SkipThis
					}

					wg.Add(1)
					bar := utils.GetNewBar(progressBar, info.Size(), de.Name(), utils.IoTX)
					dayFolder := utils.GetOrder(sortOptions, nil, osPathname, out, mediaDate, cameraName)

					x := de.Name()

					switch ftype.Type {
					case Photo, RawPhoto:
						id := x[3+8+2 : 3+8+6+2]
						if _, err := os.Stat(filepath.Join(dayFolder, "photos", id)); os.IsNotExist(err) {
							err = os.MkdirAll(filepath.Join(dayFolder, "photos", id), 0o755)
							if err != nil {
								log.Fatal(err.Error())
							}
						}

						go func(id, filename, osPathname string, bar *mpb.Bar) {
							defer wg.Done()

							err = utils.CopyFile(osPathname, filepath.Join(dayFolder, "photos", id, x), bufferSize, bar)
							if err != nil {
								bar.EwmaSetCurrent(info.Size(), 1*time.Millisecond)
								bar.EwmaIncrInt64(info.Size(), 1*time.Millisecond)
								inlineCounter.SetFailure(err, filename)
							} else {
								inlineCounter.SetSuccess()
							}
						}(id, x, osPathname, bar)
					case Video, LowResolutionVideo:
						if sortOptions.SkipAuxiliaryFiles && ftype.Type == LowResolutionVideo {
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
							err = os.MkdirAll(filepath.Join(dayFolder, slug, id), 0o755)
							if err != nil {
								log.Fatal(err.Error())
							}
						}

						go func(id, filename, osPathname string, bar *mpb.Bar) {
							defer wg.Done()

							err = utils.CopyFile(osPathname, filepath.Join(dayFolder, slug, id, x), bufferSize, bar)
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
