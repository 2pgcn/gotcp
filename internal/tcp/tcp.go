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
var inItTcbTableLock sync.RWMutex
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
	TcbLock          sync.RWMutex
}

//tcb
type tcbTable struct {
	tcbHead *Tcb
	tcbTail *Tcb
	count   int
}

func InitTcbTable() {
	inItTcbTableLock.Lock()
	defer inItTcbTableLock.Unlock()
	once.Do(func() {
		inItTcbTable = &tcbTable{
			tcbHead: nil,
			count:   0,
		}
	})
}

func GetTcbTable() *tcbTable {
	return inItTcbTable
}

func (t *tcbTable) GetFirstTcb() *Tcb {
	inItTcbTableLock.RLock()
	defer inItTcbTableLock.RUnlock()
	return t.tcbHead
}

// SearchTcb todo 读锁
func (t *tcbTable) SearchTcb(SrcIP, DstIP net.IP, SrcPort, DstPort layers.TCPPort) *Tcb {
	item := t.tcbHead
	if item == nil {
		return nil
	}
	item.TcbLock.RLock()
	for item != nil {
		if item.SrcIP.Equal(SrcIP) && item.DstIP.Equal(DstIP) && item.SrcPort == SrcPort && item.DstPort == DstPort {
			return item
		}
		item = item.Next
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
	inItTcbTableLock.Lock()
	defer inItTcbTableLock.Unlock()
	if tcbTable.tcbHead == nil {
		tcbTable.tcbHead = tcb
		tcbTable.tcbTail = tcb
	} else {
		tcb.Prev = tcbTable.tcbTail
		tcbTable.tcbTail.Next = tcb
	}
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
	tcb.TcbLock.Lock()
	defer tcb.TcbLock.Unlock()
	tcb.tcpStatus = status
}

func (tcb *Tcb) GetTcpInitPackage() *layers.TCP {
	return &layers.TCP{}
}
