package android

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/konradit/mmt/pkg/camera"
	mErrors "github.com/konradit/mmt/pkg/errors"
	"github.com/vbauerster/mpb/v8"
	adb "github.com/zach-klippenstein/goadb"
)

func pixelNameSort(filename string) (string, string) {
	if strings.Contains(filename, "MOTION") {
		s := strings.Split(filename, ".MOTION")

		return filename, s[0]
	}

	return filename, ""
}

var locationService = LocationService{}

type Entrypoint struct{}

func init() {
	camera.Register(Entrypoint{})
}

func (Entrypoint) Name() string { return "android" }

func (Entrypoint) GuessFromPath(_ string) bool {
	// Android uses ADB, not filesystem markers.
	return false
}

func (Entrypoint) Detect() (string, camera.ConnectionType, error) {
	client, err := adb.NewWithConfig(adb.ServerConfig{Port: 5037})
	if err != nil {
		return "", "", mErrors.ErrNoCameraDetected
	}

	if err := client.StartServer(); err != nil {
		return "", "", mErrors.ErrNoCameraDetected
	}

	device := client.Device(adb.AnyUsbDevice())

	info, err := device.DeviceInfo()
	if err != nil {
		return "", "", mErrors.ErrNoCameraDetected
	}

	return info.Serial, camera.SDCard, nil
}

func prepare(out string, deviceFileName string, deviceModel string, mediaDate string, sortOptions camera.SortOptions, deviceFileReader io.ReadCloser, progressBar *mpb.Progress) (*mpb.Bar, string, error) {
	localFile, err := os.CreateTemp(out, deviceFileName)
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

	bar := camera.GetNewBar(progressBar, stat.Size(), deviceFileName, camera.IoTX)

	dayFolder := camera.GetOrder(sortOptions, locationService, filepath.Join(out, localFile.Name()), out, mediaDate, deviceModel)

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

func (Entrypoint) Import(params camera.ImportParams) (*camera.Result, error) {
	var result camera.Result

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
	if params.Input != "any" {
		deviceDescriptor = adb.DeviceWithSerial(params.Input)
	}

	device := client.Device(deviceDescriptor)

	entries, err := device.ListDirEntries("/sdcard/DCIM/Camera")
	if err != nil {
		return nil, err
	}

	if entries.Err() != nil {
		return nil, entries.Err()
	}

	deviceInfo, err := device.DeviceInfo()
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup

	progressBar := mpb.New(mpb.WithWaitGroup(&wg),
		mpb.WithWidth(60),
		mpb.WithRefreshRate(180*time.Millisecond))

	inlineCounter := camera.ResultCounter{}

	for entries.Next() {
		mediaDate := camera.FormatMediaDate(entries.Entry().ModifiedAt, params.DateFormat)

		// check if is in date range
		if entries.Entry().ModifiedAt.Before(params.DateRange[0]) || entries.Entry().ModifiedAt.After(params.DateRange[1]) {
			continue
		}

		// Read Original file from device

		readfile, err := device.OpenRead("/sdcard/DCIM/Camera/" + entries.Entry().Name)
		if err != nil {
			result.Errors = append(result.Errors, err)
			result.FilesNotImported = append(result.FilesNotImported, entries.Entry().Name)
			return &result, nil //nolint
		}

		bar, dayFolder, err := prepare(
			params.Output,
			entries.Entry().Name,
			deviceInfo.Product,
			mediaDate,
			params.Sort,
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

		if err := os.MkdirAll(filepath.Join(dayFolder, "videos"), 0o755); err != nil {
			result.Errors = append(result.Errors, err)
			result.FilesNotImported = append(result.FilesNotImported, entries.Entry().Name)
			return &result, nil //nolint
		}

		if err := os.MkdirAll(filepath.Join(dayFolder, "photos"), 0o755); err != nil {
			result.Errors = append(result.Errors, err)
			result.FilesNotImported = append(result.FilesNotImported, entries.Entry().Name)
			return &result, nil //nolint
		}

		localPath := ""
		if strings.HasSuffix(strings.ToLower(entries.Entry().Name), ".mp4") {
			localPath = filepath.Join(dayFolder, "videos", entries.Entry().Name)
		}

		filename, folder := pixelNameSort(entries.Entry().Name)
		if folder != "" {
			err := os.MkdirAll(filepath.Join(dayFolder, "photos", folder), 0o755)
			if err != nil {
				result.Errors = append(result.Errors, err)
				result.FilesNotImported = append(result.FilesNotImported, entries.Entry().Name)
				return &result, nil //nolint
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
