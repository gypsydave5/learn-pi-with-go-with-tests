package pi

import "testing"

func TestTwoProcessesCommunicate(t *testing.T) {
	x := make(Name)
	y := make(Name)
	out := make(Name)

	go Send(x, y)
	go Recv(x, out)

	if got := <-out; got != y {
		t.Errorf("got %v, want %v", got, y)
	}
}
