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
	Mode         string // Treat training text as lines, or set of words to create random texts
	PhraseLength int    // default lenght for generated phrase
	Offset       int    // Offset in lines for file read
}

func (p Parameters) Validate() error {
	if len(p.Sourcefile) > 0 && len(p.Sourcetext) > 0 {
		return fmt.Errorf("choose source file or sourcetext, but not both")
	}
	if p.Offset > 0 && p.Mode != "lines" {
		return fmt.Errorf("Offset is used only with lines mode")
	}
	return nil
}

func Execute() {
	params := Parameters{}

	var rootCmd = &cobra.Command{
		Use:  "gokeybr",
		Long: Help,
		Run: func(cmd *cobra.Command, args []string) {
			if err := params.Validate(); err != nil {
				log.Fatal(err)
			}
			text, isTraining, err := phrase.FetchPhrase(
				params.Sourcefile, params.Sourcetext,
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
			fmt.Println(a.Summary())

			if len(params.Sourcefile) > 0 && params.Mode == "lines" {
				err = phrase.UpdateFileProgress(params.Sourcefile, a.LinesTyped())
				if err != nil {
					log.Fatal(err)
				}
			}
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
	pf.StringVarP(&params.Mode, "mode", "m", "lines",
		"mode in which to use source text: lines, words or stats",
	)
	pf.StringVarP(&params.Sourcetext, "text", "t", "",
		"source text to train on",
	)
	pf.IntVarP(&params.PhraseLength, "length", "l", 0,
		"Minimal lenght of text to train on (default 100 for random text, unlimited for loaded)",
	)
	pf.IntVarP(&params.Offset, "offset", "o", -1,
		"Offset in lines when loading file (default 0)",
	)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
