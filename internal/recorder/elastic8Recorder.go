/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package recorder

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	elasticsearch8 "github.com/elastic/go-elasticsearch/v8"
	esapi "github.com/elastic/go-elasticsearch/v8/esapi"
)

type Elastic8Recorder struct {
	client *elasticsearch8.Client
}

func (r Elastic8Recorder) Record(record *HoneypokeRecord) error {

	data, err := json.Marshal(record)

	if err != nil {
		log.Panicln(err)
	}

	req := esapi.IndexRequest{
		Index:   "honeypoke",
		Body:    strings.NewReader((string)(data)),
		Refresh: "true",
	}

	// Perform the request with the client.
	res, err := req.Do(context.Background(), r.client)
	if err != nil {
		log.Printf("Error getting response: %s", err)
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[%s] Error indexing document ID", res.Status())
	}

	return nil
}

func NewElastic8Recorder(config map[string]interface{}) *Elastic8Recorder {

	host, ok := config["host"].(string)
	if !ok {
		log.Fatalln("Could not find 'host' entry for elasticsearch8")
	}
	username, ok := config["username"].(string)
	if !ok {
		log.Fatalln("Could not find 'username' entry for elasticsearch8")
	}
	password, ok := config["password"].(string)
	if !ok {
		log.Fatalln("Could not find 'password' entry for elasticsearch8")
	}

	es8rec := new(Elastic8Recorder)
	cfg := elasticsearch8.Config{
		Addresses: []string{
			host,
		},
		Username: username,
		Password: password,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: time.Second,
			DialContext:           (&net.Dialer{Timeout: time.Second}).DialContext,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	client, _ := elasticsearch8.NewClient(cfg)
	es8rec.client = client

	log.Println("Created elasticsearch8 recorder")

	return es8rec
}
