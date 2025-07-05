package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"

	"github.com/anshul-kh/zep-core/internal/helper"
	"github.com/spf13/cobra"
)

var shell_command bool

const (
	SHELL_BIN = "/usr/bin"
)

type RunCommand struct {
	BinPath string   `json:"binaryPath"`
	Args    []string `json:"args"`
	Name    string   `json:"name"`
}

var runCmd = &cobra.Command{
	Use:   "run [binaryPath] ...args",
	Short: "run command :- runs and manage the binary using the binary paths",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		binPath := args[0]

		var err error
		if !shell_command {
			binPath, err = helper.GetAbsPath(binPath)
			if err != nil {
				fmt.Printf("invalid file path:%v", err)
				return
			}
		} else {
			// Find the binary in PATH
			binPath, err = exec.LookPath(binPath)
			if err != nil {
				fmt.Printf("shell command not found in PATH: %s\n", binPath)
				return
			}
		}

		payload := RunCommand{
			BinPath: binPath,
			Args:    args[1:],
			Name:    name,
		}

		jsonPayload, err := helper.JSONToBuffer(payload)
		if err != nil {
			log.Panicf("error : %v", err)
			return
		}

		res, err := http.Post(fmt.Sprintf("%s/api/newProc", BaseURL), "application/json", &jsonPayload)

		if err != nil {
			log.Printf("error from deamon:%v", err.Error())
			return
		}

		helper.HandleFinalResponse(res)
	},
}

func init() {
	runCmd.Flags().BoolVar(&shell_command, "shell-cmd", false, "runs shell commands")
}
