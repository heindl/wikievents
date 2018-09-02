package wikievents

import "github.com/heindl/wikievents/sparql"

// Participant is a materialized representation of an event participant, usually a group or state.
type Participant struct {
	URI           string `json:"uri" csv:"uri"`
	Label         string `json:"label" csv:"label"`
	InstanceURI   string `json:"instanceURI" csv:"instanceURI"`
	InstanceLabel string `json:"instanceLabel" csv:"instanceLabel"`
}

func (e *Participant) fromBinding(b sparql.Binding) (err error) {
	e.URI, err = b.MustString("participantURI")
	if err != nil {
		return err
	}
	return nil
}

type Participants []*Participant
