package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/anshul-kh/zep-core/internal/helper"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list",
	Long:  "list command shows a list running processes.",
	Args:  cobra.MaximumNArgs(0),
	Run: func(cobra *cobra.Command, args []string) {

		res, err := http.Get(fmt.Sprintf("%s/api/listProcs", BaseURL))

		if err != nil {
			log.Printf("error from deamon:%v", err.Error())
			return
		}

		helper.HandleProcListResponse(res)
	},
}
