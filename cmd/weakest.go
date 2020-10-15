package cmd

import (
	"fmt"

	"github.com/bunyk/gokeybr/app"
	"github.com/bunyk/gokeybr/stats"
	"github.com/spf13/cobra"
)

var weakestLength int

var weakestCmd = &cobra.Command{
	Use:   "weakest [flags]",
	Short: "train on sequence of your weakest character combinations",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if weakestLength < stats.MinSessionLength {
			fmt.Printf("Sequence should be at least %d characters long\n", stats.MinSessionLength)
		}
		text, err := stats.WeakestTraining(weakestLength)
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
	weakestCmd.Flags().IntVarP(&weakestLength, "length", "l", 50,
		"Minimal lenght in characters of generated text (default 50)",
	)
	rootCmd.AddCommand(weakestCmd)
}
