package sparql

import (
	"encoding/json"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"
)

type query struct {
	Endpoint string
	Body     string
}

func safeClose(c io.Closer, err *error) {
	if closeErr := c.Close(); closeErr != nil && *err == nil {
		*err = closeErr
	}
}

func (Ω *query) genHTTPRequest() (*http.Request, error) {
	logrus.WithFields(logrus.Fields{
		"query":    Ω.Body,
		"endpoint": Ω.Endpoint,
	}).Infof("Generating SparQL Request")

	req, err := http.NewRequest("GET", Ω.Endpoint, nil)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	q := req.URL.Query()
	q.Add("format", "json")
	//if options != nil && options.explain {
	//	q.Add("explain", "true")
	//}

	q.Add("query", Ω.Body)
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Accept", "application/sparql-results+json")

	return req, nil
}

var throttle = time.Tick(time.Second / 100)

func (Ω *query) request() (qResponse *QueryResponse, err error) {
	<-throttle

	req, err := Ω.genHTTPRequest()
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	defer safeClose(resp.Body, &err)

	logrus.WithFields(logrus.Fields{
		"status":        resp.Status,
		"statusCode":    resp.StatusCode,
		"contentType":   resp.Header.Get("Content-Type"),
		"contentLength": resp.ContentLength,
	}).Debug("Received get response")

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(
			fmt.Sprintf(
				"Sparql request [%s] failed with status [%s]. Possible reason is that the query reached 60 second timeout",
				Ω.Endpoint,
				resp.Status),
		)
	}

	qResponse = &QueryResponse{}
	if err := json.NewDecoder(resp.Body).Decode(qResponse); err != nil {
		return nil, errors.Wrap(err, 0)
	}

	debugFields := logrus.Fields{
		"head":         qResponse.Head,
		"bindingCount": len(qResponse.Results.Bindings),
	}

	if len(qResponse.Results.Bindings) > 0 {
		debugFields["firstBinding"] = qResponse.Results.Bindings[0]
	}

	logrus.WithFields(debugFields).Debug("JSON response parsed")

	return qResponse, nil
}
