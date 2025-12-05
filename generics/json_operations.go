/*
 *
 * Module:    BIG Modelling Bus
 * Package:   Generic
 * Component: JSON Operations
 *
 * ..... ... .. .
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), TU Wien, Austria
 *
 * Version of: XX.11.2025
 *
 */

package generics

import (
	"encoding/json"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/wI2L/jsondiff"
)

func jsonDiff(sourceJSON, targetJSON []byte) (json.RawMessage, error) {
	deltaOperations, err := jsondiff.CompareJSON(sourceJSON, targetJSON)
	if err != nil {
		return nil, err
	}

	return json.Marshal(deltaOperations)
}

func jsonApplyPatch(sourceJSON, patchJSON []byte) (json.RawMessage, error) {
	patch, err := jsonpatch.DecodePatch(patchJSON)
	if err != nil {
		return nil, err
	}

	return patch.Apply(sourceJSON)
}
