package bi

import (
	"testing"
)

func Benchmark_Pool1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v := payloadPool.Get()
		payloadPool.Put(v)
	}
}
func Benchmark_Pool2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v := timerPool.Get(1)
		timerPool.Put(v)
	}
}
