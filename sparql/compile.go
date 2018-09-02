package sparql

import (
	"bytes"
	"github.com/go-errors/errors"
	"github.com/phogolabs/parcello"
	"github.com/sirupsen/logrus"
	"io"
	"text/template"
)

// "http://dbpedia.org/sparql"
func FetchWikidataEvents(yearStart int, yearEnd int) (*QueryResponse, error) {
	s, err := parseTemplate("wikidata/events.sparql", &struct{
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
	return q.request()
}

//go:generate parcello -r -i *.go -i .DS_Store

func parseTemplate(queryFile string, templateStruct interface{}) (string, error) {

	logrus.WithFields(logrus.Fields{
		"file": queryFile,
		"data": templateStruct,
	}).Debug("Parsing template")

	file, err := parcello.Open(queryFile)
	if err != nil {
		return "", errors.Wrap(err, 0)
	}

	info, err := file.Stat()
	if err != nil {
		return "", errors.Wrap(err, 0)
	}
	if info.Size() == 0 {
		return "", errors.New("Template file is empty: "+queryFile)
	}

	sqarql := bytes.NewBuffer([]byte{})
	if _, err = io.Copy(sqarql, file); err != nil {
		return "", errors.Wrap(err, 0)
	}

	query := bytes.NewBuffer([]byte{})
	tmpl, err := template.New("").Parse(sqarql.String())
	if err := tmpl.Execute(query, templateStruct); err != nil {
		return "", errors.Wrap(err, 0)
	}

	logrus.WithFields(logrus.Fields{
		"query": query.String(),
	}).Debug("Sparql file parsed")

	return query.String(), nil
}
