package container

import (
	"testing"
)

func TestNewRing(t *testing.T) {
	r := NewRing(10)
	go func(r *Ring) {
		r.Set([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	}(r)
	go func(r *Ring) {
		r.Set([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	}(r)

	t.Logf("%+v", r)
	t.Log(r.Get())
	t.Log(len(r.Get()))
	t.Log(len(r.Get()))
	t.Logf("%+v", r)
}
