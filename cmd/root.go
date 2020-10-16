package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/bunyk/gokeybr/app"
	"github.com/bunyk/gokeybr/stats"
)

var zen bool
var rootCmd = &cobra.Command{
	Use:  "gokeybr",
	Long: Help,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func saveStats(a *app.App, isTraining bool) {
	fmt.Println(a.Summary())
	if err := stats.SaveSession(
		a.StartedAt,
		a.Text[:a.InputPosition],
		a.Timeline[:a.InputPosition],
		isTraining,
	); err != nil {
		fmt.Println(err)
	}
}

func Execute() {
	rootCmd.PersistentFlags().BoolVarP(&zen, "zen", "z", false, "run training session in \"zen mode\" (minimal screen output)")
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
