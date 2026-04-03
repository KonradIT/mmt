package cmd

import (
	"os"
	"strconv"
	"time"

	"github.com/erdaltsksn/cui"
	"github.com/fatih/color"
	"github.com/konradit/mmt/pkg/camera"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func ac(d time.Weekday) string {
	return d.String()[0:1]
}

func pad(d time.Weekday) int {
	return int(d)
}

func splitSliceInChunks(a []string, chuckSize int) [][]string {
	chunks := [][]string{}
	for chuckSize < len(a) {
		a, chunks = a[chuckSize:], append(chunks, a[0:chuckSize:chuckSize])
	}
	chunks = append(chunks, a)
	return chunks
}

var calendarView = &cobra.Command{
	Use:   "calendar",
	Short: "View days in which media was captured",
	Run: func(_ *cobra.Command, _ []string) {
		cam, err := camera.Get("gopro")
		if err != nil {
			cui.Error(err.Error())
		}

		input, connectionType, err := cam.Detect()
		if err != nil {
			cui.Error(err.Error())
		}

		cp, ok := cam.(camera.CalendarProvider)
		if !ok {
			cui.Error("camera does not support calendar view")
			return
		}

		modificationDates, err := cp.CaptureDates(input, connectionType)
		if err != nil {
			cui.Error(err.Error())
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{
			ac(time.Sunday),
			ac(time.Monday),
			ac(time.Tuesday),
			ac(time.Wednesday),
			ac(time.Thursday),
			ac(time.Friday),
			ac(time.Saturday),
		})

		// Get first day of current month
		now := time.Now()
		currentYear, currentMonth, _ := now.Date()
		currentLocation := now.Location()
		firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)

		data := []string{}

		// pad for empty days
		for i := 1; i <= pad(firstOfMonth.Weekday()); i++ {
			data = append(data, " ")
		}

		for i := 1; i <= firstOfMonth.AddDate(0, 1, -1).Day(); i++ {
			date := time.Date(currentYear, currentMonth, i, 0, 0, 0, 0, currentLocation)
			// No slices.Contains because we need to check for equality, not just presence.
			// Time.time.Equal ignores monotonic clock and location.
			found := false
			for _, d := range modificationDates {
				if d.Equal(date) {
					found = true
					break
				}
			}
			if found {
				data = append(data, color.CyanString(strconv.Itoa(i)))
			} else {
				data = append(data, color.YellowString(strconv.Itoa(i)))
			}
		}
		prepared := splitSliceInChunks(data, 7)
		for _, v := range prepared {
			table.Append(v)
		}
		table.Render()
	},
}

func init() {
	rootCmd.AddCommand(calendarView)
}
