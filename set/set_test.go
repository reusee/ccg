package set

import "testing"

func TestAll(t *testing.T) {
	s := New()
	s.Add(1)
	if !s.In(1) {
		t.Fail()
	}
	s.Del(1)
	if s.In(1) {
		t.Fail()
	}
	s.Del(1)
}
