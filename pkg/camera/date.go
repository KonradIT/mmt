package camera

import (
	"strings"
	"time"

	"github.com/konradit/mmt/pkg/utils"
)

// FormatMediaDate formats a time into a date string using the given format.
// It falls back to dd-mm-yyyy if the format string is not recognized.
func FormatMediaDate(d time.Time, dateFormat string) string {
	if strings.Contains(dateFormat, "yyyy") && strings.Contains(dateFormat, "mm") && strings.Contains(dateFormat, "dd") {
		return d.Format(utils.DateFormatReplacer.Replace(dateFormat))
	}

	return d.Format("02-01-2006")
}
