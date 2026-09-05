# Chapter 7: One or the other

## What we're covering

`+`. Choice. A process that offers two things and does exactly one of them.

It's the last piece of the grammar, it's how you write a server that can be told to stop, and it's how we finally get to write down the observer that's been hanging about offstage since chapter 1.

## Setup

```
mkdir one-or-the-other && cd one-or-the-other
```

`Send` and `Recv`, and you'll want `goleak` from last chapter.

---

## We've made something we can't turn off

Chapter 6 left us in an awkward spot.

`!x(z).P` runs forever. That's not a bug, it's the definition - a server that packed up after one request would be no use. But it means the bubble won't take it, `goleak` quite rightly complains, and there is no way in anything we've written so far to say "right, that's enough".

And you can't fix it with what we've got. Every process we can write does one thing: it's got a prefix on the front, and it waits for that one prefix, and nothing else in the world will move it. `x(z).P` waits on `x`. It doesn't have opinions about anything else.

What we need is a process that's waiting on *two* things and will take whichever turns up.

---

## Offer two, take one

```
P + Q
```

Both are on offer. One of them happens. The other is discarded.

The rule is about as small as a rule gets:

```
     P → P'
  ────────────
   P + Q → P'
```

If `P` can move, then `P + Q` can move the same way, and `Q` is gone. And by symmetry, likewise if `Q` moves.

Note what decides it: **nothing in the term does.** There's no condition, no predicate, no test. Whichever side manages to react first, wins - which means the choice is made by *the environment*, by whatever process turns up offering a matching action. The term offers; the world picks.

Spelled out for a communication, so you can see both discards happening at once:

```
( x̄⟨y⟩.P + P' )  |  ( x(z).Q + Q' )   →   P | Q{y/z}
```

`P'` and `Q'` were both real possibilities a moment ago. They're now not just unselected but *unreachable* - there is no path back to them, ever.

### The laws

`+` is a monoid too, and its unit is the same `0`:

```
P + Q       ≡  Q + P
(P + Q) + R ≡  P + (Q + R)
P + 0       ≡  P
```

That last one's worth a second. `0` offers nothing, so a branch of `0` can never be the one that's chosen, so having it there changes nothing. A choice you can't take isn't a choice.

### One restriction

We only ever write sums where **every branch starts with a prefix**:

```
x(z).P + y(w).Q       fine
x̄⟨n⟩.P + y(w).Q       fine
P + Q                 don't
```

These are **guarded sums**, and the guard is the prefix on the front. Unguarded sums cause genuine trouble once you mix them with replication and structural congruence, and they buy nothing you'd miss. Every `+` in this book has a prefix on each side.

---

## This is not `|`

Worth being very clear, because the two look adjacent and are not remotely the same.

```
P | Q     both happen
P + Q     one happens, and the other is destroyed
```

`|` adds possibilities. `+` spends them. After a `|` you've got two processes; after a `+` you've got one, and the other never existed as far as the rest of time is concerned.

That destruction is the whole content of the operator, and it's going to matter enormously in three chapters' time.

---

## In Go, and you've been using it since chapter 1

```go
select {
case z := <-x:
	// ...
case <-d:
	// ...
}
```

`select` is `+`. It offers several communications, takes whichever becomes available, and abandons the rest.

I've been using it in tests since chapter 1 and cheerfully telling you it was chapter 7's problem. Here we are.

---

## Your turn: a server you can stop

Here's the term, and it's the first one in this book that's a plausible piece of software:

```
A ≜ x(z).( ȳ⟨z⟩.0 | A )  +  d(q).0
```

Read it: offer two things. Either receive a job on `x` - in which case handle it, and go round again - or receive anything at all on `d`, in which case stop.

That's the recursion shape from chapter 6 with a second branch bolted on. And the second branch is `0`, which is the point: it's the branch that doesn't go round again.

```go
// Serve handles jobs arriving on x by forwarding them to y, until
// something arrives on d, at which point it stops.
//
//	A ≜ x(z).( ȳ⟨z⟩.0 | A ) + d(q).0
func Serve(x Name, y Name, d Name)
```

And here's the test, in which something rather nice happens:

```go
func TestServeStopsWhenTold(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		x := make(Name)
		y := make(Name)
		d := make(Name)
		n := make(Name)

		go Serve(x, y, d)

		x <- n
		if got := <-y; got != n {
			t.Errorf("got %v, want %v", got, n)
		}

		d <- n // any name will do - it's the arrival that matters
	})
}
```

**The bubble is back.**

Chapter 6 couldn't use `synctest` at all, because `!P` never reaches `0` and a bubble insists that everything does. Give the server a way to stop and it can reach `0`, and the bubble will take it again.

Which makes this test unusual, so look at what's actually being asserted. There's no `t.Fatal` about stopping. The assertion is the *absence of a panic*: if `Serve` ignored `d` and carried on looping, the send on `d` would block forever, nothing would ever finish, and `synctest.Test` would fail with a deadlock. The test passing is the proof.

Add `goleak` while you're here. It'll be happy too, for the first time since chapter 6.

---

## Go promises us something it didn't before

Try this:

```go
func TestSelectIsRandom(t *testing.T) {
	for i := 0; i < 20; i++ {
		a := make(Name)
		b := make(Name)
		n := make(Name)

		go func(a Name, n Name) { a <- n }(a, n)
		go func(b Name, n Name) { b <- n }(b, n)

		select {
		case <-a:
			t.Log("a")
		case <-b:
			t.Log("b")
		}
	}
}
```

You'll see both. No `-race` required.

Which is a change from chapter 1, and a meaningful one. Back there the two possible outcomes were both legal, and the scheduler picked the same one every single time, because nothing in the language spec had anything to say about which goroutine gets woken. We needed the race detector to jog it.

`select` is different. When more than one case is ready, Go **guarantees** it picks uniformly at random. It's in the spec. That's a language-level promise that `+` really is a choice and not a disguised preference for whatever you wrote first.

So this is the one place where Go's runtime will show you the non-determinism the calculus has been talking about since chapter 1, without being poked.

---

## The observer, at last

Right back in chapter 1 there was a loose end. We watched a race between three processes, and I pointed out that the *test* was doing something suspiciously process-like - taking an input - while not being in the term at all. I said it was chapter 7's problem.

Here it is:

```
a(q).0 + b(q).0
```

Offer an input on `a`, offer an input on `b`, take whichever arrives, abandon the other, stop.

That's the observer. That's what your `select` in a test *is*, written down as a term. We can finally put the whole thing on paper with nothing left over.

Which is more useful than tidiness. If the observer is a process, then observing is just interacting, and "what can we find out about this system" becomes "what processes can we compose it with and what happens". Chapter 10 is entirely about that question.

---

## A puzzle to take away

Two terms:

```
x(z).( ȳ⟨n⟩.0 + w̄⟨n⟩.0 )
x(z).ȳ⟨n⟩.0  +  x(z).w̄⟨n⟩.0
```

The first: receive on `x`, and *then* offer a choice of two outputs.

The second: offer a choice of two whole processes, both of which start by receiving on `x`.

Now. Run either one and write down what happened. You received on `x`, then something came out on `y` or on `w`. Both terms can produce both of those histories, and there is no sequence of events that one can produce and the other can't.

So are they the same process?

Think about it from the point of view of somebody holding a name. In the first term, after you've sent on `x`, the process still has both options open - you can take `y` or you can take `w`, your call. In the second, sending on `x` already committed it. One of those branches was destroyed before you got a look in, and you don't get to choose which.

Same histories. Different processes. And nothing we currently possess can express the difference.

Sit with that. It's chapter 10.

---

## Exercises

1. **(paper)** Show that `x(z).P + x(z).P ≡ x(z).P`. Is `+` idempotent in general? Try `x(z).P + y(w).Q + x(z).P`.

2. **(paper)** Reduce `( x̄⟨n⟩.0 + ȳ⟨n⟩.0 ) | x(z).w̄⟨z⟩.0` as far as it goes. What happened to the branch that wasn't taken, and at which step exactly did it stop being available?

3. **(paper)** Write the term for a process that offers a job channel, a stop channel, and a "tell me how you're doing" channel that replies and then carries on serving. Three branches, and only one of them stops.

4. **(Go)** Implement exercise 3's third branch in `Serve`. Test that the server is still alive after being asked how it's doing.

5. **(Go)** Take the `select` out of `Serve` and write the two branches as a `|` instead - handle jobs in one goroutine, watch `d` in another. Write a test that distinguishes it from the `select` version. What can go wrong that couldn't before?

6. **(Go)** In `TestSelectIsRandom`, make one of the two sends happen a fraction later than the other. Is it still random? What does that tell you about what Go is actually promising?

---

## Recap

- `P + Q` offers both and does one. The other is destroyed, not deferred.
- Nothing in the term decides which. The environment picks, by turning up with a matching action.
- `+` is commutative and associative with `0` as unit - a branch of `0` is a choice you can never take.
- **Guarded sums only.** Every branch starts with a prefix.
- `+` is not `|`. `|` adds possibilities, `+` spends them.
- `select` is `+`, and Go *guarantees* uniform random selection among ready cases - unlike the scheduler in chapter 1, which promised nothing.
- Choice is how a server gets a stop button, which is how it reaches `0`, which is how the bubble takes it back.
- The observer is `a(q).0 + b(q).0`. Observing is just interacting.
- Two terms with identical histories can still be different processes. That's the whole of chapter 10.

## Next

Chapter 8, and the fun begins. We've got the entire grammar now - names, input, output, `0`, `|`, `ν`, `!`, `+` - and no data whatsoever. So we build some. Booleans first, then the natural numbers, out of nothing but processes talking to each other.
