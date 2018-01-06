package bi

import (
	"testing"
)

func Benchmark_Pool1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v := Pool.pkg.Get()
		Pool.pkg.Put(v)
	}
}
func Benchmark_Pool2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v := Pool.Timer.Get(1)
		Pool.Timer.Put(v)
	}
}
