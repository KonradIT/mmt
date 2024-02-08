package cmd

import (
	"bufio"
	"fmt"
	"image/jpeg"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/erdaltsksn/cui"
	"github.com/fatih/color"
	"github.com/nfnt/resize"
	"github.com/spf13/cobra"
	"github.com/wayneashleyberry/lut/pkg/cubelut"
)

func applyLUTToFile(sourceFilename, lutFilename string, intensity float64, quality int, resizeTo string) error {
	lutFile, err := os.Open(lutFilename)
	if err != nil {
		return err
	}
	defer lutFile.Close()

	sourceFile, err := os.Open(sourceFilename)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	srcImg, err := jpeg.Decode(sourceFile)
	if err != nil {
		return err
	}

	lutReader := bufio.NewReader(lutFile)

	cubefile, err := cubelut.Parse(lutReader)
	if err != nil {
		return err
	}

	img, err := cubefile.Apply(srcImg, intensity)
	if err != nil {
		return err
	}

	resolution := regexp.MustCompile(`\d+x\d+`)

	if resizeTo != "" && resolution.MatchString(resizeTo) {
		width := strings.Split(resizeTo, "x")[0]
		height := strings.Split(resizeTo, "x")[1]

		parsedWidth, err := strconv.ParseUint(width, 10, 32)
		if err != nil {
			return err
		}
		parsedHeight, err := strconv.ParseUint(height, 10, 32)
		if err != nil {
			return err
		}

		img = resize.Resize(uint(parsedWidth), uint(parsedHeight), img, resize.Lanczos3)
	}
	destinationFile, err := os.Create(
		filepath.Join(filepath.Dir(sourceFilename),
			fmt.Sprintf("%s %s%s",
				strings.Replace(filepath.Base(sourceFilename), filepath.Ext(sourceFilename), "", -1),
				strings.Replace(filepath.Base(lutFilename), filepath.Ext(lutFilename), "", -1),
				filepath.Ext(sourceFilename),
			),
		),
	)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	return jpeg.Encode(destinationFile, img, &jpeg.Options{
		Quality: quality,
	})
}

var applyLutCmd = &cobra.Command{
	Use:   "apply-lut",
	Short: "Apply LUT to one or more images",
	Run: func(cmd *cobra.Command, args []string) {
		input := getFlagString(cmd, "input", "")
		lutFile := getFlagString(cmd, "lut", "")

		intensity := getFlagInt(cmd, "intensity", "100")
		intensityParsed := float64(intensity) / 100

		quality := getFlagInt(cmd, "quality", "100")
		resizeTo := getFlagString(cmd, "resize", "")

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
				if filepath.Ext(file.Name()) != ".JPG" {
					continue
				}
				color.Yellow(">> Applying LUT to: %s...", file.Name())
				err = applyLUTToFile(actualFilename, lutFile, intensityParsed, quality, resizeTo)
				if err != nil {
					color.Red(err.Error())
					continue
				}
				color.Green(">> Successfully applied LUT to: %s", file.Name())
			}
		}

		if !stat.IsDir() && filepath.Ext(input) == ".JPG" {
			color.Yellow(">> Applying LUT to: %s...", input)
			err = applyLUTToFile(input, lutFile, intensityParsed, quality, resizeTo)
			if err != nil {
				color.Red(err.Error())
			}
			color.Green(">> Successfully applied LUT to: %s", input)
		}
	},
}

func init() {
	rootCmd.AddCommand(applyLutCmd)
	applyLutCmd.Flags().StringP("input", "i", "", "JPG File or Directory with JPG files")
	applyLutCmd.Flags().StringP("lut", "l", "", "Path to LUT file, only .CUBE supported")
	applyLutCmd.Flags().String("intensity", "", "Intensity of filter from 1 - 100")
	applyLutCmd.Flags().String("quality", "", "JPG quality (max + default: 100)")
	applyLutCmd.Flags().String("resize", "", "Resize image ([width]x[height)")
}
