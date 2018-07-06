package bi

import (
	"testing"
)

func Benchmark_PoolPayload(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v := payloadPool.Get()
		payloadPool.Put(v)
	}
}
func Benchmark_PoolTimer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v := timerPool.Get(1)
		timerPool.Put(v)
	}
}
