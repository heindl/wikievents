// Copyright (c) 2018 Parker Heindl. All rights reserved.
//
// Use of this source code is governed by the MIT License.
// Read LICENSE.md in the project root for information.

package endpoint

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type query struct {
	Endpoint string
	Body     string
}

const entityBatchSize = 50

func (Ω *query) genHTTPRequest() (*http.Request, error) {

	req, err := http.NewRequest("GET", Ω.Endpoint, nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not generate new http request")
	}

	q := req.URL.Query()
	q.Add("format", "json")

	q.Add("query", Ω.Body)
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Accept", "application/sparql-results+json")

	return req, nil
}

func (Ω *query) request() (qResponse *queryResponse, responseError error) {

	req, err := Ω.genHTTPRequest()
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "client request failed")
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && responseError == nil {
			responseError = errors.Wrap(closeErr, "could  not close http client response")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == 429 {
			return nil, errors.New("wikidata.org thinks you're making too many requests")
		}
		if resp.StatusCode == 443 {
			return nil, errors.New("a pipe has broken ... ?")
		}
		if resp.StatusCode == 500 {
			return nil, errors.New("error 500, likely meaning the sparql request hit the 60 second timeout")
		}
		return nil, fmt.Errorf(
			"wikidata.org sparql request [%s] failed with status [%s]",
			Ω.Endpoint,
			resp.Status)
	}

	qResponse = &queryResponse{}
	if err := json.NewDecoder(resp.Body).Decode(qResponse); err != nil {
		return nil, errors.Wrap(err, "could not decode wikidata.org json response")
	}

	return qResponse, nil
}
