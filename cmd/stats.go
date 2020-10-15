package cmd

import (
	"fmt"

	"github.com/bunyk/gokeybr/stats"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "show statistics report about your typing",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		text, err := stats.GetReport()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(text)
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
