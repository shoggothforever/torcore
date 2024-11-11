/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/spf13/viper"
	"os"

	"github.com/spf13/cobra"
)

var v *viper.Viper = viper.New()
var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cmd",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application.` + cfgFile,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	//cobra.OnInitialize(initConfig)
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "cfgFile", "c", "./", "config.yaml path")
}

// 初始化配置函数
func initConfig() {
	if cfgFile != "" {
		// 如果指定了 config 参数，使用用户指定的配置文件
		v.AddConfigPath(cfgFile)
		fmt.Println("get config path ", cfgFile)
	} else {
		// 默认使用当前目录下的 config.yaml 文件
		v.AddConfigPath(".")
	}
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	// 读取环境变量（可选）
	v.AutomaticEnv()
	// 读取配置文件
	if err := v.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", v.ConfigFileUsed())
	} else {
		fmt.Println("Error reading config file:", err)
	}
}
