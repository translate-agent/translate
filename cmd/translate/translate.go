/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/

package translate

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var TranslateCmd = &cobra.Command{
	Use:   "translate",
	Short: "This command sets configuration",
	Long: `translate command sets configuration from:
			environment variables
			yaml file
			CLI arguments`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(viper.GetViper().GetString("greeting.msg"))
	},
}

func init() {}
