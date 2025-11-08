/*
 *
 * Package: mbconnect
 * Layer:   1
 * Module:  config_reader
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
	"gopkg.in/ini.v1"
)

func (b *TModellingBusConnector) readConfig(config string) {
	cfg, err := ini.Load(config)
	if err != nil {
		b.errorReporter("Failed to read config file:", err)
		return
	}

	cfgGeneralSection := cfg.Section("")
	b.experimentID = cfgGeneralSection.Key("experiment").String()
	b.AgentID = cfgGeneralSection.Key("agent").String()
	b.ftpLocalWorkFolder = cfgGeneralSection.Key("work").String()

	cfgFTPSection := cfg.Section("ftp")
	b.ftpPort = cfgFTPSection.Key("port").String()
	b.ftpUser = cfgFTPSection.Key("user").String()
	b.ftpServer = cfgFTPSection.Key("server").String()
	b.ftpPassword = cfgFTPSection.Key("password").String()
	b.ftpPathPrefix = cfgFTPSection.Key("prefix").String()

	cfgMQTTSection := cfg.Section("mqtt")
	b.mqttPort = cfgMQTTSection.Key("port").String()
	b.mqttUser = cfgMQTTSection.Key("user").String()
	b.mqttBroker = cfgMQTTSection.Key("broker").String()
	b.mqttPassword = cfgMQTTSection.Key("password").String()
	b.mqttPathPrefix = cfgMQTTSection.Key("prefix").String()
}