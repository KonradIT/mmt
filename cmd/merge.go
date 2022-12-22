package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/erdaltsksn/cui"
	"github.com/konradit/mmt/pkg/utils"
	"github.com/konradit/mmt/pkg/videomanipulation"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge two or more videos together",
	Run: func(cmd *cobra.Command, args []string) {
		videoMan := videomanipulation.New()
		videos := getFlagSlice(cmd, "input")
		totalFrames := 0
		ffprobe := utils.NewFFprobe(nil)
		for i := 0; i < len(videos); i++ {
			head, err := ffprobe.Frames(videos[i])
			if err != nil {
				cui.Error(err.Error())
			}
			totalFrames += head.Streams[0].Frames
		}

		nonAsync := mpb.New(
			mpb.WithWidth(60),
			mpb.WithRefreshRate(180*time.Millisecond))
		newBar := nonAsync.AddBar(int64(totalFrames),
			mpb.PrependDecorators(
				decor.Name(fmt.Sprintf("%s%s", "🐈", filepath.Base(videos[0]))),
				decor.Percentage(decor.WCSyncSpace),
			),
			mpb.AppendDecorators(
				decor.OnComplete(
					decor.EwmaETA(decor.ET_STYLE_GO, 60, decor.WCSyncWidth), "✔️",
				),
			),
		)

		err := videoMan.Merge(newBar, videos...)
		if err != nil {
			cui.Error(err.Error())
		}
		nonAsync.Wait()
	},
}

func init() {
	rootCmd.AddCommand(mergeCmd)
	mergeCmd.Flags().StringSlice("input", []string{}, "File to merge")
}
