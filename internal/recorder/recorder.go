/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package recorder

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/oschwald/geoip2-golang"
)

// HoneypokeRecord represents a record of input
type HoneypokeRecord struct {
	Time       string             `json:"time"`
	RemoteIP   string             `json:"remote_ip"`
	RemotePort int                `json:"remote_port"`
	Protocol   string             `json:"protocol"`
	Port       int                `json:"port"`
	Input      string             `json:"input"`
	IsBinary   bool               `json:"is_binary"`
	UseSSL     bool               `json:"use_ssl"`
	Location   map[string]float64 `json:"location"`
	Host       string             `json:"host"`
}

type HoneypokeRecorder interface {
	Record(record *HoneypokeRecord) error
}

func NewRecord(remote_ip string, remote_port uint16) *HoneypokeRecord {
	rec := new(HoneypokeRecord)
	rec.Time = time.Now().UTC().Format("2006-01-02T15:04:05-0700")
	rec.Host, _ = os.Hostname()
	return rec
}

func recorderConsumer(recorders []HoneypokeRecorder, c chan *HoneypokeRecord) {

	db, err := geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	for {
		record := <-c
		if record.RemoteIP != "" {
			ip := net.ParseIP(record.RemoteIP)
			cityData, err := db.City(ip)

			record.Location = make(map[string]float64)
			record.Location["lat"] = 0.0
			record.Location["lon"] = 0.0

			if err == nil {
				lat, err := strconv.ParseFloat(fmt.Sprintf("%.2f", cityData.Location.Latitude), 64)
				if err != nil {
					lat = 0.0
				}
				record.Location["lat"] = lat
				lon, err := strconv.ParseFloat(fmt.Sprintf("%.2f", cityData.Location.Longitude), 64)
				if err != nil {
					lon = 0.0
				}
				record.Location["lon"] = lon
			}

		}
		for _, recorder := range recorders {
			recorder.Record(record)
		}
		// fmt.Printf("Got a record: %s\nData: \n%s\n\n", record.Host, record.Input)
	}
}

func StartRecorders(recorders []HoneypokeRecorder, c chan *HoneypokeRecord) {
	go recorderConsumer(recorders, c)
}
