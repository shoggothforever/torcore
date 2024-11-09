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
		fmt.Println("marshal ", fileName)
	},
}

func init() {
	rootCmd.AddCommand(marshalCmd)
	marshalCmd.Flags().StringVarP(&fileName, "file", "f", "testFile", "input fileName that contain bencode Object text to marshal into bencode text")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// marshalCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// marshalCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}