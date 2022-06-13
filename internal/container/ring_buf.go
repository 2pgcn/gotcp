package container

import (
	errors "github.com/php403/gotcp/internal/error"
	"sync/atomic"
)

type Ring struct {
	// read
	r     uint64 //读位置
	w     uint64 //写位置
	count uint64
	data  [][]byte
}

func NewRing(num int) *Ring {
	r := new(Ring)
	r.init(num)
	return r
}

//最贴近2^n
func (r *Ring) init(num int) {
	if num&(num-1) != 0 {
		for num&(num-1) != 0 {
			num &= num - 1
		}
		num <<= 1
	}
	r.data = make([][]byte, num)
}

func (r *Ring) Get() (res []byte) {
	readNum, writeNum, count := atomic.LoadUint64(&r.r), atomic.LoadUint64(&r.w), atomic.LoadUint64(&r.count)
	if writeNum == readNum && count == 0 {
		return
	}
	if !atomic.CompareAndSwapUint64(&r.r, readNum, (readNum+1)%(uint64(len(r.data))-1)) {
		return
	}
	res = r.data[readNum]
	r.data[readNum] = []byte{}
	atomic.AddUint64(&r.count, ^uint64(0))
	return
}

func (r *Ring) Set(data []byte) (err error) {
	readNum, writeNum, count := atomic.LoadUint64(&r.r), atomic.LoadUint64(&r.w), atomic.LoadUint64(&r.count)
	if writeNum == readNum && count > 0 {
		return errors.ErrRingWrite
	}
	writeNumNew := writeNum % (uint64(len(r.data)) - 1)
	r.data[writeNum] = data
	atomic.CompareAndSwapUint64(&r.w, writeNum, writeNumNew)
	atomic.AddUint64(&r.count, 1)
	return
}
