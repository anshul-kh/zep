package cmd

import (
	"github.com/spf13/cobra"
)

/*
* * Cobra is used to develop the cli interface of zep
* - visit : github.com/spf13/cobra
 */

var name string

var rootCmd = &cobra.Command{
	Use:   "zep",
	Short: "Zep is a simple and easy-to-use Go process manager.",
	Long: `Zep is a lightweight process manager written in Go.

It is designed to easily run and manage background processes, services,
cron jobs, shell commands, and APIs with minimal setup.`,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&name, "name", "", "Name for the process")
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(listCmd)
}
