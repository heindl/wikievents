package wikievents

import (
	"github.com/go-errors/errors"
	"github.com/go-playground/validator"
	"github.com/heindl/wikievents/sparql"
	"strconv"
)

type RequestParams struct {
	StartAtYear int `validate:"required,min=-13798000001,max=2020"` // Earliest date found in Wikidata
	EndAtYear   int `validate:"required,min=-13798000001,max=2020"` // Earliest date found in Wikidata
}

var validate = validator.New()

func FetchEvents(params RequestParams) (*EventResponse, error) {

	if err := validate.Struct(params); err != nil {
		return nil, err
	}

	q, err := sparql.FetchWikidataEvents(params.StartAtYear, params.EndAtYear)
	if err != nil {
		return nil, err
	}

	e := &EventResponse{}
	if err := e.FromBindings(q.Results.Bindings); err != nil {
		return nil, err
	}

	return e, nil
}

func CountEvents(params RequestParams) (int, error) {

	validate := validator.New()
	if err := validate.Struct(params); err != nil {
		return 0, err
	}

	jsonResponse, err := sparql.FetchWikidataEvents(params.StartAtYear, params.EndAtYear)
	if err != nil {
		return 0, err
	}
	if len(jsonResponse.Results.Bindings) == 0 {
		return 0, errors.New("Zero bindings returned")
	}
	binding, found := jsonResponse.Results.Bindings[0]["count"]
	if !found {
		return 0, errors.New("Count field missing in binding")
	}
	count, err := strconv.Atoi(binding.Value)
	if err != nil {
		return 0, errors.Wrap(err, 0)
	}
	return count, nil
}
