package cmd

import (
	"fmt"
	"log"

	"github.com/bunyk/gokeybr/app"
	"github.com/bunyk/gokeybr/phrase"
	"github.com/bunyk/gokeybr/stats"
	"github.com/spf13/cobra"
)

var wordsFile string
var wordsCount int

var wordsCmd = &cobra.Command{
	Use:   "words",
	Short: "train to type words loaded from file",
	Run: func(cmd *cobra.Command, args []string) {
		if wordsCount < 1 {
			fmt.Println("Need more then one word to start exercise")
			return
		}
		text, err := phrase.Words(wordsFile, wordsCount)
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
			false,
		)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	wordsCmd.Flags().StringVarP(&wordsFile, "file", "f", "/usr/share/dict/words", "File to load words from (one word per line). \"-\" for stdin.")

	wordsCmd.Flags().IntVarP(&wordsCount, "number", "n", 10,
		"Number of words to type (default 10)",
	)
	rootCmd.AddCommand(wordsCmd)
}
