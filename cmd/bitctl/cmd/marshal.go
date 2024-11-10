/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var fileName string

// marshalCmd represents the marshal command
var marshalCmd = &cobra.Command{
	Use:   "marshal",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Aliases: []string{"m"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("marshal ", fileName, v.GetString("bencode.file"))
	},
}
