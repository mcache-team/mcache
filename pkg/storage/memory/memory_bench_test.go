package memory

import (
	"fmt"
	"testing"
	"time"

	"github.com/mcache-team/mcache/pkg/apis/v1/item"
)

// newTestMemory returns a fresh Memory instance for benchmarks.
func newTestMemory() *Memory {
	return NewStorage().(*Memory)
}

// ---- Insert ----

func BenchmarkInsert(b *testing.B) {
	m := newTestMemory()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Insert(fmt.Sprintf("key-%d", i), "value")
	}
}

func BenchmarkInsertWithTTL(b *testing.B) {
	m := newTestMemory()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Insert(fmt.Sprintf("key-%d", i), "value", item.WithTTL(time.Minute))
	}
}

func BenchmarkInsertParallel(b *testing.B) {
	m := newTestMemory()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = m.Insert(fmt.Sprintf("key-%d-%d", b.N, i), "value")
			i++
		}
	})
}

// ---- Get ----

func BenchmarkGet(b *testing.B) {
	m := newTestMemory()
	for i := 0; i < 1000; i++ {
		_ = m.Insert(fmt.Sprintf("key-%d", i), "value")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.GetOne(fmt.Sprintf("key-%d", i%1000))
	}
}

func BenchmarkGetParallel(b *testing.B) {
	m := newTestMemory()
	for i := 0; i < 1000; i++ {
		_ = m.Insert(fmt.Sprintf("key-%d", i), "value")
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_, _ = m.GetOne(fmt.Sprintf("key-%d", i%1000))
			i++
		}
	})
}

// ---- Update ----

func BenchmarkUpdate(b *testing.B) {
	m := newTestMemory()
	for i := 0; i < 1000; i++ {
		_ = m.Insert(fmt.Sprintf("key-%d", i), "value")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Update(fmt.Sprintf("key-%d", i%1000), "new-value")
	}
}

// ---- Delete ----

func BenchmarkDelete(b *testing.B) {
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		m := newTestMemory()
		_ = m.Insert("key", "value")
		b.StartTimer()
		_, _ = m.Delete("key")
		b.StopTimer()
	}
}

// ---- ListPrefix ----

func BenchmarkListPrefix_100(b *testing.B) {
	benchListPrefix(b, 100)
}

func BenchmarkListPrefix_1000(b *testing.B) {
	benchListPrefix(b, 1000)
}

func benchListPrefix(b *testing.B, n int) {
	b.Helper()
	m := newTestMemory()
	for i := 0; i < n; i++ {
		_ = m.Insert(fmt.Sprintf("root/child-%d", i), "value")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.ListPrefix("root/")
	}
}

// ---- Mixed read/write ----

func BenchmarkMixedReadWrite(b *testing.B) {
	m := newTestMemory()
	for i := 0; i < 1000; i++ {
		_ = m.Insert(fmt.Sprintf("key-%d", i), "value")
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%4 == 0 {
				_ = m.Insert(fmt.Sprintf("new-%d", i), "value")
			} else {
				_, _ = m.GetOne(fmt.Sprintf("key-%d", i%1000))
			}
			i++
		}
	})
}
