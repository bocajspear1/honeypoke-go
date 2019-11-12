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

func waitForSetup(contChan chan bool, serverCount int) {
	for i := 0; i < serverCount+1; i++ {
		_ = <-contChan
	}
	log.Printf("%d servers and the watcher have reported they are running\n", serverCount)
	permissions.DropPermissions("nobody", "nogroup")
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
		log.Fatalln("Not recorders configured")
	}

	recoderList := make([]recorder.HoneypokeRecorder, 0)

	// Start the recorders routine
	for _, recorderData := range config.Recorders {
		if recorderData.RecorderName == "elasticsearch6" {
			recoderList = append(recoderList, recorder.NewElastic6Recorder(recorderData.RecorderConfig))
		}
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

	waitForSetup(contChan, serverCount)

	for {
		time.Sleep(time.Second * 10)
	}
}
