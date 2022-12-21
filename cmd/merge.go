package cmd

import (
	"github.com/erdaltsksn/cui"
	"github.com/konradit/mmt/pkg/videomanipulation"
	"github.com/spf13/cobra"
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge two or more videos together",
	Run: func(cmd *cobra.Command, args []string) {
		videoMan := videomanipulation.New()
		videos := getFlagSlice(cmd, "input")
		err := videoMan.Merge(videos...)
		if err != nil {
			cui.Error(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(mergeCmd)
	mergeCmd.Flags().StringSlice("input", []string{}, "File to merge")
}
