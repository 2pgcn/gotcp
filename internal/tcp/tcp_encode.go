package tcp

import (
	"context"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"log"
	"math/rand"
	"net"
)

// TcpMaxSeq
// It is essential to remember that the actual sequence number space is
//  finite, though very large.  This space ranges from 0 to 2**32 - 1.
const TcpMaxSeq = 2<<31 - 1
const RecBufLen = 1024
const SendBufLen = 1024

func StateMachine(ctx context.Context, SrcIP, DstIP net.IP, tcp *layers.TCP) (err error) {
	//处理tcp
	tcb := inItTcbTable.SearchTcb(SrcIP, DstIP, tcp.SrcPort, tcp.DstPort)
	if tcb == nil {
		tcb = inItTcbTable.CreateTcb(SrcIP, DstIP, tcp.SrcPort, tcp.DstPort)
	}
	replyBuf := gopacket.NewSerializeBuffer()
	opt := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	switch tcb.tcpStatus {
	case GO_TCP_STATUS_CLOSED: //client
		break
	case GO_TCP_STATUS_LISTEN: // server
		if tcp.SYN {
			replyIp4 := getIp4TcpPackage(DstIP, SrcIP)
			//发送ACK SYN
			Seq := rand.Uint32()
			replyTcp := &layers.TCP{
				SrcPort:    tcp.DstPort,
				DstPort:    tcp.SrcPort,
				Seq:        Seq % TcpMaxSeq,
				Ack:        tcp.Seq + 1,
				DataOffset: 0,
				FIN:        false,
				SYN:        true,
				RST:        false,
				PSH:        false,
				ACK:        true,
				URG:        false,
				ECE:        false,
				CWR:        false,
				NS:         false,
				Window:     1504,
				Urgent:     0,
				Options:    nil,
				Padding:    nil,
			}
			err = replyTcp.SetNetworkLayerForChecksum(replyIp4)
			if err != nil {
				return err
			}
			err = gopacket.SerializeLayers(replyBuf, opt, replyIp4, replyTcp, gopacket.Payload(replyTcp.Payload))
			if err != nil {
				return
			}
			log.Default().Printf("replytcp:%+v", replyTcp)
			tcb.ChangeTcpStatus(GO_TCP_STATUS_SYN_RCVD)
			tcb.Seq = replyTcp.Seq
			tcb.Ack = replyTcp.Ack
			err = tcb.SendBuf.Set(replyBuf.Bytes())
			if err != nil {
				return
			}
			fmt.Printf("GO_TCP_STATUS_LISTEN ---> GO_TCP_STATUS_SYN_RCVD")
		}
		break
	case GO_TCP_STATUS_SYN_RCVD: // server
		if tcp.ACK {
			//客户端回复ACK
			if tcp.Ack == tcb.Seq+1 {
				tcb.ChangeTcpStatus(GO_TCP_STATUS_ESTABLISHED)
				fmt.Printf("GO_TCP_STATUS_SYN_RCVD ---> GO_TCP_STATUS_ESTABLISHED")
			}
		}
		break

	case GO_TCP_STATUS_SYN_SENT: // client
		break

	case GO_TCP_STATUS_ESTABLISHED:
		{ // server | client
			fmt.Printf("go pkg :%+v", tcp)
			break
		}
	case GO_TCP_STATUS_FIN_WAIT_1: //  ~client
		break

	case GO_TCP_STATUS_FIN_WAIT_2: // ~client
		break

	case GO_TCP_STATUS_CLOSING: // ~client
		break

	case GO_TCP_STATUS_TIME_WAIT: // ~client
		break

	case GO_TCP_STATUS_CLOSE_WAIT: // ~server
		break

	case GO_TCP_STATUS_LAST_ACK: // ~server
		break

	}
	return
}

func getIp4TcpPackage(SrcIP, DstIP net.IP) *layers.IPv4 {
	return &layers.IPv4{
		Version:    4,
		IHL:        5,
		TOS:        0,
		Id:         0,
		Flags:      0,
		FragOffset: 0,
		TTL:        255,
		Protocol:   layers.IPProtocolTCP,
		SrcIP:      SrcIP,
		DstIP:      DstIP,
	}
}
