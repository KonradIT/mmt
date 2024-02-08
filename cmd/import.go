package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/erdaltsksn/cui"
	"github.com/fatih/color"
	"github.com/konradit/mmt/pkg/android"
	"github.com/konradit/mmt/pkg/dji"
	mErrors "github.com/konradit/mmt/pkg/errors"
	"github.com/konradit/mmt/pkg/gopro"
	"github.com/konradit/mmt/pkg/insta360"
	"github.com/konradit/mmt/pkg/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import media",
	Run: func(cmd *cobra.Command, args []string) {
		input := getFlagString(cmd, "input", "")
		output := getFlagString(cmd, "output", "")
		camera := getFlagString(cmd, "camera", "")
		projectName := getFlagString(cmd, "name", "")

		if projectName != "" {
			_, err := os.Stat(filepath.Join(output, projectName))
			if os.IsNotExist(err) {
				err := os.Mkdir(filepath.Join(output, projectName), 0o755)
				if err != nil {
					cui.Error("Something went wrong creating project dir", err)
				}
			}
		}

		dateFormat := getFlagString(cmd, "date", "dd-mm-yyyy")
		bufferSize := getFlagInt(cmd, "buffer", "1000")
		prefix := getFlagString(cmd, "prefix", "")
		dateRange := getFlagSlice(cmd, "range")
		cameraName := getFlagString(cmd, "camera-name", "")
		connection := utils.ConnectionType(getFlagString(cmd, "connection", ""))
		skipAuxFiles := getFlagBool(cmd, "skip-aux", "true")
		sortBy := getFlagSlice(cmd, "sort-by")
		if len(sortBy) == 0 {
			sortBy = []string{"camera", "location"}
		}
		sortOptions := utils.SortOptions{
			ByLocation: slices.Contains(sortBy, "location"),
			ByCamera:   slices.Contains(sortBy, "camera"),
		}
		tagNames := getFlagSlice(cmd, "tag-names")

		if useGoPro, err := cmd.Flags().GetBool("use-gopro"); err == nil && useGoPro {
			detectedGoPro, connectionType, err := gopro.Detect()
			if err != nil {
				cui.Error(err.Error())
			}
			input = detectedGoPro
			connection = connectionType
			camera = "gopro"
		} else if useInsta360, err := cmd.Flags().GetBool("use-insta360"); err == nil && useInsta360 {
			detectedInsta360, connectionType, err := insta360.Detect()
			if err != nil {
				cui.Error(err.Error())
			}
			input = detectedInsta360
			connection = connectionType
			camera = "insta360"
		}

		if camera != "" && output != "" {
			c, err := utils.CameraGet(camera)
			if err != nil {
				cui.Error("Something went wrong", err)
			}

			switch c {
			case utils.GoPro:
				if connection == "" {
					connection = utils.SDCard
				}
			}

			params := utils.ImportParams{
				Input:              input,
				Output:             filepath.Join(output, projectName),
				CameraName:         cameraName,
				SkipAuxiliaryFiles: skipAuxFiles,
				DateFormat:         dateFormat,
				BufferSize:         bufferSize,
				Prefix:             prefix,
				DateRange:          parseDateRange(dateRange, dateFormat),
				TagNames:           tagNames,
				Connection:         connection,
				Sort:               sortOptions,
			}
			r, err := importFromCamera(c, params)
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
			table.Render() // Send output

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
	importCmd.Flags().String("camera-name", "", "Override camera name detection with specified string")

	// Camera helpers
	importCmd.Flags().Bool("use-gopro", false, "Detect GoPro camera attached")
	importCmd.Flags().Bool("use-insta360", false, "Detect Insta360 camera attached")
}

func parseDateRange(dateRange []string, dateFormat string) []time.Time {
	dateStart := time.Date(0o000, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
	dateEnd := time.Now()

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
			dateStart = today.Add((-24 * 7) * time.Hour)
		}
	}

	if len(dateRange) == 2 {
		start, err := time.Parse(utils.DateFormatReplacer.Replace(dateFormat), dateRange[0])
		if err != nil {
			log.Fatal(err.Error())
		}
		if err == nil {
			dateStart = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
		}
		end, err := time.Parse(utils.DateFormatReplacer.Replace(dateFormat), dateRange[1])
		if err != nil {
			log.Fatal(err.Error())
		}
		if err == nil {
			dateEnd = time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())
		}
	}

	return []time.Time{dateStart, dateEnd}
}

func callImport(cameraIf utils.Import, params utils.ImportParams) (*utils.Result, error) {
	return cameraIf.Import(params)
}

func importFromCamera(c utils.Camera, params utils.ImportParams) (*utils.Result, error) {
	switch c {
	case utils.GoPro:
		return callImport(gopro.Entrypoint{}, params)
	case utils.DJI:
		return callImport(dji.Entrypoint{}, params)
	case utils.Insta360:
		return callImport(insta360.Entrypoint{}, params)
	case utils.Android:
		return callImport(android.Entrypoint{}, params)
	}
	return nil, mErrors.ErrUnsupportedCamera("")
}
