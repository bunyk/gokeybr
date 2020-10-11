package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/bunyk/gokeybr/app"
	"github.com/bunyk/gokeybr/stats"
)

// Parameters define arguments with which program started
type Parameters struct {
	PhraseLength int // default lenght for generated phrase
	Offset       int // offset of file in lines
}

var params Parameters

var rootCmd = &cobra.Command{
	Use:  "gokeybr",
	Long: Help,
	Run: func(cmd *cobra.Command, args []string) {
		text, err := stats.GenerateTrainingSession(params.PhraseLength)
		if err != nil {
			log.Fatal(err)
		}
		a := app.New(text)
		err = a.Run()
		if err != nil {
			log.Fatal(err)
		}

		saveStats(a, true)
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
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
