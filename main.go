package main

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/songgao/water"
	"log"
	"os"
	"os/exec"
)

const (
	BUFFERSIZE = 100
	MTU        = "1504"
)

func main() {
	iface, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		panic(err)
	}
	log.Printf("tun interface: %s", iface.Name())
	runBin("/bin/ip", "link", "set", "dev", iface.Name(), "mtu", MTU)
	runBin("/bin/ip", "addr", "add", "192.168.2.0/24", "dev", iface.Name())
	runBin("/bin/ip", "link", "set", "dev", iface.Name(), "up")
	replyBuf := gopacket.NewSerializeBuffer()
	opt := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	buf := make([]byte, BUFFERSIZE)
	for {
		_, err := iface.Read(buf)
		if err != nil {
			panic(err)
		}
		ip4, ok := gopacket.NewPacket(buf, layers.LayerTypeIPv4, gopacket.Default).Layer(layers.LayerTypeIPv4).(*layers.IPv4)
		if !ok {
			continue
		}
		fmt.Println(ip4.LayerType(), ip4.Protocol)

		switch ip4.Protocol {
		case layers.IPProtocolICMPv4:
			{
				icmp4, ok := gopacket.NewPacket(ip4.LayerPayload(), layers.LayerTypeICMPv4, gopacket.Default).Layer(layers.LayerTypeICMPv4).(*layers.ICMPv4)
				if !ok {
					continue
				}
				if icmp4.TypeCode != layers.CreateICMPv4TypeCode(layers.ICMPv4TypeEchoRequest, 0) {
					continue
				}
				//返回ipv4
				replyIp4 := &layers.IPv4{
					Version:    4,
					IHL:        5,
					TOS:        0,
					Id:         0,
					Flags:      0,
					FragOffset: 0,
					TTL:        255,
					Protocol:   layers.IPProtocolICMPv4,
					SrcIP:      ip4.DstIP,
					DstIP:      ip4.SrcIP,
				}
				replyIcmp4 := &layers.ICMPv4{
					TypeCode: layers.CreateICMPv4TypeCode(layers.ICMPv4TypeEchoReply, 0),
					Id:       icmp4.Id,
					Seq:      icmp4.Seq,
				}
				err = gopacket.SerializeLayers(replyBuf, opt, replyIp4, replyIcmp4, gopacket.Payload(icmp4.Payload))
				if err != nil {
					log.Printf("err:%v", err)
					continue
				}

				_, err = iface.Write(replyBuf.Bytes())

				if err != nil {
					log.Printf("err:%v", err)
					continue
				}

			}
		case layers.IPProtocolTCP:
			//tcp, ok := gopacket.NewPacket(ip4.LayerPayload(), layers.LayerTypeICMPv4, gopacket.Default).Layer(layers.LayerTypeICMPv4).(*layers.TCP)
			//if !ok {
			//	continue
			//}

		}
	}
}

func runBin(bin string, args ...string) {
	cmd := exec.Command(bin, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
