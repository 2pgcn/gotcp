package container

import (
	errors "github.com/php403/gotcp/internal/error"
)

// Ring
type Ring struct {
	// read
	r     uint64 //读位置
	w     uint64 //写位置
	count uint64
	buf   [][]byte
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
	r.buf = make([][]byte, num)
}

func (r *Ring) Get() (res []byte) {
	if r.r == r.w {
		return
	}
	if r.count == 0 {
		return
	}
	//todo mutex
	res = r.buf[r.r]
	r.buf[r.r] = []byte{}
	r.r += 1
	r.r %= uint64(len(r.buf))
	r.count--
	return
}

func (r *Ring) Set(data []byte) (err error) {
	if r.w-r.r >= uint64(len(r.buf)) {
		return errors.ErrRingFull
	}
	r.buf[r.w] = data
	r.w += 1
	r.w %= uint64(len(r.buf))
	r.count++
	return
}
