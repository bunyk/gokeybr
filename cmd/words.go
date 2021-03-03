package cmd

import (
	"fmt"

	"github.com/bunyk/gokeybr/app"
	"github.com/bunyk/gokeybr/phrase"
	"github.com/spf13/cobra"
)

var wordsCount int

var wordsCmd = &cobra.Command{
	Use:   "words [flags] [optional file to load words from (one word per line, \"-\" - stdin)]",
	Short: "train to type words loaded from file",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		if wordsCount < 1 {
			fmt.Println("Need more then one word to start exercise")
			return
		}
		filename := "/usr/share/dict/words"
		if len(args) > 0 {
			filename = args[0]
		}
		text, err := phrase.Words(filename, wordsCount)
		fatal(err)
		a, err := app.New(text)
		fatal(err)
		err = a.Run()
		fatal(err)
		saveStats(a, false)
	},
}

func init() {
	wordsCmd.Flags().IntVarP(&wordsCount, "number", "n", 10,
		"Number of words to type (default 10)",
	)
	rootCmd.AddCommand(wordsCmd)
}
