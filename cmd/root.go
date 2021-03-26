package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/bunyk/gokeybr/app"
	"github.com/bunyk/gokeybr/stats"
)

var zen bool
var mute bool
var minSpeed int
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

func fatal(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Execute() {
	pf := rootCmd.PersistentFlags()
	pf.BoolVarP(&zen, "zen", "z", false, "run training session in \"zen mode\" (minimal screen output)")
	pf.BoolVarP(&mute, "mute", "m", false, "Do not produce sound when wrong key is hit")
	pf.IntVarP(&minSpeed, "min-speed", "s", 0, "Minimal speed limit in WPM")
	fatal(rootCmd.Execute())
}
