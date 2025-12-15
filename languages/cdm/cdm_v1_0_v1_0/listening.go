package cdm_v1_0_v1_0

import (
	"encoding/json"

	"github.com/erikproper/big-modelling-bus.go.v1/connect"
	"github.com/erikproper/big-modelling-bus.go.v1/generics"
)

type (
	TCDMModelListener struct {
		modelListener connect.TModellingBusArtefactConnector

		CurrentModel    TCDMModel
		UpdatedModel    TCDMModel
		ConsideredModel TCDMModel
	}
)

func NCreateCDMListener(ModellingBusConnector connect.TModellingBusConnector, reporter *generics.TReporter) TCDMModelListener {
	// Setting up a new CDM model listener
	cdmModelListener := TCDMModelListener{}
	cdmModelListener.modelListener = connect.CreateModellingBusArtefactConnector(ModellingBusConnector, ModelJSONVersion, modelID)
	cdmModelListener.CurrentModel = CreateCDMModel(reporter)
	cdmModelListener.UpdatedModel = CreateCDMModel(reporter)
	cdmModelListener.ConsideredModel = CreateCDMModel(reporter)

	// Return the created CDM model listener
	return cdmModelListener
}

// Updating the model's state from given JSON
func (m *TCDMModel) UpdateModelFromJSON(modelJSON json.RawMessage) bool {
	m.Clean()

	return m.reporter.MaybeReportError("Unmarshalling state content failed.", json.Unmarshal(modelJSON, m))
}

// Listening for model state postings on the modelling bus
func (m *TCDMModel) ListenForModelStatePostings(agentId, modelID string, handler func()) {
	m.ModelListener.ListenForJSONArtefactStatePostings(agentId, modelID, handler)
}

// Listening for model update postings on the modelling bus
func (m *TCDMModel) ListenForModelUpdatePostings(agentId, modelID string, handler func()) {
	m.ModelListener.ListenForJSONArtefactUpdatePostings(agentId, modelID, handler)
}

// Listening for model update postings on the modelling bus
func (m *TCDMModel) ListenForModelConsideringPostings(agentId, modelID string, handler func()) {
	m.ModelListener.ListenForJSONArtefactConsideringPostings(agentId, modelID, handler)
}

// Retrieving the model's state from the modelling bus
func (m *TCDMModel) GetStateFromBus(artefactBus connect.TModellingBusArtefactConnector) bool {
	return m.UpdateModelFromJSON(artefactBus.CurrentContent)
}

// Retrieving the model's updated state from the modelling bus
func (m *TCDMModel) GetUpdatedFromBus(artefactBus connect.TModellingBusArtefactConnector) bool {
	return m.UpdateModelFromJSON(artefactBus.UpdatedContent)
}

// Retrieving the model's considered state from the modelling bus
func (m *TCDMModel) GetConsideredFromBus(artefactBus connect.TModellingBusArtefactConnector) bool {
	return m.UpdateModelFromJSON(artefactBus.ConsideredContent)
}
