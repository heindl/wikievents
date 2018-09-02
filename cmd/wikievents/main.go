package main

import (
	"fmt"
	"github.com/heindl/wikievents"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	var StartAtYear int
	var EndAtYear int
	var Verbose bool

	var rootCmd = &cobra.Command{
		Use:   "wikievents",
		Short: "Fetch a graph of events from Wikidata",
		RunE: func(cmd *cobra.Command, args []string) error {
			if Verbose {
				logrus.SetLevel(logrus.DebugLevel)
				logrus.SetFormatter(&logrus.JSONFormatter{})
			}
			events, err := wikievents.FetchEvents(wikievents.RequestParams{
				StartAtYear: StartAtYear,
				EndAtYear:   EndAtYear,
			})
			if err != nil {
				return err
			}
			directoryPath, err := events.WriteToTemp()
			if err != nil {
				return err
			}
			fmt.Println(directoryPath)
			return nil
		},
	}

	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().IntVarP(&StartAtYear, "start", "s", 0, "year to start event query")
	rootCmd.PersistentFlags().IntVarP(&EndAtYear, "end", "e", 0, "year to start event query")

	rootCmd.Execute()
}
