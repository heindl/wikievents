// Copyright (c) 2018 Parker Heindl. All rights reserved.
//
// Use of this source code is governed by the MIT License.
// Read LICENSE.md in the project root for information.

package endpoint

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TODO: Mock endpoint with test files.

func TestBinding(t *testing.T) {
	b := Binding{
		"coordinate": {
			DataType: "http://www.opengis.net/ont/geosparql#wktLiteral",
			Type:     "literal",
			Value:    "Point(14.2 41.1)",
		},
	}
	lat, lng, err := b.MustCoordinates("coordinate")
	if err != nil {
		t.Errorf(err.Error())
	}
	if lat != 41.1 {
		t.Errorf("Latitude should be %f, not %+v", 41.1, lat)
	}
	if lng != 14.2 {
		t.Errorf("Longitude should be %f, not %+v", 14.2, lng)
	}
}

func TestSparQLRequestFormat(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	s, err := parseTemplate("sparql/test.sparql", nil)
	assert.NoError(t, err)

	q := &query{
		Endpoint: "https://query.wikidata.org/sparql",
		Body:     s,
	}

	res, err := q.request()
	assert.NoError(t, err)

	if len(res.Results.Bindings) != 10 {
		t.Errorf("Expected sparql query to have %d results, rather than %d", 10, len(res.Results.Bindings))
	}
}

func TestEventRequest(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	total := 0

	callback := func(b *Binding) error {
		total++
		return nil
	}
	assert.NoError(t, RequestWikidataEvents(-2, 2, callback))

	assert.Equal(t, 9397, total)
}
