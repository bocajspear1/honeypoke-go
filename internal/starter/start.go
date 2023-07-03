/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package starter

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"github.com/bocajspear1/honeypoke-go/internal/server"
	"github.com/google/gopacket/layers"

	"github.com/bocajspear1/honeypoke-go/internal/permissions"
	"github.com/bocajspear1/honeypoke-go/internal/recorder"
	"github.com/bocajspear1/honeypoke-go/internal/watcher"
)

const configPath string = "config.json"

type recorderConfig struct {
	Enabled        bool                   `json:"enabled"`
	RecorderName   string                 `json:"name"`
	RecorderConfig map[string]interface{} `json:"config"`
}

type tcpConfig struct {
	Port uint16 `json:"port"`
	SSL  bool   `json:"ssl"`
}

type honeyPokeConfig struct {
	Recorders      []recorderConfig `json:"recorders"`
	UDPPorts       []int            `json:"udp_ports"`
	TCPPorts       []tcpConfig      `json:"tcp_ports"`
	IgnoreTCPPorts []uint16         `json:"ignore_tcp_ports"`
	NewUser        string           `json:"user"`
	NewGroup       string           `json:"group"`
	Interface      string           `json:"interface"`
}

func waitForSetup(newUser string, newGroup string, contChan chan bool, serverCount int) {
	for i := 0; i < serverCount+1; i++ {
		_ = <-contChan
	}
	log.Printf("%d servers and the watcher have reported they are running\n", serverCount)
	permissions.DropPermissions(newUser, newGroup)
}

func parseJSON() (*honeyPokeConfig, error) {
	configFile, ferr := ioutil.ReadFile(configPath)
	if ferr != nil {
		return nil, errors.New("Could not parse config file: File not found")
	}

	var config honeyPokeConfig
	jerr := json.Unmarshal(configFile, &(config))
	if jerr != nil {
		return nil, jerr
	}

	return &config, nil
}

// StartHoneyPoke starts HoneyPoke and all the servers and recorders
func StartHoneyPoke() {

	config, cerr := parseJSON()
	if cerr != nil {
		log.Fatalln(cerr)
		return
	}

	// Make our communication channels
	recordChan := make(chan *recorder.HoneypokeRecord)
	contChan := make(chan bool)

	if len(config.Recorders) == 0 {
		log.Fatalln("No recorders in config file")
	}

	recoderList := make([]recorder.HoneypokeRecorder, 0)

	// Start the recorders routine
	for _, recorderData := range config.Recorders {
		if recorderData.RecorderName == "elasticsearch6" && recorderData.Enabled == true {
			recoderList = append(recoderList, recorder.NewElastic6Recorder(recorderData.RecorderConfig))
		} else if recorderData.RecorderName == "elasticsearch7" && recorderData.Enabled == true {
			recoderList = append(recoderList, recorder.NewElastic7Recorder(recorderData.RecorderConfig))
		} else if recorderData.RecorderName == "elasticsearch8" && recorderData.Enabled == true {
			recoderList = append(recoderList, recorder.NewElastic8Recorder(recorderData.RecorderConfig))
		} else {
			log.Printf("Invalid name %s\n", recorderData.RecorderName)
		}
	}

	if len(recoderList) == 0 {
		log.Fatalln("No recorders configured")
	}

	recorder.StartRecorders(recoderList, recordChan)

	serverCount := 0

	pcapFilter := ""

	// Add the TCP ignores
	for _, item := range config.IgnoreTCPPorts {
		if pcapFilter != "" {
			pcapFilter += " and not tcp port " + strconv.Itoa((int)(item))
		} else {
			pcapFilter = "not tcp port " + strconv.Itoa((int)(item))
		}
	}

	// Start the TCP servers
	for _, item := range config.TCPPorts {
		if pcapFilter != "" {
			pcapFilter += " and not tcp port " + strconv.Itoa((int)(item.Port))
		} else {
			pcapFilter = "not tcp port " + strconv.Itoa((int)(item.Port))
		}
		server.StartServer(layers.LayerTypeTCP, (int)(item.Port), item.SSL, recordChan, contChan)
		serverCount++
	}

	// Start the UDP servers
	for _, udpPort := range config.UDPPorts {
		if pcapFilter != "" {
			pcapFilter += " and not udp port " + strconv.Itoa((int)(udpPort))
		} else {
			pcapFilter = "not udp port " + strconv.Itoa((int)(udpPort))
		}
		server.StartServer(layers.LayerTypeUDP, (int)(udpPort), false, recordChan, contChan)
		serverCount++
	}

	// Start the missed port watching routine
	watcher.StartWatcher(config.Interface, pcapFilter, contChan)

	// Wait for everybody to report they are running
	waitForSetup(config.NewUser, config.NewGroup, contChan, serverCount)

	for {
		time.Sleep(time.Second * 10)
	}
}
