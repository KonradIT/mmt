package cmd

import (
	"strconv"

	"github.com/erdaltsksn/cui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func get_flag_string(cmd *cobra.Command, name string) string {
	value, err := cmd.Flags().GetString(name)
        if err != nil {
                cui.Error("Problem parsing "+name, err)
        }
        if value == "" {
          value = viper.GetString(name)
        }
        return value
}

func get_flag_slice(cmd *cobra.Command, name string) []string {
	value, err := cmd.Flags().GetStringSlice(name)
        if err != nil {
                cui.Error("Problem parsing "+name, err)
        }
        if len(value) == 0 {
          value = viper.GetStringSlice(name)
        }
        return value
}

func get_flag_int(cmd *cobra.Command, name string, default_int string) int {
	value, err := cmd.Flags().GetString(name)
        if err != nil {
                cui.Error("Problem parsing "+name, err)
        }
        if value == "" {
          value = viper.GetString(name)
        }
        if value == "" {
          value = default_int
        }
        int1, err := strconv.Atoi(value)
        return int1
}

func get_flag_bool(cmd *cobra.Command, name string, default_bool string) bool {
	value, err := cmd.Flags().GetString(name)
        if err != nil {
                cui.Error("Problem parsing "+name, err)
        }
        if value == "" {
          value = viper.GetString(name)
        }
        if value == "" {
          value = default_bool
        }
        bool1, err := strconv.ParseBool(value)
        return bool1
}

