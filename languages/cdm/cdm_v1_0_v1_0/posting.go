/*
 *
 * Module:    BIG Modelling Bus
 * Package:   Languages/Conceptual Domain Modelling, Version 1
 * Component: Definition
 *
 * This package implements the Conceptual Domain Modelling language, version 1, for the BIG Modelling Bus.
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), TU Wien, Austria
 *
 * Version of: XX.11.2025
 *
 */

package cdm_v1_0_v1_0

import (
	"encoding/json"

	"github.com/erikproper/big-modelling-bus.go.v1/connect"
)

type (
	TCDMModelPoster struct {
		modelPoster connect.TModellingBusArtefactConnector // ???
	}
)

// Posting the model's state
func (p *TCDMModelPoster) PostState(m TCDMModel) {
	p.modelPoster.PostJSONArtefactState(json.Marshal(m))
}

// Posting the model's update
func (p *TCDMModelPoster) PostUpdate(m TCDMModel) {
	p.modelPoster.PostJSONArtefactUpdate(json.Marshal(m))
}

// Posting the model's considered update
func (p *TCDMModelPoster) PostConsidering(m TCDMModel) {
	p.modelPoster.PostJSONArtefactConsidering(json.Marshal(m))
}

// Creating a CDM model poster, which uses a given ModellingBusConnector to post the model
func CreateCDMPoster(ModellingBusConnector connect.TModellingBusConnector, modelID string) TCDMModelPoster {
	// Setting up new CDM model poster
	cdmPosterModel := TCDMModelPoster{}
	cdmPosterModel.modelPoster = connect.CreateModellingBusArtefactConnector(ModellingBusConnector, ModelJSONVersion, modelID)

	// Return the created CDM model poster
	return cdmPosterModel
}
