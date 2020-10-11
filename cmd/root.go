package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/bunyk/gokeybr/app"
	"github.com/bunyk/gokeybr/phrase"
	"github.com/bunyk/gokeybr/stats"
)

// Parameters define arguments with which program started
type Parameters struct {
	Sourcefile   string // From where to read training text
	Mode         string // Treat training text as lines, or set of words to create random texts
	PhraseLength int    // default lenght for generated phrase
	Offset       int    // Offset in lines for file read
}

func (p Parameters) Validate() error {
	if p.Offset > 0 && p.Mode != "lines" {
		return fmt.Errorf("Offset is used only with lines mode")
	}
	return nil
}

var params Parameters

var rootCmd = &cobra.Command{
	Use:  "gokeybr",
	Long: Help,
	Run: func(cmd *cobra.Command, args []string) {
		if err := params.Validate(); err != nil {
			log.Fatal(err)
		}
		text, isTraining, err := phrase.FetchPhrase(
			params.Sourcefile,
			params.Mode,
			params.PhraseLength, params.Offset,
		)
		if err != nil {
			log.Fatal(err)
		}
		a := app.New(text)
		err = a.Run()
		if err != nil {
			log.Fatal(err)
		}

		if len(params.Sourcefile) > 0 && params.Mode == "lines" {
			err = phrase.UpdateFileProgress(params.Sourcefile, a.LinesTyped())
			if err != nil {
				log.Fatal(err)
			}
		}
		saveStats(a, isTraining)
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

func init() {

	pf := rootCmd.PersistentFlags()
	pf.StringVarP(&params.Mode, "mode", "m", "lines",
		"mode in which to use source text: lines, words or stats",
	)
	pf.IntVarP(&params.PhraseLength, "length", "l", 0,
		"Minimal lenght of text to train on (default 100 for random text, unlimited for loaded)",
	)
	pf.IntVarP(&params.Offset, "offset", "o", -1,
		"Offset in lines when loading file (default 0)",
	)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
