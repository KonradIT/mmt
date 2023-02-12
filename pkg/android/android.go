package android

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/konradit/mmt/pkg/utils"
	"github.com/vbauerster/mpb/v8"
	adb "github.com/zach-klippenstein/goadb"
)

func pixelNameSort(filename string) (string, string) {
	if strings.Contains(filename, "MOTION") {
		// PXL_20211212_121243677.MOTION-01.COVER.jpg -- ok
		// PXL_20211212_121243677.MOTION-02.ORIGINAL.jpg -- ok
		// PXL_20211212_121307021.jpg -- ko
		s := strings.Split(filename, ".MOTION")
		return filename, s[0]
	}
	return filename, ""
}

var (
	locationService = LocationService{}
	replacer        = strings.NewReplacer("dd", "02", "mm", "01", "yyyy", "2006")
)

func prepare(out string, deviceFileName string, deviceModel string, mediaDate string, sortOptions utils.SortOptions, deviceFileReader io.ReadCloser, progressBar *mpb.Progress) (*mpb.Bar, string, error) {
	localFile, err := ioutil.TempFile(out, deviceFileName)
	if err != nil {
		return nil, "", err
	}

	_, err = io.Copy(localFile, deviceFileReader)
	if err != nil {
		return nil, "", err
	}

	stat, err := localFile.Stat()
	if err != nil {
		return nil, "", err
	}

	bar := utils.GetNewBar(progressBar, stat.Size(), deviceFileName, utils.IoTX)

	dayFolder := utils.GetOrder(sortOptions, locationService, filepath.Join(out, localFile.Name()), out, mediaDate, deviceModel)

	err = localFile.Close()
	if err != nil {
		return nil, "", err
	}
	err = os.Remove(filepath.Join(out, localFile.Name()))
	if err != nil {
		return nil, "", err
	}
	return bar, dayFolder, nil
}

func Import(in, out, dateFormat string, bufferSize int, prefix string, dateRange []string, cameraName string, cameraOptions map[string]interface{}) (*utils.Result, error) {
	var result utils.Result

	sortOptions := utils.ParseCliOptions(cameraOptions)

	client, err := adb.NewWithConfig(adb.ServerConfig{
		Port: 5037,
	})
	if err != nil {
		return nil, err
	}
	err = client.StartServer()
	if err != nil {
		return nil, err
	}

	deviceDescriptor := adb.AnyUsbDevice()
	if in != "any" {
		deviceDescriptor = adb.DeviceWithSerial(in)
	}
	device := client.Device(deviceDescriptor)

	entries, err := device.ListDirEntries("/sdcard/DCIM/Camera")
	if err != nil {
		return nil, err
	}
	if entries.Err() != nil {
		return nil, err
	}

	deviceInfo, err := device.DeviceInfo()
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	progressBar := mpb.New(mpb.WithWaitGroup(&wg),
		mpb.WithWidth(60),
		mpb.WithRefreshRate(180*time.Millisecond))

	inlineCounter := utils.ResultCounter{}

	for entries.Next() {
		mediaDate := entries.Entry().ModifiedAt.Format("02-01-2006")
		if strings.Contains(dateFormat, "yyyy") && strings.Contains(dateFormat, "mm") && strings.Contains(dateFormat, "dd") {
			mediaDate = entries.Entry().ModifiedAt.Format(replacer.Replace(dateFormat))
		}

		// check if is in date range
		dateStart := time.Date(0o000, time.Month(1), 1, 0, 0, 0, 0, time.UTC)

		dateEnd := time.Now()

		switch len(dateRange) {
		case 1:
			switch dateRange[0] {
			case "today":
				dateStart = time.Date(dateEnd.Year(), dateEnd.Month(), dateEnd.Day(), 0, 0, 0, 0, dateEnd.Location())
			case "yesterday":
				dateStart = time.Date(dateEnd.Year(), dateEnd.Month(), dateEnd.Day(), 0, 0, 0, 0, dateEnd.Location()).Add(-24 * time.Hour)
			case "week":
				dateStart = time.Date(dateEnd.Year(), dateEnd.Month(), dateEnd.Day(), 0, 0, 0, 0, dateEnd.Location()).Add(-24 * time.Duration((int(dateEnd.Weekday()) - 1)) * time.Hour)
			}
			if entries.Entry().ModifiedAt.Before(dateStart) {
				continue
			}
			if entries.Entry().ModifiedAt.After(dateEnd) {
				continue
			}
		case 2:
			layout := replacer.Replace(dateFormat)

			start, err := time.Parse(layout, dateRange[0])
			if err != nil {
				return nil, err
			}
			end, err := time.Parse(layout, dateRange[1])
			if err != nil {
				return nil, err
			}

			if entries.Entry().ModifiedAt.Before(start) || entries.Entry().ModifiedAt.After(end) {
				continue
			}
		}

		// Read Original file from device

		readfile, err := device.OpenRead("/sdcard/DCIM/Camera/" + entries.Entry().Name)
		if err != nil {
			result.Errors = append(result.Errors, err)
			result.FilesNotImported = append(result.FilesNotImported, entries.Entry().Name)
			return &result, nil //nolint
		}

		bar, dayFolder, err := prepare(
			out,
			entries.Entry().Name,
			deviceInfo.Product,
			mediaDate,
			sortOptions,
			readfile,
			progressBar,
		)
		if err != nil {
			result.Errors = append(result.Errors, err)
			result.FilesNotImported = append(result.FilesNotImported, entries.Entry().Name)
			return &result, nil //nolint
		}

		// Add 1 to queue for concurrency
		wg.Add(1)

		if entries.Entry().Name == "." || entries.Entry().Name == ".." {
			continue
		}

		if _, err := os.Stat(filepath.Join(dayFolder, "videos")); os.IsNotExist(err) {
			err = os.MkdirAll(filepath.Join(dayFolder, "videos"), 0o755)
			if err != nil {
				result.Errors = append(result.Errors, err)
				result.FilesNotImported = append(result.FilesNotImported, entries.Entry().Name)
				return &result, nil //nolint
			}
		}
		if _, err := os.Stat(filepath.Join(dayFolder, "photos")); os.IsNotExist(err) {
			err = os.MkdirAll(filepath.Join(dayFolder, "photos"), 0o755)
			if err != nil {
				result.Errors = append(result.Errors, err)
				result.FilesNotImported = append(result.FilesNotImported, entries.Entry().Name)
				return &result, nil //nolint
			}
		}

		localPath := ""
		if strings.HasSuffix(strings.ToLower(entries.Entry().Name), ".mp4") {
			localPath = filepath.Join(dayFolder, "videos", entries.Entry().Name)
		}

		filename, folder := pixelNameSort(entries.Entry().Name)
		if folder != "" {
			if _, err := os.Stat(filepath.Join(dayFolder, "photos", folder)); os.IsNotExist(err) {
				mkdirerr := os.MkdirAll(filepath.Join(dayFolder, "photos", folder), 0o755)
				if mkdirerr != nil {
					result.Errors = append(result.Errors, mkdirerr)
					result.FilesNotImported = append(result.FilesNotImported, entries.Entry().Name)
					return &result, nil //nolint
				}
			}

			localPath = filepath.Join(dayFolder, "photos", folder, filename)
		} else if strings.HasSuffix(strings.ToLower(entries.Entry().Name), ".jpg") {
			localPath = filepath.Join(dayFolder, "photos", entries.Entry().Name)
		}

		go func(filename, localPath string, bar *mpb.Bar) {
			defer wg.Done()
			readfile, err = device.OpenRead("/sdcard/DCIM/Camera/" + filename)
			if err != nil {
				inlineCounter.SetFailure(err, filename)
				return
			}
			defer readfile.Close()
			outFile, err := os.Create(localPath)
			if err != nil {
				inlineCounter.SetFailure(err, filename)
				return
			}
			defer outFile.Close()

			proxyReader := bar.ProxyReader(readfile)
			defer proxyReader.Close()

			_, err = io.Copy(outFile, proxyReader)
			if err != nil {
				inlineCounter.SetFailure(err, localPath)
				return
			}
			inlineCounter.SetSuccess()
		}(entries.Entry().Name, localPath, bar)
	}

	wg.Wait()
	progressBar.Shutdown()

	result.Errors = append(result.Errors, inlineCounter.Get().Errors...)
	result.FilesImported += inlineCounter.Get().FilesImported
	result.FilesNotImported = append(result.FilesNotImported, inlineCounter.Get().FilesNotImported...)

	return &result, nil
}
