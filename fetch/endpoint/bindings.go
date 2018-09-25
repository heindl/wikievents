// Copyright (c) 2018 Parker Heindl. All rights reserved.
//
// Use of this source code is governed by the MIT License.
// Read LICENSE.md in the project root for information.

package endpoint

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Binding map[string]struct {
	DataType string `json:"datatype"`
	Type     string `json:"type"`
	Value    string `json:"value"`
	Lang     string `json:"xml:lang"`
}

func (b Binding) ensureKey(key string) error {
	if _, ok := b[key]; !ok {
		return errors.Errorf("key [%s] not found in binding", key)
	}
	if len(b[key].Value) == 0 {
		return errors.Errorf("key [%s] not found in binding", key)
	}
	return nil
}

func (b Binding) Coordinates(key string) (float64, float64) {
	lat, lng, err := b.MustCoordinates(key)
	if err != nil {
		return 0, 0
	}
	return lat, lng
}

func (b Binding) Values() map[string]string {
	res := map[string]string{}
	for k, v := range b {
		res[k] = v.Value
	}
	return res
}

func (b Binding) MustCoordinates(key string) (float64, float64, error) {
	if err := b.ensureKey(key); err != nil {
		return 0, 0, err
	}
	wkt := regexp.MustCompile(`\(([^\)]+)\)`).FindString(b[key].Value)
	wkt = strings.Trim(wkt, "()")
	coords := strings.Split(wkt, " ")

	if len(coords) == 0 {
		return 0, 0, errors.Errorf("Coordinate field [%s] does not match expected pattern [Point(int, int)]", key)
	}

	lng, err := strconv.ParseFloat(coords[0], 64)
	if err != nil {
		return 0, 0, errors.Wrap(err, "could not parse coordinate float")
	}

	lat, err := strconv.ParseFloat(coords[1], 64)
	if err != nil {
		return 0, 0, errors.Wrap(err, "could not parse coordinate float")
	}

	return lat, lng, nil
}

func (b Binding) Date(key string) string {
	// TODO: Golang time module does not support negative dates. Consider changing in canonical library.
	if err := b.ensureKey(key); err != nil {
		return ""
	}
	return b[key].Value
}

func (b Binding) MustDate(key string) (string, error) {
	if err := b.ensureKey(key); err != nil {
		return "", err
	}
	return b[key].Value, nil
}

func (b Binding) Interface(key string) interface{} {
	if err := b.ensureKey(key); err != nil {
		return nil
	}
	return b[key].Value
}

func (b Binding) String(key string) string {
	if err := b.ensureKey(key); err != nil {
		return ""
	}
	return b[key].Value
}

func (b Binding) MustString(key string) (string, error) {
	if err := b.ensureKey(key); err != nil {
		return "", err
	}
	return b[key].Value, nil
}

func (b Binding) Type(key string) string {
	if err := b.ensureKey(key); err != nil {
		return ""
	}
	return b[key].Type
}

func (b Binding) MustInt(key string) (int, error) {
	if err := b.ensureKey(key); err != nil {
		return 0, err
	}
	i, err := strconv.Atoi(b[key].Value)
	if err != nil {
		return 0, errors.Wrap(err, "could not parse integer")
	}
	return i, nil
}

func (b Binding) Int(key string) int {
	if err := b.ensureKey(key); err != nil {
		return 0
	}
	i, err := strconv.Atoi(b[key].Value)
	if err != nil {
		return 0
	}
	return i
}
