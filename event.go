package wikievents

import (
	"github.com/go-errors/errors"
	"github.com/gocarina/gocsv"
	"github.com/heindl/wikievents/sparql"
	"io"
	"os"
)

type EventQuery struct {
	StartAtYear int `validate:"required,min=-13798000001,max=2020"` // Earliest date found in Wikidata
	EndAtYear   int `validate:"required,min=-13798000001,max=2020"` // Earliest date found in Wikidata
}

// Event is a materialized representation of an event.
type Event struct {
	URI       string  `json:"uri" csv:"uri"`
	Label     string  `json:"label" csv:"label"`
	TypeURI   string  `json:"typeURI" csv:"typeURI"`
	Date      string  `json:"date" csv:"date"`
	Latitude  float64 `json:"latitude" csv:"latitude"`
	Longitude float64 `json:"longitude" csv:"longitude"`
}

func (e *Event) fromBinding(b sparql.Binding) (err error) {
	e.URI, err = b.MustString("eventURI")
	if err != nil {
		return err
	}
	e.Latitude, e.Longitude = b.Coordinates("coordinates")
	e.Date = b.Date("date")
	e.Label = b.String("eventLabel")
	e.TypeURI = b.String("typeURI")
	return nil
}

func safeClose(c io.Closer, err *error) {
	if closeErr := c.Close(); closeErr != nil && *err == nil {
		*err = closeErr
	}
}

func writeCSV(filepath string, Ω interface{}) error {
	f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer safeClose(f, &err)

	if err := gocsv.MarshalWithoutHeaders(Ω, f); err != nil {
		return errors.Wrap(err, 0)
	}

	return nil
}

type Link struct {
	EventURI       string `json:"eventUri" csv:"eventUri"`
	ParticipantURI string `json:"participantUri" csv:"participantUri"`
}
