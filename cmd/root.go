package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/bunyk/gokeybr/app"
	"github.com/bunyk/gokeybr/models"
)

const Help = `
gokeybr is a touch-typing training program

Examples: 

   How to run to train your bash commands typing speed, customized for your commands:

       history | go run main.go -c -f -

Key bindings:

   ESC   quit
   C-F   skip forward to the next phrase
   C-R   toggle repeat phrase mode
`

func Execute() {
	params := models.Parameters{}

	var rootCmd = &cobra.Command{
		Use:  "gokeybr",
		Long: Help,
		Run: func(cmd *cobra.Command, args []string) {
			a, err := app.New(params)
			if err != nil {
				app.Exit(1, err.Error())
			}
			a.Run()
		},
	}
	pf := rootCmd.PersistentFlags()
	pf.StringVarP(&params.Sourcefile, "source", "f", "/usr/share/dict/words",
		"path to file with source text",
	)
	pf.BoolVarP(&params.Codelines, "code", "c", false,
		"treat -f FILE as lines of code",
	)
	pf.StringVarP(&params.Sourcetext, "text", "t", "",
		"source text to train on",
	)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
