package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"github.com/erdaltsksn/cui"
	"github.com/fatih/color"
	"github.com/konradit/mmt/pkg/camera"
	"github.com/konradit/mmt/pkg/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type detector interface {
	Detect() (string, camera.ConnectionType, error)
}

type importer interface {
	Import(params camera.ImportParams) (*camera.Result, error)
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import media",
	Run: func(cmd *cobra.Command, _ []string) {
		input := getFlagString(cmd, "input", "")
		output := getFlagString(cmd, "output", "")
		cameraType := getFlagString(cmd, "camera", "")
		projectName := getFlagString(cmd, "name", "")

		if projectName != "" {
			err := os.MkdirAll(filepath.Join(output, projectName), 0o755)
			if err != nil && !errors.Is(err, fs.ErrExist) {
				cui.Error("Something went wrong creating project dir", err)
			}
		}

		dateFormat := getFlagString(cmd, "date", "dd-mm-yyyy")
		bufferSize := getFlagInt(cmd, "buffer", "1000")
		maxConcurrent := getFlagInt(cmd, "concurrent", "10")
		prefix := getFlagString(cmd, "prefix", "")
		dateRange := getFlagSlice(cmd, "range")
		cameraName := getFlagString(cmd, "camera-name", "")
		connection := camera.ConnectionType(getFlagString(cmd, "connection", ""))
		skipAuxFiles := getFlagBool(cmd, "skip-aux", "true")

		sortBy := getFlagSlice(cmd, "sort-by")
		if len(sortBy) == 0 {
			sortBy = []string{"camera", "location"}
		}

		sortOptions := camera.SortOptions{
			ByLocation: slices.Contains(sortBy, "location"),
			ByCamera:   slices.Contains(sortBy, "camera"),
		}
		tagNames := getFlagSlice(cmd, "tag-names")

		useGoPro, err := cmd.Flags().GetBool("use-gopro")
		if err == nil && useGoPro {
			cam, err := camera.Get("gopro")
			if err != nil {
				cui.Error(err.Error())
			}

			det, ok := cam.(detector)
			if !ok {
				cui.Error("gopro camera does not support detection")
			}

			detectedInput, connType, err := det.Detect()
			if err != nil {
				cui.Error(err.Error())
			}

			input = detectedInput
			connection = connType
			cameraType = "gopro"
		}

		useInsta360, err := cmd.Flags().GetBool("use-insta360")
		if err == nil && useInsta360 {
			cam, err := camera.Get("insta360")
			if err != nil {
				cui.Error(err.Error())
			}

			det, ok := cam.(detector)
			if !ok {
				cui.Error("insta360 camera does not support detection")
			}

			detectedInput, connType, err := det.Detect()
			if err != nil {
				cui.Error(err.Error())
			}

			input = detectedInput
			connection = connType
			cameraType = "insta360"
		}

		if cameraType != "" && output != "" {
			cam, err := camera.Get(cameraType)
			if err != nil {
				cui.Error("Something went wrong", err)
			}

			imp, ok := cam.(importer)
			if !ok {
				cui.Error(fmt.Sprintf("camera %q does not support import", cameraType))
			}

			if cameraType == "gopro" && connection == "" {
				connection = camera.SDCard
			}

			dateRangeParsed, err := parseDateRange(dateRange, dateFormat)
			if err != nil {
				cui.Error("Invalid date range", err)
			}

			params := camera.ImportParams{
				Input:              input,
				Output:             filepath.Join(output, projectName),
				CameraName:         cameraName,
				SkipAuxiliaryFiles: skipAuxFiles,
				DateFormat:         dateFormat,
				BufferSize:         bufferSize,
				MaxConcurrent:      maxConcurrent,
				Prefix:             prefix,
				DateRange:          dateRangeParsed,
				TagNames:           tagNames,
				Connection:         connection,
				Sort:               sortOptions,
			}

			r, err := imp.Import(params)
			if err != nil {
				cui.Error("Something went wrong", err)
			}

			data := [][]string{
				{strconv.Itoa(r.FilesImported), strconv.Itoa(len(r.FilesNotImported)), strconv.Itoa(len(r.Errors))},
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Files Imported", "Files Skipped", "Errors"})

			for _, v := range data {
				table.Append(v)
			}

			table.Render()

			if len(r.Errors) != 0 {
				fmt.Println("Errors: ")

				for _, error := range r.Errors {
					color.Red(">> " + error.Error())
				}
			}

			return
		}

		color.Red("Error: required flag(s) \"camera\", \"output\" not set")
	},
}

func init() {
	rootCmd.AddCommand(importCmd)

	importCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose")
	importCmd.Flags().StringP("input", "i", "", "Input directory for root, eg: E:\\")
	importCmd.Flags().StringP("output", "o", "", "Output directory for sorted media")
	importCmd.Flags().StringP("name", "n", "", "Project name")
	importCmd.Flags().StringP("camera", "c", "", "Camera type")
	importCmd.Flags().StringP("date", "d", "", "Date format, dd-mm-yyyy by default")
	importCmd.Flags().StringP("buffer", "b", "", "Buffer size for copying, default is 1000 bytes")
	importCmd.Flags().StringP("prefix", "p", "", "Prefix for each file, pass `cameraname` to prepend the camera name (eg: Hero9 Black)")
	importCmd.Flags().StringSlice("range", []string{}, "A date range, eg: 01-05-2020,05-05-2020 -- also accepted: `today`, `yesterday`, `week`")
	importCmd.Flags().StringP("connection", "x", "", "Connexion type: `sd_card`, `connect` (GoPro-specific)")
	importCmd.Flags().StringSlice("sort-by", []string{}, "Sort files by: `camera`, `location`")
	importCmd.Flags().StringSlice("tag-names", []string{}, "Tag names for number of HiLight tags in last 10s of video, each position being the amount, eg: 'marked 1,good stuff,important' => num of tags: 1,2,3")
	importCmd.Flags().StringP("skip-aux", "s", "true", "Skip auxiliary files (GoPro: THM, LRV. DJI: SRT)")
	importCmd.Flags().String("concurrent", "10", "Maximum number of files to copy concurrently")
	importCmd.Flags().String("camera-name", "", "Override camera name detection with specified string")

	// Camera helpers
	importCmd.Flags().Bool("use-gopro", false, "Detect GoPro camera attached")
	importCmd.Flags().Bool("use-insta360", false, "Detect Insta360 camera attached")
}

func parseDateRange(dateRange []string, dateFormat string) ([]time.Time, error) {
	dateStart := time.Date(0o000, time.January, 1, 0, 0, 0, 0, time.UTC)
	dateEnd := time.Now().UTC()

	if len(dateRange) == 1 {
		today := time.Date(dateEnd.Year(), dateEnd.Month(), dateEnd.Day(), 0, 0, 0, 0, time.UTC)

		switch dateRange[0] {
		case "today":
			dateStart = today
		case "yesterday":
			dateStart = today.Add(-24 * time.Hour)
		case "week":
			dateStart = today.Add(-24 * time.Duration((int(dateEnd.Weekday()) - 1)) * time.Hour)
		case "week-back":
			dateStart = today.Add((-24 * 7) * time.Hour)
		}
	}

	if len(dateRange) == 2 {
		start, err := time.Parse(utils.DateFormatReplacer.Replace(dateFormat), dateRange[0])
		if err != nil {
			return nil, fmt.Errorf("invalid start date %q: %w", dateRange[0], err)
		}

		dateStart = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)

		end, err := time.Parse(utils.DateFormatReplacer.Replace(dateFormat), dateRange[1])
		if err != nil {
			return nil, fmt.Errorf("invalid end date %q: %w", dateRange[1], err)
		}

		dateEnd = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, time.UTC)
	}

	return []time.Time{dateStart, dateEnd}, nil
}
