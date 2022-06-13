package container

import (
	"golang.org/x/sys/cpu"
	"unsafe"
)

const CacheLinePadSize = unsafe.Sizeof(cpu.CacheLinePad{})

type (
	Queue interface {
		Enqueue(item interface{}) (err error)
		Dequeue() (item interface{}, err error)
		Cap() uint32
		Size() uint32
		IsEmpty() (b bool)
		IsFull() (b bool)
	}

	RingBuffer interface {
		Queue
		Put(item interface{}) (err error)
		Get() (item interface{}, err error)
		Quantity() uint32
		Debug(enabled bool) (lastState bool)
		ResetCounters()
	}

	ringBuf struct {
		cap        uint32
		capModMask uint32
		head       uint32
		tail       uint32
		putWaits   uint64
		getWaits   uint64
		_          [CacheLinePadSize]byte
		data       []rbItem
		debugMode  bool
	}

	rbItem struct {
		readWrite uint64      // 0: writable, 1: readable, 2: write ok, 3: read ok
		value     interface{} // ptr
		_         [CacheLinePadSize - 8 - 8]byte
		// _         cpu.CacheLinePad
	}
)
