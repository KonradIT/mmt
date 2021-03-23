package android

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/konradit/mmt/pkg/utils"
	adb "github.com/zach-klippenstein/goadb"
)

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
		if strings.Contains(dateFormat, "year") && strings.Contains(dateFormat, "month") && strings.Contains(dateFormat, "day") {
			mediaDate = entries.Entry().ModifiedAt.Format(replacer.Replace(dateFormat))
		}

		// check if is in date range

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
			os.Mkdir(filepath.Join(dayFolder, deviceInfo.Product), 0755)
		}
		dayFolder = filepath.Join(dayFolder, deviceInfo.Product)
		fmt.Printf("\t%+v\n", *entries.Entry())
		fmt.Printf("\t%+v", entries.Entry().ModifiedAt)

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

		readfile, err := device.OpenRead("/sdcard/DCIM/Camera/" + entries.Entry().Name)
		if err != nil {
			result.Errors = append(result.Errors, err)
			result.FilesNotImported = append(result.FilesNotImported, entries.Entry().Name)
			return &result, nil

		} else {
			result.FilesImported += 1
		}

		localPath := ""
		if strings.HasSuffix(strings.ToLower(entries.Entry().Name), ".mp4") {
			localPath = filepath.Join(dayFolder, "videos", entries.Entry().Name)
		}
		if strings.HasSuffix(strings.ToLower(entries.Entry().Name), ".jpg") {
			localPath = filepath.Join(dayFolder, "photos", entries.Entry().Name)
		}
		outFile, err := os.Create(localPath)
		if err != nil {
			result.Errors = append(result.Errors, err)
			result.FilesNotImported = append(result.FilesNotImported, entries.Entry().Name)
			return &result, nil

		} else {
			result.FilesImported += 1
		}
		defer outFile.Close()
		_, err = io.Copy(outFile, readfile)
		if err != nil {
			result.Errors = append(result.Errors, err)
			result.FilesNotImported = append(result.FilesNotImported, entries.Entry().Name)
			return &result, nil
		} else {
			result.FilesImported += 1
		}

	}

	return &result, nil
}
