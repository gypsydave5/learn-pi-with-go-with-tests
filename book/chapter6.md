# Chapter 6: Again, and again, and again

## What we're covering

`!`. Replication. How to write a process that doesn't die after one go.

It's the last piece of the grammar bar one, and it's the one that turns a term into something you could plausibly call a program.

## Setup

```
mkdir again && cd again
```

`Send` and `Recv`.

---

## Everything we've built is single-use

Have a look back. `Send` sends one name and stops. `Recv` receives one and stops. `Gate` hands out one private name, forwards one message, and stops.

Every process in this book so far has done its one thing and reached `0`.

Which is fine for a book, and useless for anything else. Servers don't stop. Workers don't stop. The whole point of a system is that it's still there tomorrow.

So we need a way to say "and again", and it's this:

```
!P
```

Read it "bang P", and what it means is: as many copies of `P` as anybody could want, running side by side.

The law:

```
!P  ≡  P | !P
```

Which is either obvious or alarming depending on how long you look at it. A replication is *structurally congruent to* one copy of itself running alongside a replication. Peel a copy off and you've still got the whole thing.

That isn't a reduction, mind. There's no `→` there. It's a `≡`, which means it's not something that *happens* - it's two ways of writing the same term. `!P` doesn't unfold over time. It was always already unfolded, as many times as you like.

Which is why `!P` never reaches `0`. It can't. Every time you peel off a copy there's another `!P` behind it, and there always will be.

---

## The useful shape

In practice you almost never want `!P` for arbitrary `P`. What you want is this:

```
!x(z).P
```

**Replicated input.** Read it: any number of copies of "receive on `x`, then `P`".

Think about what that gives you. A copy is sitting on `x` waiting. Something arrives, that copy takes it and goes off to do `P` - and behind it there's another copy, already waiting on `x`, ready for the next one.

That's a server. That is exactly and precisely a server, and it's why the literature usually restricts itself to replicated input rather than general `!P`: it's the shape you actually want, and it has its own law:

```
!x(z).P  |  x̄⟨y⟩.Q   →   P{y/z}  |  !x(z).P  |  Q
```

One copy peels off and handles the message. The replication is still there. Nothing was consumed.

---

## In Go, obviously

```go
// !x(z).ȳ⟨z⟩.0
func Server(x Name, y Name) {
	for {
		z := <-x
		go Send(y, z)
	}
}
```

A `for` loop round a receive, and the work spawned off with a `go`.

That `go` is not decoration. Look at what it's doing: it makes the handling of the message happen *alongside* the loop rather than in it, so the loop gets straight back to `<-x`. That's the "another copy already waiting" bit, and it's the whole difference between this and the thing we're about to compare it with.

---

## Not just servers

Before we move on - replicated input is the shape you'll use most, but it isn't the only interesting one. Have a look at this:

```
!ȳ⟨n⟩.0
```

No input guarding it. Just an output, replicated. What's that for?

Pair it with somebody who wants to read:

```
!ȳ⟨n⟩.0 | y(z).P   →   !ȳ⟨n⟩.0 | P{n/z}
```

The reader got `n`. And the replication is exactly as it was. Read it again, get `n` again. Forever, by as many processes as fancy it, and none of them are competing.

Which is quietly a big deal, because everything else in this calculus is *consumed*. Every communication we've written so far used something up - chapter 1's race had two inputs fighting over one message precisely because there was only the one. `!` is how you make something that doesn't run out.

So `y` here is a name that always yields `n`. A constant. An immutable binding. Not a channel you take turns on - a fact that anybody can look up.

```go
func Constant(y Name, n Name) {
	for {
		y <- n
	}
}
```

Keep that shape in your head, because chapter 8 lives on it. Building numbers out of processes only works if a number can be used more than once, and a number you could only read once would be no use to anybody. `!` is what turns a message into something that stays true.[^diverge]

[^diverge]: One warning about unguarded replication. If the thing you're replicating can reduce all by itself, `!` will happily do it forever and your term will spin without anybody asking it to. Guarding with an input - or, as here, with an output that needs a partner - keeps it honest. This is the sort of thing that makes people prefer `!x(z).P` and never think about it again.

---

## The other way to go round again

There's a second way to make a process repeat, and it looks equivalent, and it isn't.

Give the process a name and let it mention itself:

```
A ≜ x(z).ȳ⟨z⟩.A
```

That `≜` means "is defined as". It's a **named process definition**, and it's the first time we've used one - back in chapter 1 I mentioned they existed only to insist that `P` and `Q` weren't examples of them. Here's the real thing. `A` is a name for a process, it appears inside its own body, and that's recursion.

Read it: receive on `x`, send on `y`, and then be `A` again.

```go
// A ≜ x(z).ȳ⟨z⟩.A
func Worker(x Name, y Name) {
	z := <-x
	Send(y, z)
	Worker(x, y)
}
```

Now. Is that the same as `Server`?

---

## No. And here's the test that proves it.

Look at where the `A` is in that term. It's behind two prefixes. To get back round to the input on `x`, the process has to complete the input *and* the output first. It cannot start a second job until the first one is entirely finished.

That's **one at a time**. A worker.

Whereas `!x(z).ȳ⟨z⟩.0` has a fresh copy waiting on `x` immediately, whether or not the previous one has managed to deliver anything.

That's **unboundedly many at once**. A server.

Here's a test that can tell them apart:

```go
func TestAcceptsWhileBusy(t *testing.T) {
	x := make(Name)
	y := make(Name)
	first := make(Name)
	second := make(Name)

	go Server(x, y)

	x <- first  // in it goes
	x <- second // and another, while nobody has collected the first

	got1 := <-y
	got2 := <-y

	if got1 == got2 {
		t.Fatal("got the same name twice")
	}
}
```

Nothing is reading `y` when the second request goes in. So the only way that second `x <- second` can complete is if something is sitting on `x` ready to take it - which means the process came back round *before* finishing the first job.

`Server` passes. Swap in `Worker` and the test hangs, because `Worker` is still stuck in `Send(y, first)` waiting for somebody to collect, and nothing is listening on `x` at all.

Two terms. Both repeat forever. Completely different systems.

---

## They are interdefinable, and that's the trap

Before you conclude that recursion is simply worse: it isn't. You can write a concurrent one, you just have to say so:

```
A ≜ x(z).( ȳ⟨z⟩.0 | A )
```

Now `A` is behind *one* prefix, and after the input the output goes off in parallel while `A` comes straight back round. That behaves like the replicated version.

```go
func Worker(x Name, y Name) {
	z := <-x
	go Send(y, z)
	go Worker(x, y)
}
```

So recursion and replication can each express the other. The trap is that the *obvious* translation between them silently changes the concurrency, and it changes it in the direction that looks fine in testing and falls over under load. One at a time is perfectly correct. It's just not what you drew on the whiteboard.

### They're still not the same term

Worth being precise here, because you may have noticed something.

`!x(z).P` is congruent to `x(z).P | x(z).P | x(z).P | …` - unboundedly many inputs on `x`, all being offered *right now*. Whereas `A ≜ x(z).( ȳ⟨z⟩.0 | A )` offers exactly one, and only makes the next one after that one has fired. Acceptance is parallel in the first and serialised in the second.

Different terms. Same process, though - and the reason is that reduction happens one step at a time. There is no `→` in this calculus that consumes two outputs at once, so "a thousand inputs available at this instant" and "one input, immediately replaced" produce exactly the same sequence of steps. Nothing can tell them apart, because there is nothing to tell them apart *with*.

Hold onto that, because it's the first time in this book that two visibly different terms have turned out to be the same thing, and working out when that's true is what chapter 10 is about.

One caveat: this only works because the replication is guarded by an input. `!ȳ⟨n⟩.0` will merrily emit output after output with nothing prompting it, and no recursion sitting behind a prefix is going to match that. Which is another reason the literature mostly sticks to `!x(z).P`.

---

## Two Go problems, one of them fatal

**The stack.** Go doesn't do tail-call optimisation. `Worker` calling `Worker` calling `Worker` grows the stack forever, and eventually the program dies. It'll survive every test you write and then fall over after a fortnight in production.

Which is a happy accident, because it pushes you towards `for` - and a `for` loop doesn't nest, doesn't grow, and doesn't pretend the second iteration is somehow inside the first.

Be clear about what you've written when you do, mind. A `for` loop round a receive offers *one* input at a time, then comes back for the next. That's `A`, the recursion. It is not literally `!x(z).P`, which offers all of them at once. We get away with using it because of what we just established - they're the same process, for input-guarded shapes, because reduction happens one step at a time. Go's channels serialise handoff for the same reason, so the runtime and the calculus agree.

**The bubble.** Try putting `TestAcceptsWhileBusy` inside `synctest.Test`.

```
panic: deadlock
```

Of course it does. Chapter 1: every process in a bubble has to be able to reach `0`. And `!P` *by definition never reaches `0`* - that's what replication is.

So for the first time we can't use the bubble, and it isn't a limitation of the tool. It's the tool correctly refusing to certify a system that doesn't finish. `synctest` is for terms that terminate, and we've just left that world.

---

## Which means we've started leaking

That `Server` goroutine is still running when the test ends. So is whatever `go Send` we spawned that nobody collected from.

In Go we have a name for that - a **goroutine leak** - and a tool:

```
go get go.uber.org/goleak
```

```go
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
```

Run the tests now and `goleak` will tell you exactly which goroutines outlived them.

And here's the thing worth sitting with: **it's right, and so are we.** The goroutines really did leak. The process really is still running. That's not a bug in `Server`, it's what `!x(z).P` *means*, and a server that tidied itself away after one request would be a broken server.

What a real Go program does is add a way to stop - a done channel, a context, some second input that breaks the loop. Which is a perfectly sensible thing to do and is *no longer replication*. You've written something else: a process that repeats until told not to. The calculus can express that too, but it needs `+`, and that's the next chapter.

---

## Exercises

1. **(paper)** Use `!P ≡ P | !P` to show that `!P | !P ≡ !P`. Two copies of a server are one server. Does that surprise you?

2. **(paper)** Reduce `!x(z).ȳ⟨z⟩.0 | x̄⟨a⟩.0 | x̄⟨b⟩.0` until nothing further can happen. How many copies of the replication did you have to peel off, and how did you decide when to stop?

3. **(paper)** Write out `A ≜ x(z).ȳ⟨z⟩.A` unfolded three times. Now do the same for `A ≜ x(z).( ȳ⟨z⟩.0 | A )`. Put the two side by side and point at where they differ.

4. **(Go)** Swap `Worker` for `Server` in `TestAcceptsWhileBusy` and watch it hang. Now use `go test -timeout 5s` so it fails rather than sulks, and read the stack trace. Which goroutine is stuck, and on what?

5. **(Go)** Write `Server` with the `go` removed - `y <- z` directly in the loop. Which of the two terms have you now written? Prove it with the test.

6. **(Go)** Add `goleak` and run the whole chapter's tests. Write down every leak it finds and, for each one, say whether it's a bug or a faithful implementation of a `!`.

---

## Recap

- `!P ≡ P | !P`. As many copies as you like, and peeling one off leaves the whole thing behind.
- It's a `≡`, not a `→`. Replication doesn't unfold over time; it was always unfolded.
- `!P` never reaches `0`. That's not a defect, it's the definition.
- **Replicated input `!x(z).P` is the shape you actually want.** It's a server.
- `A ≜ …` is a named process definition: recursion.
- **Recursion is not replication.** `A ≜ x(z).ȳ⟨z⟩.A` is one at a time. `!x(z).ȳ⟨z⟩.0` is many at once. Both repeat forever; they are not the same system.
- They're interdefinable - `A ≜ x(z).( ȳ⟨z⟩.0 | A )` - but the obvious translation quietly changes the concurrency.
- No TCO in Go, so recursion grows the stack. Use `for` - which is literally the recursion, not the replication, and is fine because for input-guarded shapes those are the same process.
- The bubble can't hold a replication, because bubbles require termination and `!` refuses it. `goleak` instead.

## Next

Chapter 7: `+`. Choice. A process that can offer two things and do only one of them - which is how you write a server that can be told to stop, and how the observer finally becomes something we can write down as a term rather than waving at.
