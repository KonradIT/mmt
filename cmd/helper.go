package cmd

import (
	"strconv"

	"github.com/erdaltsksn/cui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func getFlagString(cmd *cobra.Command, name string, defaultString string) string {
	value, err := cmd.Flags().GetString(name)
	if err != nil {
		cui.Error("Problem parsing "+name, err)
	}
	if value == "" {
		value = viper.GetString(name)
	}
	if value == "" {
		value = defaultString
	}
	return value
}

func getFlagSlice(cmd *cobra.Command, name string) []string {
	value, err := cmd.Flags().GetStringSlice(name)
	if err != nil {
		cui.Error("Problem parsing "+name, err)
	}
	if len(value) == 0 {
		value = viper.GetStringSlice(name)
	}
	return value
}

func getFlagInt(cmd *cobra.Command, name string, defaultInt string) int {
	value, err := cmd.Flags().GetString(name)
	if err != nil {
		cui.Error("Problem parsing "+name, err)
	}
	if value == "" {
		value = viper.GetString(name)
	}
	if value == "" {
		value = defaultInt
	}
	int1, err := strconv.Atoi(value)
	if err != nil {
		cui.Error("Problem parsing "+value, err)
	}
	return int1
}

func getFlagBool(cmd *cobra.Command, name string, defaultBool string) bool {
	value, err := cmd.Flags().GetString(name)
	if err != nil {
		cui.Error("Problem parsing "+name, err)
	}
	if value == "" {
		value = viper.GetString(name)
	}
	if value == "" {
		value = defaultBool
	}
	bool1, err := strconv.ParseBool(value)
	if err != nil {
		cui.Error("Problem parsing "+value, err)
	}
	return bool1
}
