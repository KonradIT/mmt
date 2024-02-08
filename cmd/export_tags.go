package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/erdaltsksn/cui"
	"github.com/fatih/color"
	"github.com/konradit/mmt/pkg/gopro"
	"github.com/spf13/cobra"
)

func tagAsDuration(tag int, increase bool) string {
	seconds := tag / int(time.Microsecond)
	if increase {
		seconds++
	}
	return fmt.Sprintf("01:00:%d:%03d", seconds, 0)
}

func exportCSV(tags gopro.HiLights, output string) error {
	csvFile, err := os.Create(output)
	if err != nil {
		return err
	}
	defer csvFile.Close()
	writer := csv.NewWriter(csvFile)
	_ = writer.Write([]string{
		"timestamps",
	})
	for _, timestamp := range tags.Timestamps {
		_ = writer.Write([]string{
			fmt.Sprintf("%d", timestamp),
		})
	}
	writer.Flush()
	return writer.Error()
}

func exportJSON(tags gopro.HiLights, output string) error {
	b, err := json.MarshalIndent(tags, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(output, b, 0o600)
}

func exportEDL(name string, tags gopro.HiLights, output string) error {
	content := `TITLE: Timeline 1
FCM: NON-DROP FRAME
`
	for index, tag := range tags.Timestamps {
		content = fmt.Sprintf("%s\n%03d  AX       V     C        %s %s %s %s\n* FROM CLIP NAME: %s\n", content, index, "00:00:00:00", "00:00:00:01", tagAsDuration(tag, false), tagAsDuration(tag, true), name)
	}
	return os.WriteFile(output, []byte(content), 0o600)
}

func extractIndividual(input, output, format string) (int, error) {
	hilights, err := gopro.GetHiLights(input)
	if err != nil {
		return 0, err
	}

	switch format {
	case "csv":
		if output == "" {
			output = strings.Replace(input, filepath.Ext(input), ".csv", -1)
		}
		err = exportCSV(*hilights, output)
	case "json":
		if output == "" {
			output = strings.Replace(input, filepath.Ext(input), ".json", -1)
		}
		err = exportJSON(*hilights, output)
	case "edl":
		if output == "" {
			output = strings.Replace(input, filepath.Ext(input), ".edl", -1)
		}
		err = exportEDL(filepath.Base(input), *hilights, output)
	}
	return hilights.Count, err
}

var exportTags = &cobra.Command{
	Use:   "export-tags",
	Short: "Export HiLight/other tags in video",
	Run: func(cmd *cobra.Command, args []string) {
		input := getFlagString(cmd, "input", "")
		format := getFlagString(cmd, "format", "")
		output := getFlagString(cmd, "output", "")

		stat, err := os.Stat(input)
		if err != nil {
			cui.Error(err.Error())
		}

		if stat.IsDir() {
			files, err := ioutil.ReadDir(input)
			if err != nil {
				cui.Error(err.Error())
			}

			for _, file := range files {
				actualFilename := filepath.Join(input, file.Name())
				count, err := extractIndividual(actualFilename, output, format)
				if err != nil {
					cui.Error(err.Error())
				}
				color.Green(">> Successfully extracted %d tags", count)
			}
		}

		if !stat.IsDir() && filepath.Ext(input) == ".MP4" {
			count, err := extractIndividual(input, output, format)
			if err != nil {
				cui.Error(err.Error())
			}
			color.Green(">> Successfully extracted %d tags", count)
		}
	},
}

func init() {
	rootCmd.AddCommand(exportTags)
	exportTags.Flags().StringP("input", "i", "", "MP4 File or directory with MP4 files")
	exportTags.Flags().StringP("format", "f", "", "formats supported: edl/json/csv")
	exportTags.Flags().StringP("output", "o", "", "Output file, do not specify to use the input file as reference")

	_ = exportTags.MarkFlagRequired("format")
	_ = exportTags.MarkFlagRequired("input")
}
