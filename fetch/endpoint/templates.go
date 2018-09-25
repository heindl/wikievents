// Copyright (c) 2018 Parker Heindl. All rights reserved.
//
// Use of this source code is governed by the MIT License.
// Read LICENSE.md in the project root for information.

package endpoint

import (
	"bytes"
	"io/ioutil"
	"text/template"

	"github.com/phogolabs/parcello"
	"github.com/pkg/errors"
)

//go:generate parcello -r -i *.go -i .DS_Store -i dbpedia* -i *test_*

func parseTemplate(queryFile string, templateStruct interface{}) (string, error) {

	file, err := parcello.Open(queryFile)
	if err != nil {
		return "", errors.Wrapf(err, "could not find sparql file %s", queryFile)
	}

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return "", errors.Wrapf(err, "could not read sparql file %s", queryFile)
	}

	query := bytes.NewBuffer([]byte{})
	tmpl, err := template.New("").Parse(string(b))
	if err != nil {
		return "", errors.Wrapf(err, "could not parse sparql template %s", queryFile)
	}
	if err := tmpl.Execute(query, templateStruct); err != nil {
		return "", errors.Wrapf(err, "could not execute sparql template %s", queryFile)
	}

	return query.String(), nil
}
