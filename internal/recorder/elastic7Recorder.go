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

	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	esapi "github.com/elastic/go-elasticsearch/v7/esapi"
)

type Elastic7Recorder struct {
	client *elasticsearch7.Client
}

func (r Elastic7Recorder) Record(record *HoneypokeRecord) error {

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

func NewElastic7Recorder(config map[string]interface{}) *Elastic7Recorder {

	host, ok := config["host"].(string)
	if !ok {
		log.Fatalln("Could not find 'host' entry for elasticsearch7")
	}
	username, ok := config["username"].(string)
	if !ok {
		log.Fatalln("Could not find 'username' entry for elasticsearch7")
	}
	password, ok := config["password"].(string)
	if !ok {
		log.Fatalln("Could not find 'password' entry for elasticsearch7")
	}

	es6rec := new(Elastic7Recorder)
	cfg := elasticsearch7.Config{
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

	client, _ := elasticsearch7.NewClient(cfg)
	es6rec.client = client

	log.Println("Created elasticsearch7 recorder")

	return es6rec
}
