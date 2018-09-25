// Copyright (c) 2018 Parker Heindl. All rights reserved.
//
// Use of this source code is governed by the MIT License.
// Read LICENSE.md in the project root for information.

package cmd

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/heindl/wikivents/fetch"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const commandName = "wikivents"

var rootCmd = &cobra.Command{
	Use:   fmt.Sprintf("%s [global-flag ...] --start --end", commandName),
	Short: "Fetch events, participants and contextual information from Wikidata.org within an epoch.",
	Long: `
	The program runs a SparQL query against query.wikidata.org for nodes that have a time value within the given epoch.
	
	It condenses the edges into labels and writes them to an [RDF](https://en.wikipedia.org/wiki/Resource_Description_Framework) and schema file in syntax understood by [DGraph](https://docs.dgraph.io/master/query-language/#schema).

	To clarify the edge values further, they are prefixed by:
		- 'f_': a feature with a descriptive literal value.
		- 't_': a type with an empty default value.
		- 'e_': a edge to the uid of another node.
	`,
	Example: fmt.Sprintf(`
		$ %s -o /tmp/ -s -70 -e 300
		$ ls /tmp
		wikivents.schema.gz		wikivents.nt.gz
	`, commandName),
	RunE: process,
}

// flags
var verbose bool
var outputDirectory string
var startYear int
var endYear int

func init() {
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "print debug information")
	rootCmd.Flags().StringVarP(&outputDirectory, "output-directory", "o", ".", "directory path to write compressed RDF files")
	rootCmd.Flags().IntVarP(&startYear, "start-year", "s", 0, "start year for query range")
	rootCmd.Flags().IntVarP(&endYear, "end-year", "e", 0, "end year for query range")
}

func process(cmd *cobra.Command, args []string) (resErr error) {
	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	rdfWriter, rdfCloser, err := gZipWriter(filepath.Join(outputDirectory, "wikivents.nt.gz"))
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := rdfCloser(); closeErr != nil && resErr == nil {
			resErr = closeErr
		}
	}()

	schemaWriter, schemaCloser, err := gZipWriter(filepath.Join(outputDirectory, "wikivents.schema.gz"))
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := schemaCloser(); closeErr != nil && resErr == nil {
			resErr = closeErr
		}
	}()

	return fetch.WikidataEvents(startYear, endYear, rdfWriter, schemaWriter)

}

func Execute() {
	rootCmd.Execute()
}

func gZipWriter(filePath string) (io.Writer, func() error, error) {

	f, err := os.Create(filePath)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "could not create file %s", filePath)
	}
	g := gzip.NewWriter(f)

	return g, func() error {
		if err := g.Close(); err != nil {
			return errors.Wrap(err, "could not close gzip")
		}
		if err := f.Close(); err != nil {
			return errors.Wrapf(err, "could not close file %s", filePath)
		}
		return nil
	}, nil
}
