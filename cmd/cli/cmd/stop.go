package cmd

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/anshul-kh/zep-core/internal/helper"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop id | --name [name]",
	Short: "stop id | --name [name]",
	Long:  "stop command can kill any process running using it's id or name, if both passed id will be prefered/chosen for killing process",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cobra *cobra.Command, args []string) {
		argLen := len(args)

		if argLen <= 0 && name == "" {
			fmt.Print("invalid command:-\nstop command requires id or --name (flag) of the process")
			return
		}

		var typ string
		var q string

		switch argLen {
		case 0:
			if name == "" {
				fmt.Print("invalid command:-\nstop command requires id or --name (flag) of the process")
				return
			}
			typ = "name"
			q = name
		case 1:
			id := args[0]
			zid, err := strconv.Atoi(id)
			if err != nil {
				fmt.Printf("invalid id:%s", id)
				return
			}
			q = fmt.Sprintf("%d", zid)
			typ = "id"
		default:
			fmt.Print("invalid command:-\nstop command requires id or --name (flag) of the process")
			return
		}

		res, err := http.Post(fmt.Sprintf("%s/api/killProc?q=%s&type=%s", BaseURL, q, typ), "", nil)

		if err != nil {
			log.Printf("error from deamon:%v", err.Error())
			return
		}

		helper.HandleFinalResponse(res)
	},
}
