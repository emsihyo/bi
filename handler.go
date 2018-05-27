package bi

import (
	"sync"
	"sync/atomic"
)

//handler handler
type handler struct {
	id    uint64
	calls map[uint64]chan []byte
	mut   sync.RWMutex
}

func newHandler() *handler {
	return &handler{calls: map[uint64]chan []byte{}}
}

func (hand *handler) addCall(callback chan []byte) uint64 {
	id := atomic.AddUint64(&hand.id, 1)
	hand.mut.Lock()
	hand.calls[id] = callback
	hand.mut.Unlock()
	return id
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
func (hand *handler) reset() {
	hand.mut.Lock()
	hand.calls = map[uint64]chan []byte{}
	hand.mut.Unlock()
}
