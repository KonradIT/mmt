package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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

func getModDates(input string) ([]time.Time, error) {
	var modificationDates = []time.Time{}
	items, err := ioutil.ReadDir(input)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.IsDir() {
			subitems, err := ioutil.ReadDir(filepath.Join(input, item.Name()))
			if err != nil {
				return nil, err
			}
			for _, subitem := range subitems {
				if subitem.IsDir() {
					return getModDates(filepath.Join(input, item.Name(), subitem.Name()))
				} else {
					fileDate := subitem.ModTime()
					parsedDate := time.Date(fileDate.Year(), fileDate.Month(), fileDate.Day(), 0, 0, 0, 0, fileDate.Location())
					if !slices.Contains(modificationDates, parsedDate) {
						modificationDates = append(modificationDates, parsedDate)
					}
				}
			}

		} else {
			fileDate := item.ModTime()
			parsedDate := time.Date(fileDate.Year(), fileDate.Month(), fileDate.Day(), 0, 0, 0, 0, fileDate.Location())
			if !slices.Contains(modificationDates, parsedDate) {
				modificationDates = append(modificationDates, parsedDate)
			}
		}
	}
	return modificationDates, nil
}

var calendarView = &cobra.Command{
	Use:   "calendar",
	Short: "View days in which media was captured",
	Run: func(cmd *cobra.Command, args []string) {
		detectedGoPro, connectionType, err := gopro.Detect()
		if err != nil {
			cui.Error(err.Error())
		}

		var modificationDates = []time.Time{}

		switch connectionType {
		case utils.Connect:
			mediaList, err := gopro.GetMediaList(detectedGoPro)
			if err != nil {
				cui.Error(err.Error())
			}
			for _, folder := range mediaList.Media {
				for _, file := range folder.Fs {
					fileDate := time.Unix(file.Cre, 0)

					parsedDate := time.Date(fileDate.Year(), fileDate.Month(), fileDate.Day(), 0, 0, 0, 0, fileDate.Location())
					if !slices.Contains(modificationDates, parsedDate) {
						modificationDates = append(modificationDates, parsedDate)
					}
				}
			}
			break
		case utils.SDCard:
			m, err := getModDates(filepath.Join(detectedGoPro, fmt.Sprintf("%s", gopro.DCIM)))
			if err != nil {
				cui.Error(err.Error())
			}
			modificationDates = m
			break
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
