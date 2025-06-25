package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/anshul-kh/zep-core/internal/helper"
	"github.com/spf13/cobra"
)

type RunCommand struct {
	BinPath string   `json:"binaryPath"`
	Args    []string `json:"args"`
}

type SuccessResponse struct {
	Success bool   `json:"success"`
	Msg     string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

var runCmd = &cobra.Command{
	Use:   "run [binaryPath] ...args",
	Short: "run command :- runs and manage the binary using the binary paths",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		binPath := args[0]

		payload := RunCommand{
			BinPath: binPath,
			Args:    args,
		}

		jsonPayload, err := helper.JSONToBuffer(payload)
		if err != nil {
			log.Panicf("error : %v", err)
		}

		res, err := http.Post(fmt.Sprintf("%s/api/newProc", BaseURL), "application/json", &jsonPayload)

		if err != nil {
			log.Panicf("error from master process:%v", err)
		}

		var response SuccessResponse
		err = helper.DecodeJSONFromResponse(res, response)
		if err != nil {
			log.Panicf("error from master process:%v", err)
		}

		fmt.Print(response)
	},
}
