// Copyright (c) 2018 Parker Heindl. All rights reserved.
//
// Use of this source code is governed by the MIT License.
// Read LICENSE.md in the project root for information.

package parse

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

type entityID string

func newEntityID(s string) (entityID, error) {
	if len(s) == 0 || !strings.Contains(s, "wikidata.org/entity/Q") {
		return entityID(""), errors.Errorf("invalid wikidata entity uri: %s", s)
	}
	return entityID(s[strings.LastIndex(s, "/")+1:]), nil
}

type entity struct {
	ID   entityID
	Name string
	Type string
}

func (Ω *entity) Write(rdf *rdf, schema *schema) error {
	if Ω.Type != "" {
		pred := newPredicate(predicateEntityType, Ω.Type)
		if err := rdf.WriteFeature(Ω.ID, pred, ""); err != nil {
			return err
		}
		if err := schema.Write(pred, schemaTypeUID); err != nil {
			return err
		}
	}
	if Ω.Name != "" {
		pred := newPredicate(predicateFeature, "label")
		if err := rdf.WriteFeature(Ω.ID, pred, escapeFeatureValue(Ω.Name)); err != nil {
			return err
		}
		if err := schema.Write(newPredicate(predicateFeature, "label"), schemaTypeString); err != nil {
			return err
		}
	}
	return nil
}

// Note that it is ok for string Value to be empty if it is an Entity type.
type parsedValue struct {
	stringValue string
	entityValue *entity
	predicate   predicate
	schemaType  schemaType
}

func (Ω *parsedValue) Write(object *entity, rdf *rdf, schema *schema) error {

	if Ω.entityValue != nil {
		if err := Ω.entityValue.Write(rdf, schema); err != nil {
			return err
		}
		if err := rdf.WriteEdge(object.ID, Ω.predicate, Ω.entityValue.ID); err != nil {
			return err
		}
	} else {
		if err := rdf.WriteFeature(object.ID, Ω.predicate, Ω.stringValue); err != nil {
			return err
		}
	}

	return schema.Write(Ω.predicate, Ω.schemaType)
}

type wikibaseOntology string

func newWikibaseOntology(o string) (wikibaseOntology, bool) {
	ontology := wikibaseOntology(o)
	_, ok := knownWikibaseOntologies[ontology]
	return ontology, ok
}

var knownWikibaseOntologies = map[wikibaseOntology]struct{}{
	"http://wikiba.se/ontology#Time":            {},
	"http://wikiba.se/ontology#WikibaseItem":    {},
	"http://wikiba.se/ontology#CommonsMedia":    {},
	"http://wikiba.se/ontology#ExternalId":      {},
	"http://wikiba.se/ontology#Url":             {},
	"http://wikiba.se/ontology#String":          {},
	"http://wikiba.se/ontology#Quantity":        {},
	"http://wikiba.se/ontology#Monolingualtext": {},
	"http://wikiba.se/ontology#GlobeCoordinate": {},
}

type predicate string
type predicateType string

const (
	predicateEdge       = predicateType("e_")
	predicateFeature    = predicateType("f_")
	predicateEntityType = predicateType("t_")
)

var alphaNumeric = regexp.MustCompile("[^a-zA-Z0-9]+")

func newPredicate(pType predicateType, s string) predicate {
	s = strings.ToLower(s)
	s = alphaNumeric.ReplaceAllString(s, "_")
	s = string(pType) + s
	return predicate(s)
}

type schemaType string

const (
	schemaTypeGeo     = schemaType("geo")
	schemaTypeInt     = schemaType("int")
	schemaTypeFloat   = schemaType("float")
	schemaTypeString  = schemaType("string")
	schemaTypeDefault = schemaType("default")
	schemaTypeBool    = schemaType("bool")
	schemaTypeUID     = schemaType("uid")
)

type schema struct {
	m      map[string]struct{}
	writer io.Writer
	sync.Mutex
}

func (Ω *schema) Write(p predicate, t schemaType) error {
	line := fmt.Sprintf("%s: %s .\n", p, t)
	if t == schemaTypeUID {
		line = fmt.Sprintf("%s: %s @reverse .\n", p, t)
	}
	Ω.Lock()
	defer Ω.Unlock()
	if _, ok := Ω.m[line]; !ok {
		Ω.m[line] = struct{}{}
		if _, err := Ω.writer.Write([]byte(line)); err != nil {
			return errors.Wrapf(err, "could not Write [%s, %s]", p, t)
		}
	}
	return nil

}

type rdf struct {
	m      map[string]struct{}
	writer io.Writer
	sync.Mutex
}

func (Ω *rdf) WriteFeature(entityID entityID, predicate predicate, value string) error {
	line := fmt.Sprintf(
		`_:%s <%s> "%s" .`,
		entityID,
		predicate,
		value,
	) + "\n"
	Ω.Lock()
	defer Ω.Unlock()
	if _, ok := Ω.m[line]; !ok {
		Ω.m[line] = struct{}{}
		if _, err := Ω.writer.Write([]byte(line)); err != nil {
			return errors.Wrapf(err, "could not Write [%s] [%s] [%s]", entityID, predicate, value)
		}
	}
	return nil
}

func (Ω *rdf) WriteEdge(object entityID, predicate predicate, subject entityID) error {

	line := fmt.Sprintf(
		"_:%s <%s> _:%s .\n",
		object,
		predicate,
		subject,
	)
	Ω.Lock()
	defer Ω.Unlock()
	if _, ok := Ω.m[line]; !ok {
		Ω.m[line] = struct{}{}
		if _, err := Ω.writer.Write([]byte(line)); err != nil {
			return errors.Wrapf(err, "could not Write [%s, %s, %s]", object, predicate, subject)
		}
	}
	return nil
}
