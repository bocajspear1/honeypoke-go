package recorder

import (
	"fmt"
	"os"
	"time"
)

// HoneypokeRecord represents a record of input
type HoneypokeRecord struct {
	Time       string
	RemoteIP   string
	RemotePort uint16
	Protocol   string
	Port       uint16
	Input      string
	IsBinary   bool
	UseSSL     bool
	Location   map[string]string
	Host       string
}

type HoneypokeRecorder interface {
	record(record HoneypokeRecord) error
	config() error
}

func NewRecord(remote_ip string, remote_port uint16) *HoneypokeRecord {
	rec := new(HoneypokeRecord)
	rec.Time = time.Now().UTC().Format("2006-01-02T15:04:05-0700")
	rec.Host, _ = os.Hostname()
	return rec
}

func recorderConsumer(c chan *HoneypokeRecord) {
	for {
		record := <-c
		fmt.Printf("Got a record: %s\nData: \n%s\n\n", record.Host, record.Input)
	}
}

func StartRecorders(c chan *HoneypokeRecord) {
	go recorderConsumer(c)
}
