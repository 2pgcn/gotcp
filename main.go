package main

import (
	"context"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/php403/gotcp/internal/tcp"
	"github.com/songgao/water"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

const (
	BUFFERSIZE = 100
	MTU        = "1504"
)

func main() {
	rand.Seed(time.Now().UnixNano())
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
	ctx := context.Background()
	buf := make([]byte, BUFFERSIZE)
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				log.Printf("ctx done 1111")
				return
			default:
				_, err := iface.Read(buf)
				if err != nil {
					panic(err)
				}
				if len(buf) == 0 {
					continue
				}
				ip4, ok := gopacket.NewPacket(buf, layers.LayerTypeIPv4, gopacket.Default).Layer(layers.LayerTypeIPv4).(*layers.IPv4)
				if !ok {
					continue
				}
				switch ip4.Protocol {
				case layers.IPProtocolICMPv4:
					{
						replyBuf := gopacket.NewSerializeBuffer()
						opt := gopacket.SerializeOptions{
							FixLengths:       true,
							ComputeChecksums: true,
						}
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
					tcpLayers, ok := gopacket.NewPacket(ip4.LayerPayload(), layers.LayerTypeTCP, gopacket.Default).Layer(layers.LayerTypeTCP).(*layers.TCP)
					if !ok {
						continue
					}
					log.Default().Printf("tcp ack:%v,syn:%v\n", tcpLayers.Ack, tcpLayers.SYN)
					err = tcp.StateMachine(ctx, ip4.SrcIP, ip4.DstIP, tcpLayers)
					if err != nil {
						log.Default().Printf("tcp2 StateMachine err:%v", err)
					}
				}
			}
		}
	}(ctx)

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				tcbHead := tcp.GetTcbTable().GetFirstTcb()
				if tcbHead == nil {
					break
				}
				for tcbHead != nil {
					for sendData := tcbHead.SendBuf.Get(); len(sendData) > 0; {
						_, err = iface.Write(sendData)
						sendData = tcbHead.SendBuf.Get()
						if err != nil {
							log.Default().Printf("iface Write err:%v", err)
							continue
						}
					}
					tcbHead = tcbHead.Next
				}

			}
		}
	}(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			ctx.Done()
			return
		case syscall.SIGHUP:
		default:
			return
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
