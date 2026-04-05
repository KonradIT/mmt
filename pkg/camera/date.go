package camera

import (
	"strings"
	"time"

	"github.com/konradit/mmt/pkg/utils"
)

// ModTimeAsUTC reinterprets a wall-clock time as UTC. FAT32/exFAT
// filesystems store times without timezone information; the OS adds the
// local offset when reporting ModTime. This undoes that interpretation so
// all camera timestamps are comparable in a single timezone.
func ModTimeAsUTC(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(),
		d.Hour(), d.Minute(), d.Second(), d.Nanosecond(), time.UTC)
}

// FormatMediaDate formats a time into a date string using the given format.
// It falls back to dd-mm-yyyy if the format string is not recognized.
func FormatMediaDate(d time.Time, dateFormat string) string {
	if strings.Contains(dateFormat, "yyyy") && strings.Contains(dateFormat, "mm") && strings.Contains(dateFormat, "dd") {
		return d.Format(utils.DateFormatReplacer.Replace(dateFormat))
	}

	return d.Format("02-01-2006")
}
