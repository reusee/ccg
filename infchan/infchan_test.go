package infchan

import "testing"

func TestAll(t *testing.T) {
	in, out, kill := New()
	defer close(kill)
	count := 512
	for i := 0; i < count; i++ {
		in <- i
	}
	for i := 0; i < count; i++ {
		v := <-out
		if v != i {
			t.Fatalf("expected %d got %d", i, v)
		}
	}
}

func BenchmarkSendRecv(b *testing.B) {
	in, out, kill := New()
	defer close(kill)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		in <- 42
		<-out
	}
}
