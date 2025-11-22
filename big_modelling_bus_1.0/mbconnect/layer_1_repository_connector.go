/*
 *
 * Package: mbconnect
 * Layer:   1
 * Module:  repository_connector
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
	"fmt"
	"github.com/secsy/goftp"
	"os"
	"path/filepath"
	"strings"
)

type (
	tModellingBusRepositoryConnector struct {
		port,
		user,
		server,
		password,
		agentRoot,
		localWorkDirectory string
		activeTransfers,
		singleServerMode bool

		createdPaths map[string]bool

		reporter *TReporter
	}
)

type tRepositoryEvent struct {
	Server        string `json:"server,omitempty"`
	Port          string `json:"port,omitempty"`
	FilePath      string `json:"file path,omitempty"`
	FileExtension string `json:"file extension,omitempty"`
}

func (r *tModellingBusRepositoryConnector) ftpConnect() (*goftp.Client, error) {
	config := goftp.Config{}
	config.User = r.user
	config.Password = r.password
	config.ActiveTransfers = r.activeTransfers

	serverDefinition := r.server + ":" + r.port
	client, err := goftp.DialConfig(config, serverDefinition)
	if err != nil {
		r.reporter.Error("Error connecting to the FTP server. %s", err)
		return client, err
	}

	return client, err
}

func (r *tModellingBusRepositoryConnector) mkRepositoryDirectoryPath(remoteDirectoryPath string) {
	if !r.createdPaths[remoteDirectoryPath] {
		// Connect to the FTP server
		client, err := r.ftpConnect()
		if err != nil {
			r.reporter.Error("Couldn't open an FTP connection. %s", err)
			return
		}

		pathCovered := ""
		for _, Directory := range strings.Split(remoteDirectoryPath, "/") {
			pathCovered = pathCovered + Directory + "/"
			client.Mkdir(pathCovered)
		}

		client.Close()

		r.createdPaths[remoteDirectoryPath] = true
	}
}

func (r *tModellingBusRepositoryConnector) addFile(topicPath, fileName, fileExtension, localFilePath string) tRepositoryEvent {
	remoteDirectoryPath := r.agentRoot + "/" + topicPath
	remoteFilePath := remoteDirectoryPath + "/" + fileName + fileExtension

	r.mkRepositoryDirectoryPath(remoteDirectoryPath)

	repositoryEvent := tRepositoryEvent{}

	// Connect to the FTP server
	client, err := r.ftpConnect()
	{
		if err != nil {
			r.reporter.Error("Couldn't open an FTP connection. %s", err)
			return repositoryEvent
		}

		file, err := os.Open(filepath.FromSlash(localFilePath))
		if err != nil {
			r.reporter.Error("Error opening File for reading. %s", err)
			return repositoryEvent
		}

		err = client.Store(remoteFilePath, file)
		if err != nil {
			r.reporter.Error("Error uploading File to ftp server. %s", err)
			return repositoryEvent
		}
	}
	client.Close()

	if !r.singleServerMode {
		repositoryEvent.Server = r.server
		repositoryEvent.Port = r.port
	}
	repositoryEvent.FilePath = r.agentRoot + "/" + topicPath + "/" + fileName
	repositoryEvent.FileExtension = fileExtension

	return repositoryEvent
}

func deleteRepositoryPath(client *goftp.Client, deletePath string) {
	fmt.Println("Deleting:", deletePath)
	client.Delete(deletePath)
	fileInfos, _ := client.ReadDir(deletePath)
	if len(fileInfos) > 0 {
		//		if fileInfos.IsDir() {
		fmt.Println("Deleting directory:", deletePath)
		for _, fileInfo := range fileInfos {
			deleteRepositoryPath(client, deletePath+"/"+fileInfo.Name())
		}
		client.Rmdir(deletePath)
	}
}

func (r *tModellingBusRepositoryConnector) deletePath(topicPath string) {
	// Connect to the FTP server
	client, err := r.ftpConnect()
	if err != nil {
		r.reporter.Error("Couldn't open an FTP connection:", err)
		return
	}

	// Delete path from server
	remotePath := r.agentRoot + "/" + topicPath
	deleteRepositoryPath(client, remotePath)
}

func (r *tModellingBusRepositoryConnector) deleteFile(topicPath, fileName, fileExtension string) {
	filePath := topicPath + "/" + fileName + fileExtension

	r.deletePath(filePath)
}

func (r *tModellingBusRepositoryConnector) addJSONAsFile(topicPath string, json []byte) tRepositoryEvent {
	// Define the temporary local file path
	localFilePath := r.localWorkDirectory + "/" + GetTimestamp() + jsonFileExtension

	// Create a temporary local file with the JSON record
	err := os.WriteFile(filepath.FromSlash(localFilePath), json, 0644)
	if err != nil {
		r.reporter.Error("Error writing to temporary file. %s", err)
	}

	// Cleanup the temporary file afterwards
	defer os.Remove(filepath.FromSlash(localFilePath))

	return r.addFile(topicPath, jsonFileName, jsonFileExtension, localFilePath)
}

func (r *tModellingBusRepositoryConnector) getFile(repositoryEvent tRepositoryEvent, timestamp string) string {
	localFileName := r.localWorkDirectory + "/" + timestamp + repositoryEvent.FileExtension

	config := goftp.Config{}
	config.ActiveTransfers = r.activeTransfers
	serverConnection := ""

	if r.singleServerMode {
		serverConnection = r.server + ":" + r.port

		config.User = r.user
		config.Password = r.password
	} else {
		serverConnection = repositoryEvent.Server + ":" + repositoryEvent.Port
	}

	client, err := goftp.DialConfig(config, serverConnection)
	if err != nil {
		r.reporter.Error("Something went wrong connecting to the FTP server", err)
		return ""
	}

	// Download a File to local storage
	File, err := os.Create(localFileName)
	if err != nil {
		r.reporter.Error("Something went wrong creating local file", err)
		return ""
	}

	err = client.Retrieve(repositoryEvent.FilePath+repositoryEvent.FileExtension, File)
	if err != nil {
		r.reporter.Error("Something went wrong retrieving file", err)
		return ""
	}

	return localFileName
}

func createModellingBusRepositoryConnector(topicBase, agentID string, configData *TConfigData, reporter *TReporter) *tModellingBusRepositoryConnector {
	r := tModellingBusRepositoryConnector{}

	r.reporter = reporter

	// Get data from the config file
	r.localWorkDirectory = configData.GetValue("", "work").String()
	r.port = configData.GetValue("ftp", "port").String()
	r.user = configData.GetValue("ftp", "user").String()
	r.server = configData.GetValue("ftp", "server").String()
	r.password = configData.GetValue("ftp", "password").String()
	r.singleServerMode = configData.GetValue("ftp", "single_server_mode").BoolWithDefault(false)
	r.activeTransfers = configData.GetValue("ftp", "active_transfers").BoolWithDefault(false)
	r.agentRoot = configData.GetValue("ftp", "prefix").String() + "/" + topicBase + "/" + agentID

	r.createdPaths = map[string]bool{}

	if r.singleServerMode {
		r.reporter.Progress("Running the FTP connection in single server mode")
	} else {
		r.reporter.Progress("Running the FTP connection in multi server mode")
	}

	if r.activeTransfers {
		r.reporter.Progress("Running the FTP connection in active transfer mode")
	} else {
		r.reporter.Progress("Running the FTP connection in passive transfer mode")
	}

	return &r
}
