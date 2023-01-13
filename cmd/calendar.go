package cmd

import (
	"os"
	"strconv"
	"time"

	"github.com/erdaltsksn/cui"
	"github.com/fatih/color"
	"github.com/konradit/mmt/pkg/gopro"
	"github.com/konradit/mmt/pkg/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

func ac(d time.Weekday) string {
	return d.String()[0:1]
}

func pad(d time.Weekday) int {
	// How much to pad
	// Monday: 0
	// Sunday: 6
	return int(d)
}

func SplitSliceInChunks(a []string, chuckSize int) [][]string {
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
	Run: func(cmd *cobra.Command, args []string) {
		detectedGoPro, connectionType, err := gopro.Detect()
		if err != nil {
			cui.Error(err.Error())
		}
		if connectionType != utils.Connect {
			cui.Error("Not GoPro Connect")
		}

		mediaList, err := gopro.GetMediaList(detectedGoPro)
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

		var modificationDates = []time.Time{}

		for _, folder := range mediaList.Media {
			for _, file := range folder.Fs {
				fileDate := time.Unix(file.Cre, 0)

				parsedDate := time.Date(fileDate.Year(), fileDate.Month(), fileDate.Day(), 0, 0, 0, 0, fileDate.Location())
				if !slices.Contains(modificationDates, parsedDate) {
					modificationDates = append(modificationDates, parsedDate)
				}
			}
		}

		for i := 1; i <= 31; i++ {
			date := time.Date(currentYear, currentMonth, i, 0, 0, 0, 0, currentLocation)
			if slices.Contains(modificationDates, date) {
				data = append(data, color.CyanString(strconv.Itoa(i)))
			} else {
				data = append(data, color.YellowString(strconv.Itoa(i)))
			}
		}
		prepared := SplitSliceInChunks(data, 7)
		for _, v := range prepared {
			table.Append(v)
		}
		table.Render()
	},
}

func init() {
	rootCmd.AddCommand(calendarView)
}
