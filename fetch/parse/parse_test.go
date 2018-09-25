// Copyright (c) 2018 Parker Heindl. All rights reserved.
//
// Use of this source code is governed by the MIT License.
// Read LICENSE.md in the project root for information.

package parse

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/heindl/wikivents/fetch/endpoint"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func TestParser(t *testing.T) {

	// Get stored bindings.
	// This file should be updated when the endpoint wiki data changes in endpoint test.
	b, err := ioutil.ReadFile("./testdata/bindings.json")
	assert.NoError(t, err)
	var bindings []*endpoint.Binding
	assert.NoError(t, json.Unmarshal(b, &bindings))

	// Set up mock writers.
	rdfBuffer := bytes.NewBuffer([]byte{})
	schemaBuffer := bytes.NewBuffer([]byte{})

	rdfWriter := bufio.NewWriter(rdfBuffer)
	schemaWriter := bufio.NewWriter(schemaBuffer)

	writer := NewWriter(rdfWriter, schemaWriter)

	// Run concurrently to check Write safety.
	eg := errgroup.Group{}
	eg.Go(func() error {
		for _, _b := range bindings {
			b := _b
			eg.Go(func() error {
				return writer.ParseBinding(b)
			})
		}
		return nil
	})
	assert.NoError(t, eg.Wait())

	assert.NoError(t, rdfWriter.Flush())
	assert.NoError(t, schemaWriter.Flush())

	assert.Equal(t, 5957, len(bytes.Split(rdfBuffer.Bytes(), []byte("\n"))))
	assert.Equal(t, 698, len(bytes.Split(schemaBuffer.Bytes(), []byte("\n"))))
}
