package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/erdaltsksn/cui"
	"github.com/fatih/color"
	"github.com/konradit/mmt/pkg/android"
	"github.com/konradit/mmt/pkg/dji"
	"github.com/konradit/mmt/pkg/gopro"
	"github.com/konradit/mmt/pkg/insta360"
	"github.com/konradit/mmt/pkg/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func get_flag_string(cmd *cobra.Command, name string) string {
	value, err := cmd.Flags().GetString(name)
        if err != nil {
                cui.Error("Problem parsing "+name, err)
        }
        if value == "" {
          value = viper.GetString("import."+name)
        }
        return value
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import media",
	Run: func(cmd *cobra.Command, args []string) {
                input := get_flag_string(cmd, "input")
                output := get_flag_string(cmd, "output")
                camera := get_flag_string(cmd, "camera")
                projectName := get_flag_string(cmd, "name")

		if projectName != "" {
			os.Mkdir(filepath.Join(output, projectName), 0755)
		}

                dateFormat := get_flag_string(cmd, "date")

		bufferSize, err := cmd.Flags().GetInt("buffer")
		if err != nil {
			cui.Error("Problem parsing buffer", err)
		}

                prefix := get_flag_string(cmd, "prefix")

		dateRange, err := cmd.Flags().GetStringSlice("range")
		if err != nil {
			cui.Error("Problem parsing range", err)
		}

		if camera != "" {
			c, err := utils.CameraGet(camera)
			if err != nil {
				cui.Error("Something went wrong", err)
			}

			customCameraOpts := make(map[string]interface{})
			if c == utils.GoPro {
				skipAuxFiles, err := cmd.Flags().GetBool("skip_aux")
				if err == nil {
					customCameraOpts["skip_aux"] = skipAuxFiles
				}

				sortBy, err := cmd.Flags().GetStringSlice("sort_by")
				if err == nil {
					customCameraOpts["sort_by"] = sortBy
				}

                                connection := get_flag_string(cmd, "connection")
				if connection == "" {
                                        connection = "sd_card"
				}
				customCameraOpts["connection"] = connection
			}
			r, err := importFromCamera(c, input, filepath.Join(output, projectName), dateFormat, bufferSize, prefix, dateRange, customCameraOpts)
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
		color.Red("No camera selected.")
	},
}

func init() {
	rootCmd.AddCommand(importCmd)

	importCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose")
	importCmd.Flags().StringP("input", "i", "", "Input directory for root, eg: E:\\")
	importCmd.Flags().StringP("output", "o", "", "Output directory for sorted media")
	importCmd.Flags().StringP("name", "n", "", "Project name")
	importCmd.Flags().StringP("camera", "c", "", "Camera type")
	importCmd.Flags().StringP("date", "d", "dd-mm-yyyy", "Date format, dd-mm-yyyy by default")
	importCmd.Flags().IntP("buffer", "b", 1000, "Buffer size for copying, default is 1000 bytes")
	importCmd.Flags().StringP("prefix", "p", "", "Prefix for each file, pass `cameraname` to prepend the camera name (eg: Hero9 Black)")
	importCmd.Flags().StringSlice("range", []string{}, "A date range, eg: 01-05-2020,05-05-2020 -- also accepted: `today`, `yesterday`, `week`")
	importCmd.Flags().StringP("connection", "x", "", "Connexion type: `sd_card`, `connect` (GoPro-specific)")

	// GoPro-specific options

	importCmd.Flags().BoolP("skip_aux", "s", true, "GoPro: skip auxiliary files (THM, LRV)")
	importCmd.Flags().StringSlice("sort_by", []string{"camera", "days"}, "GoPro: sort files by: `camera` and/or `days`")
/*
	for _, item := range []string{
		"output", "camera",
	} {
		importCmd.MarkFlagRequired(item)
	}
        */

}

func importFromCamera(c utils.Camera, input string, output string, dateFormat string, bufferSize int, prefix string, dateRange []string, camOpts map[string]interface{}) (*utils.Result, error) {
	switch c {
	case utils.GoPro:
		return gopro.Import(input, output, dateFormat, bufferSize, prefix, dateRange, camOpts)
	case utils.DJI:
		return dji.Import(input, output, dateFormat, bufferSize, prefix, dateRange)
	case utils.Insta360:
		return insta360.Import(input, output, dateFormat, bufferSize, prefix, dateRange)
	case utils.Android:
		return android.Import(input, output, dateFormat, bufferSize, prefix, dateRange)
	default:
		return nil, errors.New("Unsupported camera!")
	}
}
