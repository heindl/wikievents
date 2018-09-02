package sparql

import (
	"github.com/go-errors/errors"
	"regexp"
	"strconv"
	"strings"
)

type QueryResponse struct {
	Head struct {
		Link []string `json:"link"`
		Vars []string `json:"vars"`
	} `json:"head"`
	Results struct {
		Distinct bool      `json:"distinct"`
		Ordered  bool      `json:"ordered"`
		Bindings []Binding `json:"bindings"`
	}
}

type Binding map[string]struct {
	DataType string `json:"datatype"`
	Type     string `json:"type"`
	Value    string `json:"value"`
	Lang     string `json:"xml:lang"`
}

var ErrKeyNotFound = errors.Errorf("Key not found in binding")

func (b Binding) ensureKey(key string) error {
	if _, ok := b[key]; !ok {
		return errors.WrapPrefix(ErrKeyNotFound, key, 0)
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
		return 0, 0, errors.Wrap(err, 0)
	}

	lat, err := strconv.ParseFloat(coords[1], 64)
	if err != nil {
		return 0, 0, errors.Wrap(err, 0)
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

func (b Binding) MustInt(key string) (int, error) {
	if err := b.ensureKey(key); err != nil {
		return 0, err
	}
	i, err := strconv.Atoi(b[key].Value)
	if err != nil {
		return 0, errors.Wrap(err, 0)
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
