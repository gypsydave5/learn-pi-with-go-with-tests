# Chapter 1: A name is a channel is a name

**Draft 4.** You need Go 1.25 or later.

## What we're covering

The single data type of the π-calculus, and the single thing that can happen in it.

That's it. That's the chapter. If it feels like I'm going slowly it's because everything else in this book gets built out of these two ideas and nothing else.

## Setup

```
mkdir pi && cd pi
go mod init pi
```

Check your `go.mod` says `go 1.25` or higher. We'll want `testing/synctest` before the end.

Each chapter gets its own directory. This one:

```
mkdir name-is-a-channel && cd name-is-a-channel
```

with two files in it, which we'll build up as we go:

```
name-is-a-channel/
  name.go
  name_test.go
```

Every chapter is self-contained. Chapter 5 will redefine `Name` rather than importing it from here - a bit of duplication, in exchange for being able to drop in anywhere.

---

## Why bother?

As you probably know, the λ-calculus is built out of three things: variables, functions, and applying one to the other. That's the entire grammar. Everything else - booleans, numbers, pairs, recursion - is something you *build* out of functions, rather than something the calculus hands you.

It's a reductive move, and a very good one. **Everything is a function.** If you've ever written `λf.λx.f (f x)` on a whiteboard and felt pleased with yourself, that's what you were feeling pleased about: numbers turning out to be functions all along. And if you haven't done this, don't worry, we're not going to rely on whether you've done lots of lambda calculus before. But your eyes may glaze over in the next few paragraphs.

So the λ-calculus also comes with a theorem attached, we need to get rid of it early on, because its failing shows us why we need the pi-calculus.

The theorem is **Church-Rosser**, and it says: if a term can be reduced in two different ways, the two results can always be brought back together by reducing further. Reduction is *confluent*. Concretely, every term has at most one normal form, and the order you reduce in doesn't change where you end up.

That theorem is most of why the λ-calculus is pleasant to work with. "What does this reduce to?" has one answer. Evaluation strategies become an optimisation problem rather than a correctness one. The whole question of what a program *means* rests on it being true.

Now back to reality. Two goroutines and one channel:

```go
a := make(chan int)
go func() { a <- 1 }()
go func() { fmt.Println("A", <-a) }()
go func() { fmt.Println("B", <-a) }()
```

One message. Two receivers. Which one gets it?

Neither. Both. It depends. Whatever. There isn't an answer, and there *shouldn't* be. It's not that we've been sloppy and failed to specify something - two things racing for the same resource is the correct description of what's happening, and so if a formalism insists on a single and unique answer to this, then so much the worse for the formalism.

So Church-Rosser is gone. And normal forms go with it, and so does "what does this reduce to", because the honest answer is "errr lol idk, one of these two?".

Here's a second thing to keep you awake at night. What is the *value* of a web server?

It hasn't got one. It isn't computing a result. If it ever does reduce all the way to a normal form and stop, someone (hopefully not me) is going to get woken up. The whole reason it exists is the interacting, the back and forth, and treating interaction as the tedious plumbing on the way to a return value has it precisely backwards.

You can't fix this by encoding, either. The temptation is to say: fine, I'll model concurrency *inside* the λ-calculus - continuations, a state monad, an interleaving semantics that enumerates the possible schedules. All of these work, and people have built serious things on them. But what you end up with is a description of concurrency written in a language whose primitive is the function call, so the thing you actually care about - two independent things meeting - is never primitive. It's always an encoding, and you're always translating. As though you'd decided to notate a symphony by describing the precise coordinates that each player should be putting their fingers. You can do it, but it's just not a good fit.

So Milner flipped the story on its head, and chose to be reductive about something else. Not *everything is a function*, but **everything is an interaction**. Same wildly liberating reductionism as Church, just a different primitive.

And once you've said that, the rest of the book is forced: if interaction is the primitive, then data has to be built out of interactions, in much the same way that Church built data out of functions. Exciting!

### Where this came from[^ccs]

Milner had already built CCS - the Calculus of Communicating Systems - in the late seventies, and it handled concurrency perfectly well. Processes had named ports; a process offering `a` could synchronise with one offering the complementary `ā`; you could compose processes in parallel, restrict which port names were visible from outside, and offer a choice between behaviours. Bisimulation, which is how we'll eventually decide whether two processes are the same, comes from CCS rather than the π-calculus.

What CCS couldn't do was change its mind about who talks to whom. The port names in a CCS term are fixed when you write the term down. You can hide them, you can rename them, but the *shape* of the system - the graph of which process can reach which other process - is static for all time.

The π-calculus is what happens when you allow one more thing: port names can be sent as messages. That's it, that's the entire difference, and it turns out to change everything. Milner's own summary of the resulting calculus is that <cite index="6-1">the most prominent ingredient is the notion of a name</cite> - which by the end of this book I hope reads as a startling thing to have said, because it means there isn't anything else in there.

[^ccs]: CCS is worth a look on its own account if you like this stuff - it's simpler than the π-calculus and the bisimulation material is easier to meet there first. Milner's *A Calculus of Communicating Systems* (1980) is the original; *Communication and Concurrency* (1989) is the more approachable rewrite. Note that both have titles confusingly close to the π-calculus book, *Communicating and Mobile Systems* (1999).

---

## So what do we actually need?

Let's be as stingy as Church was.

Two processes have to be able to meet and hand something over. So, minimally, we need something to communicate **over**, and something to communicate.

You already know what the first one looks like:

```go
done := make(chan struct{})
```

That's a Go channel carrying nothing at all. The payload is the empty struct, zero bytes, and it's a perfectly useful thing - all the information is in the *fact that a send happened*. The event is the message.

Then there's the ordinary case:

```go
results := make(chan int)
lines := make(chan string)
jobs := make(chan Job)
```

Channels carrying values. Fine, familiar, and each one requires us to have decided what a value *is*. Which is the problem: whatever we pick, we've picked something arbitrary, and the calculus isn't minimal any more, it's just small.

But here's a thing you've certainly written:

```go
type Request struct {
	Data  string
	Reply chan Response
}

requests <- Request{Data: "hello", Reply: make(chan Response)}
```

The reply channel. Every Go programmer writes this within a fortnight of picking the language up, and it's more interesting than it looks: **you just sent a channel over a channel.** You handed a stranger the means of getting back to you. Before that send, nothing in the receiving goroutine could have referred to your reply channel - there was no expression it could have written. After it, there is.

You've met this shape before under other names. It's continuation-passing: rather than returning a value, you're handed somewhere to *put* the answer, and you put it there. It's also, at a grubbier scale, a webhook - here's a URL, call me back on it. The pattern keeps turning up because it's the only way to get an answer to someone whose address you didn't previously have.

Hold onto the continuation-passing one. In chapter 9 it stops being an analogy and becomes the actual mechanism by which the λ-calculus turns into π-calculus terms.

You've been doing the interesting part all along and calling it request/reply.

So: what if that's the *only* thing you can send?

```go
type Name chan Name
```

A name is a channel is a name. And a name is a channel that can send a name. And a channel is a name that can send a channel.

That's the type system. Not a simplification for chapter one, not a starting point we extend later - that's the whole thing, for the rest of the book.

No integers. No strings, no booleans, no functions, no structs. There are **names**, and a name is a channel that carries names. Go's recursive type definitions let us write it down literally, which is a piece of luck I've never entirely got over.

Your reaction to this should be that it obviously can't be enough. Hold onto that - it's the most productive feeling available to you right now. Chapter 8 builds the natural numbers out of it. Chapter 9 builds the λ-calculus out of it.

For now, take it as a constraint with teeth: **if you can't say something with names alone, that's the calculus telling you something.** It isn't Go being inadequate, and it isn't me withholding a feature until chapter 4.

Making one:

```go
s := make(Name)
```

Unbuffered - the rule we're about to meet has an output *and* an input in it, and an unbuffered channel is exactly that rendezvous.[^async]

[^async]: If you're wondering what happens with a buffered one: you get a different calculus. The asynchronous π-calculus is a real thing with its own literature, and it's just not this.

---

## The only thing that happens

Two operations and one rule joining them up.

**Output.** In the calculus this is `x̄y` - send the name `y` on the channel `x`:

```go
x <- y
```

**Input.** `x(z)` - receive on `x`, and call whatever arrives `z`:

```go
z := <-x
```

The overbar on the channel name is the output. It's the notation Milner used and it's what you'll meet in the literature, so we're going to live with it.

> **A note on single letters.** The overbar is a combining character, so it only sits over one letter - `c̄h` looks like a mistake rather than a channel called `ch`. This is why π-calculus papers use single-letter names for everything, and why we will too, in the calculus. In the Go, names can be as long as they need to be.

### Where does any of this happen?

Names, and the two things you can do with one. That isn't yet enough to write anything down, because both of those things need somebody to be *doing* them - and, much more importantly, need somebody *else* to be doing the other one.

That's the actual subject. Two independent things, running at the same time, meeting. Almost all the notation from here on exists to describe that, rather than to describe names.

So: a **process** is a thing that runs and interacts. That's the whole definition, and the vagueness is deliberate - we never say what a process is *made of*, only what it can do. We write them with capital letters, `P`, `Q`, `R`, the way we've been writing `x` and `y` for names.

Three pieces of notation and then we can state the rule.

**`0`** is the inert process. Finished. Does nothing, offers nothing, nobody can interact with it.

**The dot** is "and then". `x̄y.P` reads *output `y` on `x`, and then carry on as `P`*, and everything to the right of the dot stays frozen until that output has actually happened.

If "and then `P`" makes it look like we're *naming* the process `P`, that's the notation being unhelpful and your eyes are working fine. `P` is a variable - a **metavariable**, which is to say it belongs to the language we use to talk about terms, not to the terms themselves. It stands in for however much more process there is, and in a real term you fill it in:

```
x̄y.P            where P is w̄v.0
x̄y.w̄v.0         the same term, written out
```

There's no `P` left in the finished thing. It's the same trick as `λz.M` - the `M` isn't the function's name, it's the body, and it vanishes as soon as you write an actual body.[^defs]

[^defs]: Worth keeping this separate from something you may meet elsewhere. Some presentations *do* let you name a process, with definitions like `A(x) ≜ x(z).āz.0`, and then `A` genuinely is a name for a process and genuinely does appear inside terms. It's a convenience for writing recursive behaviour down without going mad. We won't need it - we get repetition from `!` in chapter 6 - and `P` and `Q` are never that.

This is the only sequencing in the language - there is no `;`. If you want two things in order, you chain prefixes: `x(z).y(w).P` inputs on `x`, and *then* on `y`, and the second one cannot possibly go first.

And look - chain a couple of prefixes onto a `0` and you've written a whole process:

```
x(z).ȳz.0
```

Take a name off `x`, pass it along on `y`, stop. An input, an output, and a full stop. No boilerplate, nothing declared, nothing hiding - that's a complete program in the language, and you could read it. Cute.

Notice that the prefix notation already assumed a process was coming. The dot's entire job is to say what happens next, so there had to be a next.

**Parallel composition**, `P | Q`, is P and Q running at the same time. Both live, neither waiting on the other, and either one free to interact with anything else that's about. It gets a chapter of its own, because "at the same time" has more in it than you'd think, but "at the same time" will do for now.

And now the rule:

```
x̄y.P  |  x(z).Q   →   P  |  Q{y/z}
```

Two processes, side by side. One is offering an output on `x`, the other an input on `x`. They meet, and the whole thing becomes something new: the two continuations, still side by side, with a substitution applied to the one that received.

> **`Q{y/z}`** is read "Q with y for z". Every occurrence of `z` in `Q` is replaced by `y`.
>
> It's easy to get backwards, and the notation is no help at all. The **new** name goes on top; the **old** one underneath. It's a fraction you're cancelling: `z` goes away, `y` arrives.
>
> If you're unsure, reason it out from the rule rather than trying to remember. The input `x(z).Q` chose the letter `z` as a placeholder, before it had any idea what would arrive. Then `y` arrived. Obviously it's `z` that has to go.
>
> ```
> x̄y.0  |  x(z).āz.0   →   0  |  āy.0
> ```
>
> The `z` became a `y`. Never the other way round.[^subst]

[^subst]: Some authors write `Q[y/z]`, and a few use an arrow, `Q[z ↦ y]`, which has the virtue of being impossible to misread. The `{y/z}` form is the most common and we'll stick with it. There's also a wrinkle I'm skating over: strictly it's every *free* occurrence of `z` that gets replaced, because `Q` might contain its own binder that happens to use the same letter, and that one is a different `z` entirely. Chapter 2.

Look at that substitution, because you've seen it before:

```
(λz.Q) y   →   Q{y/z}
```

Same rule. β-reduction is what happens when a function meets an argument, and juxtaposition on the page is what decides they've met. Communication is what happens when an output meets an input, and *naming the same channel* is what decides they've met.

The π-calculus takes β-reduction and splits it into two halves that have to go and find each other first. Everything difficult and everything interesting falls out of that one change.

---

## The first test

Let's watch one communication happen.

```go
package pi

import "testing"

type Name chan Name

func TestOneCommunication(t *testing.T) {
	x := make(Name) // the channel
	y := make(Name) // the name we'll send

	go func() { x <- y }() // x̄y

	got := <-x // x(got)

	if got != y {
		t.Errorf("got %v, want %v", got, y)
	}
}
```

Run it. It passes. Not much of a test, but it's the smallest true thing we can say.

Three things to notice.

**Names are compared by identity.** `got != y` asks "is this the same channel?", not "does it hold the same thing". Names have no contents. They *are* the thing.

**The test is one half of the communication.** There's only one `go` in there, because the test function itself plays the input side. We'll lean on this constantly: the test is the **observer**, standing outside the system, poking it and watching what comes back.

**Nothing is waiting on anything else.** The output and the input found each other because there was exactly one of each. That won't last.

---

## What a process is in Go

A process is a function that takes names and returns nothing.

```go
func SomeProcess(x Name, y Name) { ... }
```

Three rules.

**It only touches the names it was given.** No closures reaching out into the enclosing scope, no package-level state, no globals. If a process knows about a name, somebody handed it that name. That's the only way there is.

**No return values.** Not "usually not" - never. A process that wants to tell you something sends it on a name you gave it.

**`go` is parallel composition; a plain call is the dot.** Spawning is how you say `P | Q`. Calling is how you say "and then". A goroutine is a thread of causality, and `go` is the only way to get another one - everything reachable by plain calls from a single `go` is one process working through its prefixes in order.

```go
func Recv(x Name, out Name) {
	z := <-x    // x(z).
	Send(out, z) //   ōut z   ← a call: this IS the continuation
}

func Both(x Name, y Name) {
	go P(x)     // P | Q
	go Q(y)     //   ← go: this is parallel composition
}
```

A process blocked on a channel is a process sitting at a prefix, waiting to interact. Not running, not finished - waiting. That's a perfectly respectable state and it's most of the state space.

Now notice what the signature *doesn't* tell you:

```go
func SomeProcess(x Name, y Name)
```

Which of those does it read from? Which does it write to? Both? Neither? You can't tell, and neither can the compiler, because there's one type and it hasn't got a direction. `Send` and `Recv` are about to have identical signatures and opposite behaviour.

That should bother you. It's the gap that i/o types were invented to close, and session types after them, and we'll come back to it - but not for a while, because the untyped version has to make sense first.

---

## Your turn

Here's the test.

```go
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
```

It won't compile, because those two processes don't exist yet. Here's what it wants:

```go
// Send outputs payload on ch, then stops.
//
//	x̄y.0        where x is ch, and y is payload
func Send(ch Name, payload Name)

// Recv inputs on ch, binding whatever arrives as z,
// then outputs z on out.
//
//	x(z).ȳz.0   where x is ch, and y is out
func Recv(ch Name, out Name)
```

`0` is the inert process - the one that does nothing at all. In Go, that's the end of the function body.

Each one is a line or two. Go and do it, I'll wait.

---

## The race, in the flesh

Back at the start I claimed Church-Rosser was gone. Let's watch it go.

Here's the term we're going to write:

```
x̄y.0  |  x(z).āz.0  |  x(z).b̄z.0
```

Everything in there is something you've already met - the overbar, the dot, `0`, and `|` for two things running at the same time.[^par] So before you read on: say what it means, out loud, on your own. I'll wait again.

[^par]: Parallel composition gets a chapter to itself, because it turns out to have more to it than "and also". For now, "and also" is fine.

Now together.

There's a process that outputs the name `y` on channel `x`, and then stops. Running at the same time, a second process that takes a `z` off channel `x`, then outputs that `z` on channel `a`, then stops. And alongside both of those, a third process that takes a `z` off channel `x`, then outputs that `z` on channel `b`, then stops.

Three processes. One message. Two of them want it.

In Go:

```go
func TestTwoInputsOneOutput(t *testing.T) {
	x := make(Name)
	y := make(Name)
	a := make(Name)
	b := make(Name)

	go Send(x, y) // x̄y.0
	go Recv(x, a) // x(z).āz.0
	go Recv(x, b) // x(z).b̄z.0

	select {
	case <-a:
		t.Log("the second process won")
	case <-b:
		t.Log("the third process won")
	}
}
```

Run it with `go test -v -count=20` and read the log lines. You should see both.

### Reducing it, twice

Two log lines means two behaviours. Let's do it on paper and see the same thing.

The rule, once more:

```
x̄y.P  |  x(z).Q   →   P  |  Q{y/z}
```

**First way.** The output pairs with the second process. Matching against the rule, `P` is `0` and `Q` is `āz.0`, so the continuation becomes `āz.0{y/z}`, which is `āy.0`:

```
x̄y.0  |  x(z).āz.0  |  x(z).b̄z.0
   →     0  |  āy.0  |  x(z).b̄z.0
```

**Second way.** Same starting term. The output pairs with the third process instead. Now `Q` is `b̄z.0`, and the continuation becomes `b̄y.0`:

```
x̄y.0  |  x(z).āz.0  |  x(z).b̄z.0
   →     0  |  x(z).āz.0  |  b̄y.0
```

One term. Two reductions. Both perfectly legal, and nothing in the rules picks between them.

Now look at where we've ended up. In the first, there's an output waiting on `a` and the third process is stuck on `x` forever, because there is no output left on `x` to pair with. In the second, exactly the reverse.

Those two states are not the same state, and there's no sequence of reductions that turns either into the other. Let the observer consume the pending output and you're left with `x(z).b̄z.0` in one case and `x(z).āz.0` in the other. Still different. Still stuck. Nothing to be done.

Which is the `t.Log` line you just watched flip.

You might be wondering where *we* are in that description.

The term has three processes in it. The Go has four things running: those three, and the test, sitting on a `select`, waiting to see which of `a` or `b` delivers. That fourth one isn't in the term at all.

That's deliberate, and it's the observer again. Everything on paper is *the system*; the test stands outside it and pokes. But it's a bit of a cheat, because the test is plainly doing something a process does - taking an input - and if it's doing process things then it ought to be writable as a process.

It is. It's roughly `a(v).0 + b(v).0` - offer an input on `a` and an input on `b`, take whichever arrives, abandon the other. That `+` is choice, it's what `select` compiles to in our heads, and it's chapter 7. Until then the observer stays offstage, and when it comes onstage in chapter 10 it stops being a convenience and becomes the entire subject: what two processes *mean* turns out to be what an observer can tell apart.

(You may also have noticed I'm leaving `0`s lying about in the reductions rather than tidying them up. `0 | āy.0` is obviously just `āy.0`. Saying *why* you're allowed to cross it out is chapter 3. Cross it out anyway.)

### What the λ-calculus would have done

Take a term that also reduces two ways:

```
(λu.u) ((λv.v) w)
```

Reduce the outer application first:

```
(λu.u) ((λv.v) w)   →   (λv.v) w   →   w
```

Or the inner one first:

```
(λu.u) ((λv.v) w)   →   (λu.u) w   →   w
```

Same answer. That's not luck, it's Church-Rosser: whenever a λ-term forks, the branches can always be brought back together. Ordering is a question about efficiency, never about outcome.

Our term forks and stays forked. That's the difference, and it isn't a defect we clean up in a later chapter - it's the reason the book exists. A calculus that couldn't express this couldn't describe two people trying to book the last seat on a flight, which is most of what computers do all day.

---

## Testing that nothing happens

Which brings us to the loser.

An input with nobody sending is just *stuck*. In the calculus, `x(z).P` on its own is a perfectly good term that doesn't reduce. Not an error, not a bug - a normal form. The system is waiting, possibly forever, and that's allowed.

So how do you test for it?

Here's the obvious attempt:

```go
func TestReceiveWithNoSenderIsStuck_Bad(t *testing.T) {
	x := make(Name)
	out := make(Name)

	go Recv(x, out)

	time.Sleep(50 * time.Millisecond)

	select {
	case <-out:
		t.Fatal("Recv produced something with no sender")
	default:
	}
}
```

And that's rubbish, isn't it. It's slow. It's flaky - fifty milliseconds is either far too long or, on a loaded CI box, nowhere near long enough. And even when it's green it hasn't shown what we wanted. It says *nothing had happened yet when I looked*. We wanted *nothing could have happened*.

That gap is most of why testing concurrent code is miserable, and it's usually where a book like this starts apologising and reaching for a longer sleep.

We're not going to. This is what `testing/synctest` is for.

`synctest.Test` runs your test inside a **bubble** - it and every goroutine started within it. And inside the bubble:

```go
synctest.Wait()
```

`Wait` blocks until every *other* goroutine in the bubble is **durably blocked**: parked on a channel operation that only another goroutine in the same bubble could ever unblock.

Read that with your π-calculus hat on. When `Wait` returns, no reduction is possible. Nothing can move without the observer moving first.

It's a normal form detector, in the standard library.

```go
func TestReceiveWithNoSenderIsStuck(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		x := make(Name)
		out := make(Name)

		go Recv(x, out) // nobody will ever send on x

		synctest.Wait()

		select {
		case <-out:
			t.Fatal("Recv produced something with no sender")
		default:
		}
	})
}
```

No sleep, no timeout, no flake, and the claim it makes is the strong one.

### One hazard, before it bites you

If every goroutine in the bubble is durably blocked and nobody has called `Wait`, `synctest.Test` calls it a deadlock and fails the test.

Which is awkward, because we just wrote a test asserting that a stuck process exists. Stuck is legal here.

So: **the observer must reach `Wait` before it parks.** The test goroutine holds the bubble open. If you block on a channel nothing will ever deliver on, you get a deadlock failure instead of the assertion you wanted - and that failure is about your test, not about the term.

You will do this. Everyone does this.

---

## What we've deliberately not done

- **Parallel composition** as an idea in its own right. We've typed `go` a few times without saying what `P | Q` *means*, or why it's commutative. Chapter 3.
- **Fresh names.** `make(Name)` is doing something significant - creating a name nobody else can get hold of, written `(νx)P` in the calculus - and I've walked straight past it. Chapter 4.
- **Anything happening twice.** Every process here does one thing and dies. Chapter 6, where we meet `!P`.
- **Choice.** I used a `select` above without explaining it. Chapter 7.

---

## Exercises

1. **Chain of three.** Pass a name through `Recv` and then onward to a second process before it reaches `out`. How many channels did you need? Why that many?

2. **Delete the `go`.** In `TestTwoProcessesCommunicate`, call `Send(x, y)` directly instead of spawning it. Predict the failure before you run it. Then run it. Was it the failure you predicted?

3. **Both inputs.** Extend `TestTwoInputsOneOutput` with a second `Send`. Now both inputs can be satisfied. Does the non-determinism go away? Write out the reductions and say precisely what has and hasn't changed.

4. **Break the good test.** Make `TestReceiveWithNoSenderIsStuck` fail by adding exactly one line to the test body. What does that tell you about what it's really asserting?

---

## Recap

- `type Name chan Name`. Names are all there is, and they're compared by identity, because they have no contents.
- Output is `x̄y`, input is `x(z)`, and one of each on the same channel is the only thing that ever happens.
- The dot is a prefix: the only sequencing we have, and it freezes everything to its right.
- A process is a function that takes names and returns nothing. `go` is parallel composition; a plain call is the dot.
- Church-Rosser doesn't hold, and that's the point rather than a problem.
- `synctest.Wait()` plus `select`/`default` buys you sound negative assertions about concurrent code.

## Next

Chapter 2: the dot, properly. What is a continuation actually *for*, why is `x(z).P` a binder from the same family as `λz.P`, and what breaks if you leave the `P` off the end.
