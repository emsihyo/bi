package bi

import (
	"sync"
	"sync/atomic"
)

//handler handler
type handler struct {
	id    uint32
	calls map[uint32]chan []byte
	mut   sync.RWMutex
}

func newHandler() *handler {
	return &handler{calls: map[uint32]chan []byte{}}
}

func (hand *handler) addCall(callback chan []byte) uint32 {
	id := atomic.AddUint32(&hand.id, 1)
	hand.mut.Lock()
	hand.calls[id] = callback
	hand.mut.Unlock()
	return id
}

func (hand *handler) removeCall(id uint32) {
	hand.mut.Lock()
	delete(hand.calls, id)
	hand.mut.Unlock()
}

func (hand *handler) getCall(id uint32) chan []byte {
	hand.mut.RLock()
	defer hand.mut.RUnlock()
	return hand.calls[id]
}

func (hand *handler) reset() {
	hand.mut.Lock()
	hand.calls = map[uint32]chan []byte{}
	hand.mut.Unlock()
}
