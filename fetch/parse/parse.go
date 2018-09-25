// Copyright (c) 2018 Parker Heindl. All rights reserved.
//
// Use of this source code is governed by the MIT License.
// Read LICENSE.md in the project root for information.

package parse

import (
	"fmt"
	"strings"

	"github.com/heindl/wikivents/fetch/endpoint"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type parser struct {
	binding *endpoint.Binding
}

func escapeFeatureValue(v string) string {
	return strings.Replace(v, `"`, "'", -1)
}

func (Ω *parser) Entity(key string) (*entity, error) {

	if Ω.binding.Type(key) == "bnode" {
		logrus.Debugf("received bNode entity: %s", key)
		return nil, nil
	}

	id, err := Ω.binding.MustString(key)
	if err != nil {
		return nil, errors.Wrapf(err, "expected object propertyLabel [%s], which implies data structure issues", key)
	}

	e := &entity{}
	e.ID, err = newEntityID(id)
	if err != nil {
		return nil, err
	}

	e.Name, err = Ω.binding.MustString(key + "Label")
	if err != nil {
		// Restrict strongly toward well formed entities, but no need to crash the system.
		logrus.Warnf("received incomplete entity: %v", Ω.binding.Values())
		return nil, nil
	}

	// Within subject entities, these are often missing so just need to ignore.
	e.Type = Ω.binding.String(key + "InstanceOfLabel")
	return e, nil
}

func (Ω *parser) label() (string, error) {
	label, err := Ω.binding.MustString("propertyLabel")
	if err != nil {
		return "", errors.Wrap(err, "binding missing propertyLabel, though all should have one")
	}
	if label == "instance of" {
		return "", nil
	}
	return label, nil
}

func (Ω *parser) ontology() (wikibaseOntology, error) {
	ontology, err := Ω.binding.MustString("wikibaseType")
	if err != nil {
		return "", errors.Wrap(err, "binding missing wikibase ontology")
	}

	if _, ok := newWikibaseOntology(ontology); !ok {
		return "", errors.Errorf("unknown ontology: %s", ontology)
	}
	return wikibaseOntology(ontology), nil
}

func (Ω *parser) Value() (*parsedValue, error) {

	label, err := Ω.label()
	if err != nil || label == "" {
		return nil, err
	}

	stringVal := strings.TrimSpace(Ω.binding.String("value"))
	if stringVal == "" {
		return nil, errors.Errorf("empty value [%s, %s]", Ω.binding.String("object"), label)
	}

	if label == "subclass of" {
		return &parsedValue{
			predicate:  newPredicate(predicateEntityType, escapeFeatureValue(stringVal)),
			schemaType: schemaTypeDefault,
		}, nil
	}

	ontology, err := Ω.ontology()
	if err != nil || ontology == "" {
		return nil, err
	}

	switch ontology {
	case "http://wikiba.se/ontology#CommonsMedia", "http://wikiba.se/ontology#ExternalId", "http://wikiba.se/ontology#Url":
		return nil, nil
	case "http://wikiba.se/ontology#WikibaseItem":
		subject, err := Ω.Entity("value")
		if err != nil || subject == nil {
			return nil, err
		}
		return &parsedValue{
			entityValue: subject,
			predicate:   newPredicate(predicateEdge, label),
			schemaType:  schemaTypeUID,
		}, nil

	case "http://wikiba.se/ontology#GlobeCoordinate":
		lat, lng, err := Ω.binding.MustCoordinates("value")
		if err != nil {
			return nil, err
		}
		gj := fmt.Sprintf(`{"type":"feature","geometry":{"type": "Point","coordinates":[%f,%f]}}`, lng, lat)
		return &parsedValue{
			stringValue: escapeFeatureValue(gj),
			predicate:   newPredicate(predicateFeature, label),
			schemaType:  schemaTypeGeo,
		}, nil
	case "http://wikiba.se/ontology#Time":
		dateFields := strings.Split(stringVal, "-")
		if len(dateFields) < 2 {
			return nil, nil
		}
		year := dateFields[0]
		if year == "" {
			// Means it was a negative year.
			year = "-" + dateFields[1]
		}
		return &parsedValue{
			stringValue: year,
			predicate:   newPredicate(predicateFeature, label),
			schemaType:  schemaTypeInt,
		}, nil
	default:
		// "http://wikiba.se/ontology#String",
		// "http://wikiba.se/ontology#Quantity",
		// "http://wikiba.se/ontology#Monolingualtext"
		return &parsedValue{
			stringValue: escapeFeatureValue(stringVal),
			predicate:   newPredicate(predicateFeature, label),
			schemaType:  schemaTypeInt,
		}, nil
	}
}
