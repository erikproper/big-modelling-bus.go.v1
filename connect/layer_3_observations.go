// /*
//  *
//  * Package: connect
//  * Layer:   3
//  * Module:  raw_observation
//  *
//  * ..... ... .. .
//  *
//  * Creator: Henderik A. Proper (e.proper@acm.org), TU Wien, Austria
//  *
//  * Version of: XX.11.2025
//  *
//  */
// 
// package connect
// 
// import (
// 	"github.com/erikproper/big-modelling-bus.go.v1/generics"
// )
// 
// const (
// 	rawObservationFilePathElement = "observation/raw"
// )
// 
// /*
//  *
//  * Externally visible functionality
//  *
//  */
// 
// func (b *TModellingBusConnector) PostRawArtefact(topicPath, localFilePath string) {
// 	b.postFile(topicPath, localFilePath)
// }
// 
// func (b *TModellingBusConnector) ListenForRawArtefactPostings(agentID, topicPath string, postingHandler func(string)) {
// 	b.listenForFilePostings(agentID, topicPath, generics.JSONFileName, func(localFilePath, _ string) {
// 		postingHandler(localFilePath)
// 	})
// }
// 
// func (b *TModellingBusConnector) GetRawArtefact(agentID, topicPath, localFileName string) string {
// 	localFilePath, _ := b.getFileFromPosting(agentID, topicPath, localFileName)
// 	return localFilePath
// }
// 
// func (b *TModellingBusConnector) DeleteRawArtefact(topicPath string) {
// 	b.deletePosting(topicPath)
// }
