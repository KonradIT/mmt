package media

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dsoprea/go-exif/v3"
	"github.com/konradit/gopro-utils/telemetry"
	mErrors "github.com/konradit/mmt/pkg/errors"
	"github.com/konradit/mmt/pkg/utils"
	"github.com/konradit/mmt/pkg/videomanipulation"
	"github.com/rwcarlsen/goexif/exif"
	"gopkg.in/djherbis/times.v1"
)

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

	f, err := os.Open(osPathname)
	if err != nil {
		return d
	}
	defer f.Close()
	x, err := exif.Decode(f)
	if err != nil {
		return d
	}

	// First search in gps track
	if strings.Contains(osPathname, ".MP4") {
		err := GetTimeFromMP4(osPathname, &date)
		if err == nil {
			return date
		}
	}

	var gpsDateTime string
	gpsDateStamp, _ := x.Get(exif.GPSDateStamp)
	gpsTimeStamp, _ := x.Get(exif.GPSTimeStamp)

	gpsT, _ := gpsTimeStamp.StringVal()
	gpsD, _ := gpsDateStamp.StringVal()

	gpsDateTime = gpsD + " " + gpsT

	// parse the string into time
	date, err = time.Parse("2006:01:02 15:04:05", gpsDateTime)
	if err == nil {
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

func GetTimeFromMP4(videoPath string, date *time.Time) error {
	vman := videomanipulation.New()
	data, err := vman.ExtractGPMF(videoPath)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(*data)

	lastEvent := &telemetry.TELEM{}

	for {
		event, err := telemetry.Read(reader)
		if err != nil && err != io.EOF {
			return err
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
			return err
		}

		telems := lastEvent.ShitJson()
		for _, telem := range telems {
			if telem.Latitude != 0 && telem.Longitude != 0 {
				*date = time.Unix(0, telem.TS)

				return nil
			}
		}
		*lastEvent = *event
	}

	return mErrors.ErrNoGPS
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
