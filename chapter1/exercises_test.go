package pi

import "testing"

func TestExerciseOne(t *testing.T) {
	payload := make(Name)
	a := make(Name)
	b := make(Name)
	c := make(Name)

	go Send(a, payload)
	go Recv(a, b)
	go Recv(b, c)

	out := <-c

	if out != payload {
		t.Errorf("Got %v, wanted %v", out, payload)
	}
}

// Exercise Two - it freezes!

// Exercise Three - still non-deterministic, as still two valid reductions, with the sends going to different recievers and vice versa

// Exercise Four - meh
