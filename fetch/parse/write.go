// Copyright (c) 2018 Parker Heindl. All rights reserved.
//
// Use of this source code is governed by the MIT License.
// Read LICENSE.md in the project root for information.

package parse

import (
	"io"
	"sync"

	"github.com/heindl/wikivents/fetch/endpoint"
)

type Writer struct {
	schema *schema
	rdf    *rdf
}

func NewWriter(rdfWriter, schemaWriter io.Writer) *Writer {
	return &Writer{
		schema: &schema{
			m:      new(sync.Map),
			writer: schemaWriter,
		},
		rdf: &rdf{
			m:      new(sync.Map),
			writer: rdfWriter,
		},
	}
}

func (w *Writer) ParseBinding(b *endpoint.Binding) error {

	p := parser{b}
	object, err := p.Entity("object")
	if err != nil || object == nil {
		return err
	}
	if err := object.Write(w.rdf, w.schema); err != nil {
		return err
	}
	value, err := p.Value()
	if err != nil || value == nil {
		return err
	}
	return value.Write(object, w.rdf, w.schema)

}
