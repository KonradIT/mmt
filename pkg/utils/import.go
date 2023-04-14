package utils

import (
	"log"
	"strings"
	"time"
)

type ImportParams struct {
	Input, Output, CameraName string
	SkipAuxiliaryFiles        bool
	DateFormat                string
	BufferSize                int
	Prefix                    string
	TagNames                  []string
	Connection                ConnectionType
	Sort                      SortOptions
}

type Import interface {
	Import(params ImportParams) (*Result, error)
}

var (
	dateEnd            time.Time
	dateStart          time.Time
	DateFormatReplacer = strings.NewReplacer("dd", "02", "mm", "01", "yyyy", "2006")
)

func ParseDateRange(dateRange []string, dateFormat string) {
	if len(dateRange) == 1 {
		today := time.Date(dateEnd.Year(), dateEnd.Month(), dateEnd.Day(), 0, 0, 0, 0, dateEnd.Location())
		switch dateRange[0] {
		case "today":
			dateStart = today
		case "yesterday":
			dateStart = today.Add(-24 * time.Hour)
		case "week":
			dateStart = today.Add(-24 * time.Duration((int(dateEnd.Weekday()) - 1)) * time.Hour)
		case "week-back":
			dateStart = today.Add(-24 * 7 * time.Hour)
		}
	}

	if len(dateRange) == 2 {
		start, err := time.Parse(DateFormatReplacer.Replace(dateFormat), dateRange[0])
		if err != nil {
			log.Fatal(err.Error())
		}
		if err == nil {
			dateStart = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
		}
		end, err := time.Parse(DateFormatReplacer.Replace(dateFormat), dateRange[1])
		if err != nil {
			log.Fatal(err.Error())
		}
		if err == nil {
			dateEnd = time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())
		}
	}
}

func IsValidDate(tm time.Time) bool {
	if (!dateStart.IsZero() && tm.Before(dateStart)) || (!dateEnd.IsZero() && tm.After(dateEnd)) {
		return false
	}

	return true
}

func DateZone() (string, int) {
	if dateEnd.IsZero() {
		return time.Now().Zone()
	}

	return dateEnd.Zone()
}
