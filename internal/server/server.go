/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package server

import (
	"crypto/tls"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/bocajspear1/honeypoke-go/internal/recorder"
)

const toFileSize = 4096

// Max 35k files
const maxTCPSize = 40 * 1024

func tcpHandler(port int, conn net.Conn, c chan *recorder.HoneypokeRecord) {

	addrSplit := strings.Split(conn.RemoteAddr().String(), ":")

	remoteAddr := addrSplit[0]
	remotePort, err := (strconv.Atoi(addrSplit[1]))
	if err != nil {
		remotePort = 0
	}

	const chunkSize = 512

	connDied := false
	finalBuffer := make([]byte, 0)

	var outFile *os.File
	var outPath string

	for bytesReadTotal := 0; bytesReadTotal <= maxTCPSize && !connDied; {
		smallBuffer := make([]byte, chunkSize)

		bytesRead, err := conn.Read(smallBuffer)

		if bytesRead > 0 {
			bytesReadTotal += bytesRead
			if bytesReadTotal < toFileSize {
				finalBuffer = append(finalBuffer, smallBuffer[0:bytesRead]...)
			} else {
				if outFile == nil {
					outPath = "./large/tcp-" + strconv.Itoa(port) + "-" + strconv.FormatInt(time.Now().Unix(), 10) + ".large"
					outFile, err = os.OpenFile(outPath, os.O_RDWR|os.O_CREATE, 0444)
					if err != nil {
						log.Printf("Could not open large file: %s\n", err)
						conn.Close()
						connDied = true
					}
					defer outFile.Close()
					outFile.Write(finalBuffer)
				}

				outFile.Write(smallBuffer[0:bytesRead])
			}

		}

		if err != nil {
			// Maybe do something if the error is of a certian type?
			connDied = true
		}
	}

	if !connDied {
		log.Println("Max buffer reached")
		conn.Close()
	}

	record := recorder.NewRecord(remoteAddr, (uint16)(remotePort))

	input := strconv.Quote(string(finalBuffer))
	if outFile == nil {
		record.Input = input[1 : len(input)-1]
	} else {
		record.Input = "Input sent to file " + outPath
	}

	record.RemoteIP = remoteAddr
	record.RemotePort = remotePort
	record.Port = port
	record.Protocol = "tcp"

	c <- record

}

func runTCPServer(port int, ssl bool, recChan chan *recorder.HoneypokeRecord, contChan chan bool) {
	log.Printf("Started server for port %d", port)

	var listener net.Listener

	if !ssl {
		var err error
		listener, err = net.Listen("tcp", ":"+strconv.Itoa(port))
		if err != nil {
			log.Fatalf("Failed to start TCP server on port %d\n", port)
			return
		}
	} else {
		certs, err := tls.LoadX509KeyPair("honeypoke_cert.pem", "honeypoke_key.pem")
		if err != nil {
			log.Println(err)
			return
		}
		config := &tls.Config{Certificates: []tls.Certificate{certs}}
		listener, err = tls.Listen("tcp", ":"+strconv.Itoa(port), config)
	}

	defer listener.Close()

	contChan <- true

	for syscall.Getuid() == 0 {
		log.Printf("TCP:%d Waiting for permissions to drop...", port)
		time.Sleep(time.Second * 2)
	}

	for {
		conn, aerr := listener.Accept()

		if aerr != nil {
			log.Printf("Failed connection")
			return
		}

		go tcpHandler(port, conn, recChan)
	}

}

func runUDPServer(port int, recChan chan *recorder.HoneypokeRecord, contChan chan bool) {
	udpList, err := net.ListenPacket("udp", ":"+strconv.Itoa(port))

	contChan <- true

	for syscall.Getuid() == 0 {
		log.Printf("UDP:%d Waiting for permissions to drop...", port)
		time.Sleep(time.Second * 2)
	}

	if err != nil {
		log.Fatal(err)
	}
	defer udpList.Close()

	buffer := make([]byte, 2048)

	for {

		bytesRead, remoteAddrData, err := udpList.ReadFrom(buffer)

		addrSplit := strings.Split(remoteAddrData.String(), ":")

		remoteAddr := addrSplit[0]
		remotePort, err := (strconv.Atoi(addrSplit[1]))
		if err != nil {
			remotePort = 0
		}

		if err != nil {
			log.Printf("Error getting packet: %s", err)

		} else if bytesRead > 0 {
			record := recorder.NewRecord(remoteAddr, (uint16)(remotePort))

			input := strconv.Quote(string(buffer[0:bytesRead]))
			record.Input = input[1 : len(input)-1]
			record.RemoteIP = remoteAddr
			record.RemotePort = remotePort
			record.Port = port
			record.Protocol = "udp"

			recChan <- record
		}
	}

	//simple read

}

// StartServer starts a listener on a port
func StartServer(protocol gopacket.LayerType, port int, ssl bool, recChan chan *recorder.HoneypokeRecord, contChan chan bool) {
	if protocol == layers.LayerTypeTCP {
		go runTCPServer(port, ssl, recChan, contChan)
	} else if protocol == layers.LayerTypeUDP {
		go runUDPServer(port, recChan, contChan)
	}
}
