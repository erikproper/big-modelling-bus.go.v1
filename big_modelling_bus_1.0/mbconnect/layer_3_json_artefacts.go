/*
 *
 * Package: mbconnect
 * Layer:   3
 * Module:  json_artefacts
 *
 * ..... ... .. .
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), TU Wien, Austria
 *
 * Version of: XX.11.2025
 *
 */

package mbconnect

import (
	"encoding/json"
)

const (
	artefactsPathElement           = "artefacts/file"
	artefactStatePathElement       = "state"
	artefactConsideringPathElement = "considering"
	artefactUpdatePathElement      = "update"
)

type (
	TModellingBusArtefactConnector struct {
		ModellingBusConnector TModellingBusConnector
		Timestamp             string `json:"timestamp"`
		JSONVersion           string `json:"json version"`
		ArtefactID            string `json:"artefact id"`

		// Externally visible
		ArtefactCurrentContent    json.RawMessage `json:"content"`
		ArtefactUpdatedContent    json.RawMessage `json:"-"`
		ArtefactConsideredContent json.RawMessage `json:"-"`

		// Before we can communicate updates or considering postings, we must have
		// communicated the state of the model first
		stateCommunicated bool `json:"-"`
	}
)

/*
 *
 * Internal functionality
 *
 */

func (b *TModellingBusArtefactConnector) artefactsTopicPath(artefactID string) string {
	return artefactsPathElement +
		"/" + artefactID +
		"/" + b.JSONVersion
}

func (b *TModellingBusArtefactConnector) artefactsStateTopicPath(artefactID string) string {
	return b.artefactsTopicPath(artefactID) +
		"/" + artefactStatePathElement
}

func (b *TModellingBusArtefactConnector) artefactsUpdateTopicPath(artefactID string) string {
	return b.artefactsTopicPath(artefactID) +
		"/" + artefactUpdatePathElement
}

func (b *TModellingBusArtefactConnector) artefactsConsideringTopicPath(artefactID string) string {
	return b.artefactsTopicPath(artefactID) +
		"/" + artefactConsideringPathElement
}

type TJSONDelta struct {
	Operations     json.RawMessage `json:"operations"`
	Timestamp      string          `json:"timestamp"`
	StateTimestamp string          `json:"state timestamp"`
}

func (b *TModellingBusArtefactConnector) postDelta(deltaTopicPath string, oldStateJSON, newStateJSON []byte, err error) {
	if err != nil {
		b.ModellingBusConnector.reporter.Error("Something went wrong when converting to JSON. %s", err)
		return
	}

	deltaOperationsJSON, err := jsonDiff(oldStateJSON, newStateJSON)
	if err != nil {
		b.ModellingBusConnector.reporter.Error("Something went wrong running the JSON diff. %s", err)
		return
	}

	delta := TJSONDelta{}
	delta.Timestamp = GetTimestamp()
	delta.StateTimestamp = b.Timestamp
	delta.Operations = deltaOperationsJSON

	deltaJSON, err := json.Marshal(delta)
	if err != nil {
		b.ModellingBusConnector.reporter.Error("Something went wrong JSONing the diff patch. %s", err)
		return
	}

	b.ModellingBusConnector.postJSON(deltaTopicPath, b.JSONVersion, deltaJSON, delta.Timestamp)
}

func (b *TModellingBusArtefactConnector) applyDelta(currentJSONState json.RawMessage, deltaJSON []byte) (json.RawMessage, bool) {
	delta := TJSONDelta{}
	err := json.Unmarshal(deltaJSON, &delta)
	if err != nil {
		b.ModellingBusConnector.reporter.Error("Something went wrong unJSONing the received diff patch. %s", err)
		return currentJSONState, false
	}

	if delta.StateTimestamp != b.Timestamp {
		b.ModellingBusConnector.reporter.Error("Received JSON delta out of order.")
		return currentJSONState, false
	}

	newJSONState, err := jsonApplyPatch(currentJSONState, delta.Operations)
	if err != nil {
		b.ModellingBusConnector.reporter.Error("Applying patch didn't work. %s", err)
		return currentJSONState, false
	}

	return newJSONState, true
}

/*
 *
 * Externally visible functionality
 *
 */

/*
 * Posting
 */

func (b *TModellingBusArtefactConnector) PrepareForPosting(ArtefactID string) {
	b.ArtefactID = ArtefactID
}

func (b *TModellingBusArtefactConnector) PostConsidering(consideringStateJSON []byte, err error) {
	if b.stateCommunicated {
		b.ArtefactConsideredContent = consideringStateJSON

		b.postDelta(b.artefactsUpdateTopicPath(b.ArtefactID), b.ArtefactCurrentContent, b.ArtefactUpdatedContent, err)
		b.postDelta(b.artefactsConsideringTopicPath(b.ArtefactID), b.ArtefactUpdatedContent, b.ArtefactConsideredContent, err)
	} else {
		b.ModellingBusConnector.reporter.Error("We must always see a state posting, before a considering posting!")
	}
}

func (b *TModellingBusArtefactConnector) PostUpdate(updatedStateJSON []byte, err error) {
	if b.stateCommunicated {
		b.ArtefactUpdatedContent = updatedStateJSON

		b.postDelta(b.artefactsUpdateTopicPath(b.ArtefactID), b.ArtefactCurrentContent, b.ArtefactUpdatedContent, err)
	} else {
		b.PostState(updatedStateJSON, err)
	}
}

func (b *TModellingBusArtefactConnector) PostState(stateJSON []byte, err error) {
	if err != nil {
		b.ModellingBusConnector.reporter.Error("Something went wrong when converting to JSON. %s", err)
		return
	}

	b.Timestamp = GetTimestamp()
	b.ArtefactCurrentContent = stateJSON

	if err != nil {
		b.ModellingBusConnector.reporter.Error("Something went wrong JSONing the artefact. %s", err)
		return
	}

	b.ModellingBusConnector.postJSON(b.artefactsStateTopicPath(b.ArtefactID), b.JSONVersion, stateJSON, b.Timestamp)
	b.stateCommunicated = true
}

/*
 * Listening
 */

func (b *TModellingBusArtefactConnector) ListenForStatePostings(agentID, ArtefactID string, handler func()) {
	b.ModellingBusConnector.listenForJSONPostings(agentID, b.artefactsStateTopicPath(ArtefactID), func(json []byte, timestamp string) {
		b.ArtefactCurrentContent = json
		b.ArtefactUpdatedContent = json
		b.ArtefactConsideredContent = json
		b.Timestamp = timestamp

		handler()
	})
}

func (b *TModellingBusArtefactConnector) ListenForUpdatePostings(agentID, ArtefactID string, handler func()) {
	b.ModellingBusConnector.listenForJSONPostings(agentID, b.artefactsUpdateTopicPath(ArtefactID), func(json []byte, timestamp string) {
		ok := false
		b.ArtefactUpdatedContent, ok = b.applyDelta(b.ArtefactCurrentContent, json)
		if ok {
			b.ArtefactConsideredContent = b.ArtefactUpdatedContent

			handler()
		}
	})
}

func (b *TModellingBusArtefactConnector) ListenForConsideringPostings(agentID, ArtefactID string, handler func()) {
	b.ModellingBusConnector.listenForJSONPostings(agentID, b.artefactsConsideringTopicPath(ArtefactID), func(json []byte, timestamp string) {
		ok := false
		b.ArtefactConsideredContent, ok = b.applyDelta(b.ArtefactUpdatedContent, json)
		if ok {
			handler()
		}
	})
}

/*
 * Creation
 */

func CreateModellingBusArtefactConnector(ModellingBusConnector TModellingBusConnector, JSONVersion string) TModellingBusArtefactConnector {
	ModellingBusArtefactConnector := TModellingBusArtefactConnector{}
	ModellingBusArtefactConnector.ModellingBusConnector = ModellingBusConnector
	ModellingBusArtefactConnector.JSONVersion = JSONVersion
	ModellingBusArtefactConnector.ArtefactCurrentContent = []byte{}
	ModellingBusArtefactConnector.ArtefactUpdatedContent = []byte{}
	ModellingBusArtefactConnector.ArtefactConsideredContent = []byte{}
	ModellingBusArtefactConnector.Timestamp = GetTimestamp()
	ModellingBusArtefactConnector.stateCommunicated = false

	return ModellingBusArtefactConnector
}
