package android

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/konradit/mmt/pkg/utils"
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
func Import(in, out, dateFormat string, bufferSize int, prefix string, dateRange []string) (*utils.Result, error) {

	var result utils.Result

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

	device := client.Device(adb.AnyUsbDevice())
	entries, err := device.ListDirEntries("/sdcard/DCIM/Camera")
	if err != nil {
		return nil, err
	}
	if entries.Err() != nil {
		return nil, err
	}

	for entries.Next() {

		replacer := strings.NewReplacer("dd", "02", "mm", "01", "yyyy", "2006")
		mediaDate := entries.Entry().ModifiedAt.Format("02-01-2006")
		if strings.Contains(dateFormat, "yyyy") && strings.Contains(dateFormat, "mm") && strings.Contains(dateFormat, "dd") {
			mediaDate = entries.Entry().ModifiedAt.Format(replacer.Replace(dateFormat))
		}

		// check if is in date range
		dateStart := time.Date(0000, time.Month(1), 1, 0, 0, 0, 0, time.UTC)

		dateEnd := time.Now()

		if len(dateRange) == 1 {
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

		}

		if len(dateRange) == 2 {

			layout := replacer.Replace(dateFormat)

			start, err1 := time.Parse(layout, dateRange[0])
			end, err2 := time.Parse(layout, dateRange[1])
			if err1 == nil && err2 == nil {
				if entries.Entry().ModifiedAt.Before(start) {
					continue
				}
				if entries.Entry().ModifiedAt.After(end) {
					continue
				}
			}

		}

		dayFolder := filepath.Join(out, mediaDate)
		if _, err := os.Stat(dayFolder); os.IsNotExist(err) {
			os.Mkdir(dayFolder, 0755)
		}

		deviceInfo, err := device.DeviceInfo()
		if err != nil {
			return nil, err
		}
		if _, err := os.Stat(filepath.Join(dayFolder, deviceInfo.Product)); os.IsNotExist(err) {
			_ = os.Mkdir(filepath.Join(dayFolder, deviceInfo.Product), 0755)
		}
		dayFolder = filepath.Join(dayFolder, deviceInfo.Product)

		if entries.Entry().Name == "." || entries.Entry().Name == ".." {
			continue
		}

		if _, err := os.Stat(filepath.Join(dayFolder, "videos")); os.IsNotExist(err) {
			err = os.MkdirAll(filepath.Join(dayFolder, "videos"), 0755)
			if err != nil {
				log.Fatal(err.Error())
			}
		}
		if _, err := os.Stat(filepath.Join(dayFolder, "photos")); os.IsNotExist(err) {
			err = os.MkdirAll(filepath.Join(dayFolder, "photos"), 0755)
			if err != nil {
				log.Fatal(err.Error())
			}
		}
		color.Cyan(">>> " + entries.Entry().Name)
		readfile, err := device.OpenRead("/sdcard/DCIM/Camera/" + entries.Entry().Name)
		if err != nil {
			result.Errors = append(result.Errors, err)
			result.FilesNotImported = append(result.FilesNotImported, entries.Entry().Name)
			return &result, nil

		}
		localPath := ""
		if strings.HasSuffix(strings.ToLower(entries.Entry().Name), ".mp4") {
			localPath = filepath.Join(dayFolder, "videos", entries.Entry().Name)
		}
		if strings.HasSuffix(strings.ToLower(entries.Entry().Name), ".jpg") {
			filename, folder := pixelNameSort(entries.Entry().Name)
			if folder != "" {
				if _, err := os.Stat(filepath.Join(dayFolder, "photos", folder)); os.IsNotExist(err) {
					err = os.MkdirAll(filepath.Join(dayFolder, "photos", folder), 0755)
					if err != nil {
						log.Fatal(err.Error())
					}
				}
			}
			localPath = filepath.Join(dayFolder, "photos", folder, filename)
		}
		outFile, err := os.Create(localPath)
		if err != nil {
			result.Errors = append(result.Errors, err)
			result.FilesNotImported = append(result.FilesNotImported, entries.Entry().Name)
			return &result, nil

		}
		defer outFile.Close()
		_, err = io.Copy(outFile, readfile)
		if err != nil {
			result.Errors = append(result.Errors, err)
			result.FilesNotImported = append(result.FilesNotImported, entries.Entry().Name)
			return &result, nil
		}
		result.FilesImported += 1

	}

	return &result, nil
}
