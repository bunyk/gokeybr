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
	Sourcetext   string // Training text itself (optional)
	Mode         string // Treat training text as paragraphs, or set of words to create random texts
	PhraseLength int    // default lenght for generated phrase
}

func Execute() {
	params := Parameters{}

	var rootCmd = &cobra.Command{
		Use:  "gokeybr",
		Long: Help,
		Run: func(cmd *cobra.Command, args []string) {
			text, isTraining, err := phrase.FetchPhrase(
				params.Sourcefile, params.Sourcetext, params.Mode, params.PhraseLength,
			)
			if err != nil {
				log.Fatal(err)
			}
			a := app.New(text)
			err = a.Run()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(a.Summary())

			err = stats.SaveSession(
				a.StartedAt,
				a.Text[:a.InputPosition],
				a.Timeline[:a.InputPosition],
				isTraining,
			)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
	pf := rootCmd.PersistentFlags()
	pf.StringVarP(&params.Sourcefile, "file", "f", "",
		"path to file with source text",
	)
	pf.StringVarP(&params.Mode, "mode", "m", "paragraphs",
		"mode in which to use source text: paragraphs, words or stats",
	)
	pf.StringVarP(&params.Sourcetext, "text", "t", "",
		"source text to train on",
	)
	pf.IntVarP(&params.PhraseLength, "length", "l", 50,
		"Lenght of random phrase to train on",
	)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
