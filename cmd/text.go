package cmd

import (
	"fmt"
	"log"

	"github.com/bunyk/gokeybr/app"
	"github.com/bunyk/gokeybr/phrase"
	"github.com/spf13/cobra"
)

var textCmd = &cobra.Command{
	Use:   "text [flags] [file with text (\"-\" - stdin)]",
	Short: "train to type contents of some file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		text, err := phrase.FromFile(args[0], 0, wordsCount)
		if err != nil {
			log.Fatal(err)
		}
		a := app.New(text)
		err = a.Run()
		if err != nil {
			log.Fatal(err)
		}
		saveStats(a, false)
		err = phrase.UpdateFileProgress(args[0], a.LinesTyped())
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	textCmd.Flags().IntVarP(&params.PhraseLength, "length", "l", 0,
		"Minimal lenght of text to train on (default 100 for random text, unlimited for loaded)",
	)
	textCmd.Flags().IntVarP(&params.Offset, "offset", "o", -1,
		"Offset in lines when loading file (default 0)",
	)
	rootCmd.AddCommand(textCmd)
}
