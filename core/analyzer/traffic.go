package analyzer

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"gfa/common/kafka"
	"gfa/common/log"
	"gfa/common/utils"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

type traffic struct {
	analyzer
	FrameLength int               `json:"frame_length"`
	SourceIP    string            `json:"source_ip"`
	DestIP      string            `json:"dest_ip"`
	SourcePort  string            `json:"source_port"`
	DestPort    string            `json:"dest_port"`
	Protocol    string            `json:"protocol"`
	TcpFlags    string            `json:"tcp_flags,omitempty"`
	Service     string            `json:"service,omitempty"`
	URL         string            `json:"request_http_url,omitempty"`
	Path        string            `json:"request_http_path,omitempty"`
	Method      string            `json:"request_http_method,omitempty"`
	Body        string            `json:"request_http_body,omitempty"`
	Header      map[string]string `json:"request_http_header,omitempty"`
}

func (t *traffic) Do() {
	var (
		timeNow        = time.Now()
		shouldContinue = true
	)
	handle, err := pcap.OpenOffline(fmt.Sprintf("%s%s", dirPath, t.Pcap))
	if err != nil {
		log.Errorf("open pcap file %s err:%s", t.Pcap, err)
		shouldContinue = false
	}
	if shouldContinue {
		var (
			packetChan = make(chan gopacket.Packet, 500000)
			wg         sync.WaitGroup
		)
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		defer func() {
			close(packetChan)
			wg.Wait()
			handle.Close()
			log.Infof("parse %s task time %vs", t.Pcap, utils.SubtractTime(timeNow, time.Now()))
		}()
		for i := 0; i < pcapWorkers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for packet := range packetChan {
					var (
						url      string
						path     string
						method   string
						body     string
						tcpFlags string
						header   = make(map[string]string)
					)
					ethLayer := packet.Layer(layers.LayerTypeEthernet)
					ipLayer := packet.Layer(layers.LayerTypeIPv4)
					tcpLayer := packet.Layer(layers.LayerTypeTCP)
					udpLayer := packet.Layer(layers.LayerTypeUDP)
					if ethLayer != nil && ipLayer != nil {
						ipPacket, _ := ipLayer.(*layers.IPv4)
						srcIP := ipPacket.SrcIP.String()
						dstIP := ipPacket.DstIP.String()
						protocol := ipPacket.Protocol.String()
						srcPort := ""
						dstPort := ""
						if tcpLayer != nil {
							tcpPacket, _ := tcpLayer.(*layers.TCP)
							srcPort = tcpPacket.SrcPort.String()
							dstPort = tcpPacket.DstPort.String()
							if tcpPacket.FIN {
								tcpFlags += "FIN "
							}
							if tcpPacket.SYN {
								tcpFlags += "SYN "
							}
							if tcpPacket.RST {
								tcpFlags += "RST "
							}
							if tcpPacket.PSH {
								tcpFlags += "PSH "
							}
							if tcpPacket.ACK {
								tcpFlags += "ACK "
							}
							if tcpPacket.URG {
								tcpFlags += "URG "
							}
							if tcpPacket.ECE {
								tcpFlags += "ECE "
							}
							if tcpPacket.CWR {
								tcpFlags += "CWR "
							}
						} else if udpLayer != nil {
							udpPacket, _ := udpLayer.(*layers.UDP)
							srcPort = udpPacket.SrcPort.String()
							dstPort = udpPacket.DstPort.String()
						}
						frameLen := packet.Metadata().CaptureLength
						newSrcPort, newDestPort, newService := extractPortsAndProtocol(srcPort, dstPort)
						applicationLayer := packet.ApplicationLayer()
						if applicationLayer != nil {
							payload := applicationLayer.Payload()
							if len(payload) > 0 {
								func() {
									httpData, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(payload)))
									if err == nil {
										defer httpData.Body.Close()
										url = httpData.URL.String()
										path = httpData.URL.Path
										method = httpData.Method
										for name, headers := range httpData.Header {
											for _, h := range headers {
												header[name] = h
											}
										}
										if httpData.Body != nil {
											b, _ := ioutil.ReadAll(httpData.Body)
											body = string(b)
										}
									}
								}()
							}
						}
						result, _ := json.Marshal(&traffic{
							analyzer: analyzer{
								Pcap:      t.Pcap,
								Location:  location,
								TimeStamp: utils.TimeFormatToKafka(t.Pcap),
							},
							URL:         url,
							Method:      method,
							Path:        path,
							Header:      header,
							Body:        body,
							Protocol:    protocol,
							SourceIP:    srcIP,
							SourcePort:  newSrcPort,
							DestIP:      dstIP,
							DestPort:    newDestPort,
							TcpFlags:    tcpFlags,
							Service:     newService,
							FrameLength: frameLen * 8,
						})
						kafka.Q(result)
					}
				}
			}()
		}
		for packet := range packetSource.Packets() {
			packetChan <- packet
		}
	}
}

func extractPortsAndProtocol(src, dest string) (string, string, string) {
	var isContains func(string) bool
	var matchProtocol func(string) (string, string)
	isContains = func(str string) bool {
		return strings.Contains(str, "(")
	}
	matchProtocol = func(str string) (string, string) {
		rePort := regexp.MustCompile(`\d+`)
		port := rePort.FindString(str)
		reProtocol := regexp.MustCompile(`\(\w+\)`)
		protocol := reProtocol.FindString(str)
		if protocol != "" {
			protocol = protocol[1 : len(protocol)-1]
		}
		return port, protocol
	}
	if !isContains(src) && !isContains(dest) {
		return src, dest, ""
	}
	if !isContains(src) && isContains(dest) {
		dstPort, protocol := matchProtocol(dest)
		return src, dstPort, protocol
	}
	if isContains(src) && !isContains(dest) {
		srcPort, protocol := matchProtocol(src)
		return srcPort, dest, protocol
	}
	if isContains(src) && isContains(dest) {
		srcPort, _ := matchProtocol(src)
		dstPort, _ := matchProtocol(dest)
		return srcPort, dstPort, fmt.Sprintf("%s/%s", src[strings.Index(src, "(")+1:strings.Index(src, ")")], dest[strings.Index(dest, "(")+1:strings.Index(dest, ")")])
	}
	return "", "", ""
}
