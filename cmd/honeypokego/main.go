package main

import (
	"log"
	"time"

	"github.com/bocajspear1/honeypoke-go/internal/server"
	"github.com/google/gopacket/layers"

	"github.com/bocajspear1/honeypoke-go/internal/permissions"
	"github.com/bocajspear1/honeypoke-go/internal/recorder"
	"github.com/bocajspear1/honeypoke-go/internal/watcher"
)

type config struct {
}

func waitForSetup(contChan chan bool, serverCount int) {
	for i := 0; i < serverCount+1; i++ {
		_ = <-contChan
	}
	log.Printf("%d servers and the watcher have reported they are running\n", serverCount)
	permissions.DropPermissions("nobody", "nogroup")
}

func main() {
	// Make our communication channels
	recordChan := make(chan *recorder.HoneypokeRecord)
	contChan := make(chan bool)

	// Start the recorders routine
	recorder.StartRecorders(recordChan)

	// Start the servers
	server.StartServer(layers.LayerTypeTCP, 80, false, recordChan, contChan)

	// Start the missed port watching routine
	watcher.StartWatcher("eth0", "", contChan)

	waitForSetup(contChan, 1)

	for {
		time.Sleep(time.Second * 10)
	}
}
