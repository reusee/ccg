package infchan

import "container/list"

type T interface{}

func New() (chan<- T, <-chan T, chan struct{}) {
	store := list.New()
	in := make(chan T)
	out := make(chan T)
	kill := make(chan struct{})

	go func() {
		defer func() {
			close(in)
			close(out)
		}()
		for {
			if store.Len() > 0 {
				e := store.Front().Value.(T)
				select {
				case out <- e:
					store.Remove(store.Front())
				case v := <-in:
					store.PushBack(v)
				case <-kill:
					return
				}
			} else {
				select {
				case v := <-in:
					store.PushBack(v)
				case <-kill:
					return
				}
			}
		}
	}()

	return in, out, kill
}
