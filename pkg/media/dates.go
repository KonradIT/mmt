package media

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/konradit/gopro-utils/telemetry"
	"github.com/konradit/mmt/pkg/utils"
	"github.com/konradit/mmt/pkg/videomanipulation"
	"github.com/rwcarlsen/goexif/exif"
	"gopkg.in/djherbis/times.v1"
)

func GetGPSTime(x *exif.Exif, date *time.Time) bool {
	var gpsDateTime string

	gpsDateStamp, err := x.Get(exif.GPSDateStamp)
	if err != nil {
		return false
	}
	gpsTimeStamp, err := x.Get(exif.GPSTimeStamp)
	if err != nil {
		return false
	}

	gpsD, err := gpsDateStamp.StringVal()
	if err != nil {
		return false
	}

	// convert rational to string
	gpsH, _, err := gpsTimeStamp.Rat2(0)
	if err != nil {
		return false
	}
	gpsM, _, err := gpsTimeStamp.Rat2(1)
	if err != nil {
		return false
	}
	gpsS, _, err := gpsTimeStamp.Rat2(2)
	if err != nil {
		return false
	}

	gpsDateTime = fmt.Sprintf("%s %02d:%02d:%02d", gpsD, gpsH, gpsM, gpsS)

	// parse the string into time
	d, err := time.Parse("2006:01:02 15:04:05", gpsDateTime)
	if err == nil {
		*date = d
		return true
	}

	return false
}

func GetFileTimeExif(osPathname string) time.Time {
	var date time.Time

	if strings.Contains(osPathname, ".WAV") {
		osPathname = osPathname[:len(osPathname)-len(filepath.Ext(osPathname))] + ".MP4"
	}

	t, err := times.Stat(osPathname)
	if err != nil {
		log.Fatal(err.Error())
	}

	d := t.ModTime()

	// First search in gps track
	if strings.Contains(osPathname, ".MP4") {
		if GetTimeFromMP4(osPathname, &date) {
			return date
		}
	}

	f, err := os.Open(osPathname)
	if err != nil {
		return d
	}
	defer f.Close()
	x, err := exif.Decode(f)
	if err != nil {
		return d
	}

	if GetGPSTime(x, &date) {
		return date
	}

	// define the list of possible tags to extract date from
	dateTags := []string{"DateTimeOriginal", "DateTime", "DateTimeDigitized"}

	// loop for each tag and return the first valid date
	for _, tag := range dateTags {
		// get value of tag from exif
		tt, err := x.Get(exif.FieldName(tag))
		if err != nil {
			tts, _ := tt.StringVal()
			date, err = time.Parse("2006:01:02 15:04:05", tts)
			if err != nil {
				continue
			}
			return date
		}
	}

	return d
}

func GetTimeFromMP4(videoPath string, date *time.Time) bool {
	vman := videomanipulation.New()
	data, err := vman.ExtractGPMF(videoPath)
	if err != nil {
		return false
	}

	reader := bytes.NewReader(*data)

	lastEvent := &telemetry.TELEM{}

	for {
		event, err := telemetry.Read(reader)
		if err != nil && err != io.EOF {
			return false
		} else if err == io.EOF || event == nil {
			break
		}

		if lastEvent.IsZero() {
			*lastEvent = *event
			event.Clear()
			continue
		}

		err = lastEvent.FillTimes(event.Time.Time)
		if err != nil {
			return false
		}

		telems := lastEvent.ShitJson()
		for _, telem := range telems {
			if telem.Latitude != 0 && telem.Longitude != 0 {
				*date = time.UnixMicro(telem.TS)

				return true
			}
		}
		*lastEvent = *event
	}

	return false
}

func GetFileTime(osPathname string, utcFix bool) time.Time {
	t := GetFileTimeExif(osPathname)

	if utcFix {
		zoneName, _ := t.Zone()
		newTime := strings.Replace(t.Format(time.UnixDate), zoneName, "UTC", -1)
		t, _ = time.Parse(time.UnixDate, newTime)
	}

	return t
}

func GetMediaDate(d time.Time, dateFormat string) string {
	mediaDate := d.Format("02-01-2006")
	if strings.Contains(dateFormat, "yyyy") && strings.Contains(dateFormat, "mm") && strings.Contains(dateFormat, "dd") {
		mediaDate = d.Format(utils.DateFormatReplacer.Replace(dateFormat))
	}

	return mediaDate
}
