/*
 *
 * Package: mbconnect
 * Layer:   3
 * Module:  raw_artefacts
 *
 * ..... ... .. .
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), TU Wien, Austria
 *
 * Version of: XX.11.2025
 *
 */

package connect

import (
	"github.com/erikproper/big-modelling-bus.go.v1/generics"
)

const (
	rawArtefactsFilePathElement = "artefacts/file"
)

func (b *TModellingBusConnector) rawArtefactPath(context, format, fileName string) string {
	return rawArtefactsFilePathElement +
		"/" + context +
		"/" + format +
		"/" + fileName
}

/*
 *
 * Externally visible functionality
 *
 */

func (b *TModellingBusConnector) PostRawArtefact(context, format, fileName, localFilePath string) {
	b.postFile(b.rawArtefactPath(context, format, fileName), localFilePath)
}

func (b *TModellingBusConnector) ListenForRawArtefactPostings(agentID, context, format, fileName string, postingHandler func(string)) {
	topicPath := b.rawArtefactPath(context, format, fileName)

	b.listenForFilePostings(agentID, topicPath, generics.JSONFileName, func(localFilePath, _ string) {
		postingHandler(localFilePath)
	})
}

func (b *TModellingBusConnector) GetRawArtefact(agentID, context, format, fileName, localFileName string) string {
	topicPath := b.rawArtefactPath(context, format, fileName)

	localFilePath, _ := b.getFileFromPosting(agentID, topicPath, localFileName)
	return localFilePath
}

func (b *TModellingBusConnector) DeleteRawArtefact(context, format, fileName string) {
	b.deletePosting(b.rawArtefactPath(context, format, fileName))
}
