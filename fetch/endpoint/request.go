// Copyright (c) 2018 Parker Heindl. All rights reserved.
//
// Use of this source code is governed by the MIT License.
// Read LICENSE.md in the project root for information.

package endpoint

import (
	"math"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type entityURI string

type BindingCallbackFunc func(*Binding) error

func RequestWikidataEvents(startYear, endYear int, callback BindingCallbackFunc) error {

	if startYear == 0 || endYear == 0 {
		return errors.New("start and end year required")
	}

	entityBatches, err := fetchWikidataEntities(startYear, endYear)
	if err != nil {
		return err
	}

	logrus.Infof("received %d entity references from the wikidata.org SPARQL endpoint", len(entityBatches)*entityBatchSize)

	if len(entityBatches) == 0 {
		return nil
	}

	logrus.Infof("requesting complete entity records in %d batches, and this can be slow because wikidata.org heavily rate limits", len(entityBatches))

	lmtr := make(chan struct{}, 5)
	for range [5]struct{}{} {
		lmtr <- struct{}{}
	}
	eg := errgroup.Group{}
	completed := float64(0)
	total := float64(len(entityBatches))
	eg.Go(func() error {
		for _, _eb := range entityBatches {
			<-lmtr
			eb := _eb
			eg.Go(func() error {
				defer func() {
					lmtr <- struct{}{}
				}()
				if err := fetchEntityBatch(eb, callback); err != nil {
					return err
				}
				completed++
				logrus.Infof("%.0f%% returned", math.Round((completed/total)*100))
				return nil
			})
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return err
	}
	logrus.Infof("finished with sparql requests from wikidata.org")
	return nil
}

var classesToIgnore = map[string]struct{}{
	"year":                   {},
	"solar eclipse":          {},
	"list of persons":        {},
	"wikimedia list article": {},
	"decade":                 {},
	"year BC":                {},
}

type queryResponse struct {
	Head struct {
		Link []string `json:"link"`
		Vars []string `json:"vars"`
	} `json:"head"`
	Results struct {
		Distinct bool       `json:"distinct"`
		Ordered  bool       `json:"ordered"`
		Bindings []*Binding `json:"bindings"`
	}
}

// "http://dbpedia.org/sparql"
// TODO: For smaller queries this is fine, but ensure this isn't paginated.
func fetchWikidataEntities(yearStart int, yearEnd int) ([][entityBatchSize]entityURI, error) {
	s, err := parseTemplate("sparql/dated-entities.sparql", &struct {
		YearEnd   int
		YearStart int
	}{yearEnd, yearStart})
	if err != nil {
		return nil, err
	}

	q := &query{
		Endpoint: "https://query.wikidata.org/sparql",
		Body:     s,
	}
	requestResponse, err := q.request()
	if err != nil {
		return nil, err
	}

	// De-duplicate ..
	entities := map[entityURI]struct{}{}
	for _, binding := range requestResponse.Results.Bindings {

		if _, ok := classesToIgnore[strings.ToLower(binding.String("instanceOfLabel"))]; ok {
			continue
		}
		for _, _e := range strings.Split(binding.String("entities"), " ") {
			e := _e
			entities[entityURI(e)] = struct{}{}
		}
	}

	if len(entities) == 0 {
		return nil, nil
	}

	batchCount := int(math.Ceil(float64(len(entities)) / entityBatchSize))

	batchArray := make([][entityBatchSize]entityURI, batchCount)
	batch := 0
	index := 0
	for e := range entities {
		batchArray[batch][index] = e
		index++
		if index >= entityBatchSize {
			batch++
			index = 0
		}
	}

	return batchArray, nil
}

func fetchEntityBatch(entities [entityBatchSize]entityURI, callback BindingCallbackFunc) error {

	s, err := parseTemplate("sparql/entity.sparql", &struct {
		Entities [entityBatchSize]entityURI
	}{Entities: entities})
	if err != nil {
		return err
	}
	q := &query{
		Endpoint: "https://query.wikidata.org/sparql",
		Body:     s,
	}
	requestResponse, err := q.request()
	if err != nil {
		return err
	}
	if requestResponse == nil {
		return nil
	}

	// TODO: Attempted to run this as an errgroup with a new routine for each callback,
	// but the final test count was inconsistent, sometimes dramatically.
	// The reason may be that a new callback is not being allocated for every go routine, in
	// the same way the range variable has to be instantiated on the local scope, but need to
	// learn what is happening here.
	for _, b := range requestResponse.Results.Bindings {
		if err := callback(b); err != nil {
			return err
		}
	}
	return nil
}
