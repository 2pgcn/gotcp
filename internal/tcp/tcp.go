package tcp

import (
	"github.com/google/gopacket/layers"
	"github.com/php403/gotcp/internal/container"
	"net"
	"sync"
)

type TcpStatus uint8

//tcp状态机
const (
	GO_TCP_STATUS_CLOSED TcpStatus = iota
	GO_TCP_STATUS_LISTEN
	GO_TCP_STATUS_SYN_RCVD
	GO_TCP_STATUS_SYN_SENT
	GO_TCP_STATUS_ESTABLISHED
	GO_TCP_STATUS_FIN_WAIT_1
	GO_TCP_STATUS_FIN_WAIT_2
	GO_TCP_STATUS_CLOSING
	GO_TCP_STATUS_TIME_WAIT
	GO_TCP_STATUS_CLOSE_WAIT
	GO_TCP_STATUS_LAST_ACK
)

//todo 修改需要加锁
var inItTcbTable *tcbTable
var once sync.Once

type Tcb struct {
	//todo fd 改成io.Reader
	fd               int
	SrcIP, DstIP     net.IP
	SrcPort, DstPort layers.TCPPort
	tcpStatus        TcpStatus
	SendBuf, RecBuf  *container.Ring
	Next, Prev       *Tcb
	Seq, Ack         uint32
}

//tcb
type tcbTable struct {
	tcbHead *Tcb
	count   int
}

func GetTcbTable() *tcbTable {
	once.Do(func() {
		inItTcbTable = &tcbTable{
			tcbHead: nil,
			count:   0,
		}
	})
	return inItTcbTable
}

func (t *tcbTable) GetFirstTcb() *Tcb {
	return t.tcbHead
}

// SearchTcb todo 读锁
func (t *tcbTable) SearchTcb(SrcIP, DstIP net.IP, SrcPort, DstPort layers.TCPPort) *Tcb {
	for item := t.tcbHead; item != nil; item = item.Next {
		if item.SrcIP.Equal(SrcIP) && item.DstIP.Equal(DstIP) && item.SrcPort == SrcPort && item.DstPort == DstPort {
			return item
		}
	}
	return nil
}

// CreateTcb todo 加读写锁
func (t *tcbTable) CreateTcb(SrcIP, DstIP net.IP, SrcPort, DstPort layers.TCPPort) *Tcb {
	tcb := &Tcb{
		SrcIP:     SrcIP,
		DstIP:     DstIP,
		SrcPort:   SrcPort,
		DstPort:   DstPort,
		tcpStatus: GO_TCP_STATUS_LISTEN,
		RecBuf:    container.NewRing(RecBufLen),
		SendBuf:   container.NewRing(SendBufLen),
	}
	tcbTable := GetTcbTable()
	tcb.Next = tcbTable.tcbHead
	tcbTable.tcbHead = tcb
	return tcb
}

// RemoveTcb todo 读写锁
func (t *tcbTable) RemoveTcb(tcb *Tcb) *Tcb {
	if tcb.Prev != nil {
		tcb.Prev.Next = tcb.Next
	}
	if tcb.Next != nil {
		tcb.Next.Prev = tcb.Prev
	}
	return tcb
}

func (tcb *Tcb) ChangeTcpStatus(status TcpStatus) {
	tcb.tcpStatus = status
}
