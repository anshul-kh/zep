package cmd

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/anshul-kh/zep-core/internal/helper"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats id | --name [name]",
	Short: "stats id | --name [name]",
	Long:  "stats command show the resource usage of a particular running process.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cobra *cobra.Command, args []string) {
		argLen := len(args)

		if argLen <= 0 && name == "" {
			fmt.Print("invalid command:-\nstats command requires id or --name (flag) of the process")
			return
		}

		var typ string
		var q string

		switch argLen {
		case 0:
			if name == "" {
				fmt.Print("invalid command:-\nstats command requires id or --name (flag) of the process")
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

		res, err := http.Get(fmt.Sprintf("%s/api/showStats?q=%s&type=%s", BaseURL, q, typ))

		if err != nil {
			log.Printf("error from deamon:%v", err.Error())
			return
		}

		helper.HandleStatsResponse(res)
	},
}
