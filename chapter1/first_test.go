package pi

import "testing"

func TestOneCommunication(t *testing.T) {
	a := make(Name)
	w := make(Name)

	go func() { a <- w }()

	got := <-a

	if got != w {
		t.Errorf("got %v, want %v", got, w)
	}
}
