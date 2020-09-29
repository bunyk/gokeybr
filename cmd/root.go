package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/bunyk/gokeybr/app"
)

const Help = `
gokeybr is a touch-typing training program

Examples: 

   How to run to train your bash commands typing speed, customized for your commands:

       history | go run main.go -f -

   Or train to type random quote, like on typeracer:
   
       fortune | go run main.go -l 1000 -f -

Key bindings:

   ESC   quit
   C-F   skip forward to the next phrase
`

func Execute() {
	params := app.Parameters{}

	var rootCmd = &cobra.Command{
		Use:  "gokeybr",
		Long: Help,
		Run: func(cmd *cobra.Command, args []string) {
			a, err := app.New(params)
			if err != nil {
				log.Fatal(err)
			}
			a.Run()
		},
	}
	pf := rootCmd.PersistentFlags()
	pf.StringVarP(&params.Sourcefile, "file", "f", "",
		"path to file with source text",
	)
	pf.StringVarP(&params.Mode, "mode", "m", "paragraphs",
		"mode in which to use source text: paragraphs or words ",
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
