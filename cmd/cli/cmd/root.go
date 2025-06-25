package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "zep",
	Short: "Zep is a simple and easy-to-use Go process manager.",
	Long: `Zep is a lightweight process manager written in Go.

It is designed to easily run and manage background processes, services,
cron jobs, shell commands, and APIs with minimal setup.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Hello %s\n", "world")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
