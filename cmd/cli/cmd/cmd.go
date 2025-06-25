package cmd

import (
	"fmt"
	"os"
)

const (
	BaseURL = "http://localhost:3746"
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("failed to run command:%v", err)
		os.Exit(0)
	}
}
