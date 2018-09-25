// Copyright (c) 2018 Parker Heindl. All rights reserved.
//
// Use of this source code is governed by the MIT License.
// Read LICENSE.md in the project root for information.

package fetch

import (
	"io"

	"github.com/heindl/wikivents/fetch/endpoint"
	"github.com/heindl/wikivents/fetch/parse"
	"github.com/pkg/errors"
)

func WikidataEvents(startYear, endYear int, rdfWriter, schemaWriter io.Writer) error {
	if (startYear == 0 && endYear == 0) || (endYear-startYear < 0) {
		return errors.New("valid start and end year required")
	}
	writer := parse.NewWriter(rdfWriter, schemaWriter)
	return endpoint.RequestWikidataEvents(startYear, endYear, writer.ParseBinding)
}
