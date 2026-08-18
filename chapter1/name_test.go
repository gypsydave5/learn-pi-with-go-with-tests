package pi

import (
	"testing"
	"testing/synctest"
)

func TestOneCommunication(t *testing.T) {
	a := make(Name)
	w := make(Name)

	go func() { a <- w }()

	got := <-a

	if got != w {
		t.Errorf("got %v, want %v", got, w)
	}
}

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

func TestTwoInputsOneOutput(t *testing.T) {
	x := make(Name)
	y := make(Name)
	a := make(Name)
	b := make(Name)

	go Send(x, y)
	go Recv(x, a)
	go Recv(x, b)

	select {
	case <-a:
		t.Log("the second process won")
	case <-b:
		t.Log("the third process won")
	}
}

func TestReceiveWithNoSenderIsStuck_Bad(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {

		x := make(Name)
		y := make(Name)
		out := make(Name)

		go Recv(x, out)

		synctest.Wait()

		// if we get here we've normalized - either all processes have ended (reached 0) or they're blocked on send/recieve
		select {
		case <-out:
			t.Fatal("Recv produced something with no sender")
		default:
		}

		x <- y // what we were waiting for

		// and now I expect to be unblocked...
		if got := <-out; got != y {
			t.Errorf("got %v, wanted %v", got, y)
		}

	})
}
