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

	elasticsearch6 "github.com/elastic/go-elasticsearch/v6"
	esapi "github.com/elastic/go-elasticsearch/v6/esapi"
)

type Elastic6Recorder struct {
	client *elasticsearch6.Client
}

func (r Elastic6Recorder) Record(record *HoneypokeRecord) error {

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

func NewElastic6Recorder(config map[string]interface{}) *Elastic6Recorder {

	host, ok := config["host"].(string)
	if !ok {
		log.Fatalln("Could not find 'host' entry for elasticsearch6")
	}
	username, ok := config["username"].(string)
	if !ok {
		log.Fatalln("Could not find 'username' entry for elasticsearch6")
	}
	password, ok := config["password"].(string)
	if !ok {
		log.Fatalln("Could not find 'password' entry for elasticsearch6")
	}

	es6rec := new(Elastic6Recorder)
	cfg := elasticsearch6.Config{
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

	client, _ := elasticsearch6.NewClient(cfg)
	es6rec.client = client

	log.Println("Created elasticsearch6 recorder")

	return es6rec
}
