package watcher

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// https://godoc.org/github.com/google/gopacket
// https://godoc.org/github.com/google/gopacket/pcap

const portCountLength = 5

func logPacket(protocol gopacket.LayerType, port uint16, outFile *os.File) {

	lineLength := (int64)(6 + portCountLength + 1 + portCountLength + 1)
	offset := (int64)(port-1)*lineLength + 6
	outFile.Seek(offset, 0)

	intBuffer := make([]byte, portCountLength)
	if protocol == layers.LayerTypeTCP {
		outFile.Read(intBuffer)
	} else if protocol == layers.LayerTypeUDP {
		outFile.Seek((int64)(portCountLength+1), 1)
		outFile.Read(intBuffer)
	} else {
		return
	}
	intCount, err := strconv.Atoi(strings.TrimSpace(string(intBuffer)))
	if err != nil {
		log.Fatalf("missed.txt is corrupted! Got %s\n", string(intBuffer))
		return
	}

	intCount++

	outFile.Seek(-portCountLength, 1)
	fmt.Fprintf(outFile, "%"+strconv.Itoa(portCountLength)+"d", intCount)
	outFile.Sync()

}

func watcherRun(iface string, pcapFilter string, contChan chan bool) {

	missedPath := "missed.txt"

	missedFile, err := os.OpenFile(missedPath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatalf("Could not open missed file: %s\n", err)
		return
	}
	defer missedFile.Close()

	stat, err := missedFile.Stat()
	if err != nil {
		log.Fatalf("Could not stat missed file: %s\n", err)
		return
	}

	lineLength := 6 + portCountLength + 1 + portCountLength + 1
	maxSize := (int64)(lineLength) * 65535

	if stat.Size() < maxSize {
		log.Println("Setting up missed.txt")
		var i int64
		for i = 0; i < 65535; i++ {
			fmt.Fprintf(missedFile, "%5d ", i+1)
			fmt.Fprintf(missedFile, "%"+strconv.Itoa(portCountLength)+"d", 0)
			missedFile.WriteString("|")
			fmt.Fprintf(missedFile, "%"+strconv.Itoa(portCountLength)+"d", 0)
			missedFile.WriteString("\n")
		}
	}

	pcapHandle, err := pcap.OpenLive(iface, 1600, true, pcap.BlockForever)

	if err != nil {
		log.Fatalf("Could not open interface %s for listening: %s\n", iface, err)
		return
	}

	err = pcapHandle.SetBPFFilter(pcapFilter)

	if err != nil {
		log.Fatalf("Could not set filter %s for interface %s: %s\n", pcapFilter, iface, err)
		return
	}

	log.Println("Missed port watcher listening...")
	contChan <- true

	var tcp layers.TCP
	var udp layers.UDP
	var eth layers.Ethernet
	var ip4 layers.IPv4
	var ip6 layers.IPv6

	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ip4, &ip6, &tcp, &udp)
	decodedPacket := []gopacket.LayerType{}

	pcapPacket := gopacket.NewPacketSource(pcapHandle, pcapHandle.LinkType())
	for packet := range pcapPacket.Packets() {

		err = parser.DecodeLayers(packet.Data(), &decodedPacket)
		// if err != nil {
		// 	fmt.Printf("err: %s\n", err)
		// 	continue
		// }

		for _, layerType := range decodedPacket {
			if layerType == layers.LayerTypeTCP {
				logPacket(layers.LayerTypeTCP, (uint16)(tcp.DstPort), missedFile)
			} else if layerType == layers.LayerTypeUDP {
				logPacket(layers.LayerTypeUDP, (uint16)(udp.DstPort), missedFile)
			}
		}

	}

}

// StartWatcher starts the missed port watcher
func StartWatcher(iface string, pcapFilter string, contChan chan bool) {

	newFilter := pcapFilter

	// Get interface address info so we don't intercept data sent by us
	ifaceInfo, err := net.InterfaceByName(iface)
	if err != nil {
		log.Fatalf("Could not get interface %s\n", iface)
		return
	}

	addrs, err := ifaceInfo.Addrs()
	if err != nil {
		log.Fatalf("Could not get interface %s addresses\n", iface)
		return
	}

	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if newFilter != "" {
			newFilter += " and not src host " + ip.String()
		} else {
			newFilter += "not src host " + ip.String()
		}

	}

	log.Printf("Watcher filter: \n\n%s\n\n", newFilter)

	log.Println("Starting missed port watcher...")
	go watcherRun(iface, newFilter, contChan)
}
