package cmd

import (
	"fmt"

	"github.com/bunyk/gokeybr/app"
	"github.com/bunyk/gokeybr/stats"
	"github.com/spf13/cobra"
)

var markovLength int

var markovCmd = &cobra.Command{
	Use:     "random [flags]",
	Aliases: []string{"markov"},
	Short:   "train on text generated by markov chains",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if markovLength < stats.MinSessionLength {
			fmt.Printf("Sequence should be at least %d characters long\n", stats.MinSessionLength)
		}
		text, err := stats.RandomTraining(markovLength)
		if err != nil {
			fmt.Println(err)
			return
		}
		a := app.New(text)
		err = a.Run()
		if err != nil {
			fmt.Println(err)
			return
		}

		saveStats(a, true)
	},
}

func init() {
	markovCmd.Flags().IntVarP(&markovLength, "length", "l", 100,
		"Minimal lenght in characters of generated text (default 100)",
	)
	rootCmd.AddCommand(markovCmd)
}
