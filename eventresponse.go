package wikievents

import (
	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/heindl/wikievents/sparql"
	"os"
	"path"
)

type EventResponse struct {
	Events       []*Event     `json:"events"`
	Links        []*Link      `json:"links"`
	Participants Participants `json:"participants"`
}

func (r *EventResponse) WriteToTemp() (outputPath string, err error) {
	dirPath := path.Join(os.TempDir(), uuid.Must(uuid.NewRandom()).String())
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return "", errors.Wrap(err, 0)
	}

	if err := writeCSV(path.Join(dirPath, "events.csv"), r.Events); err != nil {
		return "", err
	}

	if err := writeCSV(path.Join(dirPath, "participants.csv"), r.Participants); err != nil {
		return "", err
	}

	if err := writeCSV(path.Join(dirPath, "links.csv"), r.Links); err != nil {
		return "", err
	}
	return dirPath, nil
}

func (r *EventResponse) FromBindings(bindings []sparql.Binding) error {

	if r == nil {
		return errors.New("Nil event response")
	}

	eventSet := map[string]*Event{}
	participantURIs := []string{}

	for _, b := range bindings {
		event := &Event{}
		if err := event.fromBinding(b); err != nil {
			return err
		}
		eventSet[event.URI] = event
		participantURI, err := b.MustString("participantURI")
		if err != nil {
			return err
		}
		participantURIs = append(participantURIs, participantURI)
		r.Links = append(r.Links, &Link{
			ParticipantURI: participantURI,
			EventURI:       event.URI,
		})
	}

	for _, v := range eventSet {
		r.Events = append(r.Events, v)
	}
	return nil
}
