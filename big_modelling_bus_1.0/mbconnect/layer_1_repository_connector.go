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
	"github.com/secsy/goftp"
	"os"
	"path/filepath"
	"strings"
	"fmt"
)

type (
	tModellingBusRepositoryConnector struct {
		ftpPort,
		ftpUser,
		ftpAgentRoot,
		ftpServer,
		ftpPassword,
		ftpLocalWorkDirectory string
		ftpSingleServerMode bool

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
	config.User = r.ftpUser
	config.Password = r.ftpPassword

	ftpServerDefinition := r.ftpServer + ":" + r.ftpPort
	client, err := goftp.DialConfig(config, ftpServerDefinition)
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
	remoteDirectoryPath := r.ftpAgentRoot + "/" + topicPath
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

	if ! r.ftpSingleServerMode {
		repositoryEvent.Server = r.ftpServer
		repositoryEvent.Port = r.ftpPort
	}
	repositoryEvent.FilePath = r.ftpAgentRoot + "/" + topicPath + "/" + fileName
	repositoryEvent.FileExtension = fileExtension

	return repositoryEvent
}

func (r *tModellingBusRepositoryConnector) deleteFile(topicPath, fileName, fileExtension string) {
	// Connect to the FTP server
	client, err := r.ftpConnect()
	if err != nil {
		r.reporter.Error("Couldn't open an FTP connection. %s", err)
		return
	}

	err = client.Delete(r.ftpAgentRoot + "/" + topicPath + "/" + fileName + fileExtension)
	if err != nil {
		r.reporter.Error("Couldn't delete File. %s", err)
		return
	}
}

func (r *tModellingBusRepositoryConnector) addJSONAsFile(topicPath string, json []byte) tRepositoryEvent {
	// Define the temporary local file path
	localFilePath := r.ftpLocalWorkDirectory + "/" + GetTimestamp() + jsonFileExtension

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
	localFileName := r.ftpLocalWorkDirectory + "/" + timestamp + repositoryEvent.FileExtension

	config := goftp.Config{}
	serverConnection := ""

	if r.ftpSingleServerMode {
		fmt.Println("SSM")
		serverConnection = r.ftpServer + ":" + r.ftpPort

		config.User = r.ftpUser
		config.Password = r.ftpPassword
	} else {
		serverConnection = repositoryEvent.Server + ":" + repositoryEvent.Port
		fmt.Println("NOTSSM")
	}
			fmt.Println(serverConnection)

	
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
	r.ftpLocalWorkDirectory = configData.GetValue("", "work").String()
	r.ftpPort = configData.GetValue("ftp", "port").String()
	r.ftpUser = configData.GetValue("ftp", "user").String()
	r.ftpServer = configData.GetValue("ftp", "server").String()
	r.ftpPassword = configData.GetValue("ftp", "password").String()
	r.ftpSingleServerMode = configData.GetValue("ftp", "single_server_mode").BoolWithDefault(false)
	r.ftpAgentRoot = configData.GetValue("ftp", "prefix").String() + "/" + topicBase + "/" + agentID

	r.createdPaths = map[string]bool{}

	return &r
}
