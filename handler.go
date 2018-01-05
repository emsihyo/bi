package bi

import (
	"sync"
	"sync/atomic"
)

//Handler Handler
type handler struct {
	id    uint64
	calls map[uint64]chan []byte
	mut   sync.RWMutex
}

func newHandler() *handler {
	return &handler{calls: map[uint64]chan []byte{}}
}

func (hand *handler) nextCallID() uint64 {
	return atomic.AddUint64(&hand.id, 1)
}

func (hand *handler) addCall(id uint64, callback chan []byte) {
	hand.mut.Lock()
	hand.calls[id] = callback
	hand.mut.Unlock()
}

func (hand *handler) removeCall(id uint64) {
	hand.mut.Lock()
	delete(hand.calls, id)
	hand.mut.Unlock()
}

func (hand *handler) getCall(id uint64) chan []byte {
	hand.mut.RLock()
	defer hand.mut.RUnlock()
	return hand.calls[id]
}
