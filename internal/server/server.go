package server

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/bocajspear1/honeypoke-go/internal/recorder"
)

const maxTCPSize = 8192

func tcpHandler(conn net.Conn, c chan *recorder.HoneypokeRecord) {

	addrSplit := strings.Split(conn.RemoteAddr().String(), ":")

	remoteAddr := addrSplit[0]
	remotePort, err := (strconv.Atoi(addrSplit[1]))
	if err != nil {
		remotePort = 0
	}

	fmt.Printf("Connection from %s:%d\n", remoteAddr, remotePort)

	const chunkSize = 512

	finalBuffer := make([]byte, 0)

	for bytesReadTotal := 0; bytesReadTotal <= maxTCPSize; {
		smallBuffer := make([]byte, chunkSize)

		bytesRead, err := conn.Read(smallBuffer)

		if bytesRead > 0 {
			bytesReadTotal += bytesRead
			fmt.Println("Read Data...")
			finalBuffer = append(finalBuffer, smallBuffer[0:bytesReadTotal]...)
		}

		if err != nil {
			fmt.Println("Connection error, closing...")

			record := recorder.NewRecord(remoteAddr, (uint16)(remotePort))

			input := strconv.Quote(string(finalBuffer))
			record.Input = input

			c <- record

			return
		}
	}

	fmt.Println("Max buffer reached")

	conn.Close()

}

func runTCPServer(port int, ssl bool, recChan chan *recorder.HoneypokeRecord, contChan chan bool) {
	log.Printf("Started server for port %d", port)
	listener, err := net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(port))
	if err != nil {
		log.Fatalf("Failed to listen on port %d\n", port)
		return
	}
	defer listener.Close()

	contChan <- true

	for syscall.Getuid() == 0 {
		log.Printf("Waiting for permissions to drop...")
		time.Sleep(time.Second * 2)
	}

	for {
		conn, aerr := listener.Accept()
		fmt.Printf("accept")
		if aerr != nil {
			fmt.Printf("failed connection")
			return
		}
		//It's common to handle accepted connection on different goroutines
		go tcpHandler(conn, recChan)
	}

}

func runUDPServer(port int) {

}

// StartServer starts a listener on a port
func StartServer(protocol gopacket.LayerType, port int, ssl bool, recChan chan *recorder.HoneypokeRecord, contChan chan bool) {
	if protocol == layers.LayerTypeTCP {
		go runTCPServer(port, ssl, recChan, contChan)
	} else if protocol == layers.LayerTypeUDP {
		go runUDPServer(port)
	}
}
