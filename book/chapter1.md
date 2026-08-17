# Chapter 1: A name is a channel is a name

**Draft 5.** You need Go 1.25 or later.

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

------

## Why bother?

As you probably know, the λ-calculus is built out of three things: variables, functions, and applying one to the other. That's the entire grammar. Everything else - booleans, numbers, pairs, recursion - is something you *build* out of functions, rather than something the calculus hands you.

It's a reductive move, and a very good one. **Everything is a function.** If you've ever written `λf.λx.f (f x)` on a whiteboard and felt pleased with yourself, that's what you were feeling pleased about: numbers turning out to be functions all along. And if you haven't done this, don't worry, we're not going to rely on whether you've done lots of lambda calculus before. But your eyes may glaze over in the next few paragraphs.

The λ-calculus also comes with a theorem attached, and we need to get rid of it early on, because its failing shows us why we need the π-calculus.

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

The π-calculus is what happens when you allow one more thing: port names can be sent as messages. That's it, that's the entire difference, and it turns out to change everything. Milner's own summary of the resulting calculus is that the most prominent ingredient is the notion of a name - which by the end of this book I hope reads as a startling thing to have said, because it means there isn't anything else in there.

[^ccs]: CCS is worth a look on its own account if you like this stuff - it's simpler than the π-calculus and the bisimulation material is easier to meet there first. Milner's *A Calculus of Communicating Systems* (1980) is the original; *Communication and Concurrency* (1989) is the more approachable rewrite. Note that both have titles confusingly close to the π-calculus book, *Communicating and Mobile Systems* (1999).

------

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

But here's a thing you've probably written when practising with channels in Go:

```go
type Request struct {
	Data  string
	Reply chan Response
}

requests <- Request{Data: "hello", Reply: make(chan Response)}
```

A reply channel. It's more exciting than it looks: **you just sent a channel over a channel.** You handed a stranger the means of getting back to you. Before that send, nothing in the receiving goroutine could have referred to your reply channel - there was no expression it could have written. After it, there is. Channels aren't just piping and wiring, they're first-class values in Go.

You've met this shape before under other names. It's continuation-passing: rather than returning a value, you're handed somewhere to *put* the answer, and you put it there. It's also, at a grubbier scale, a webhook - here's a URL, call me back on it. The pattern keeps turning up because it's the only way to get an answer to someone whose address you didn't previously have.

So: what if a channel is the *only* thing you can send?

```go
type Name chan Name
```

A name is a channel is a name. And a name is a channel that can send a name. And a channel is a name that can send a channel. Ok, all together now...

This is the type system. Not an over-simplification for the sake of chapter one, not a starting point we fix later - that's the whole thing, for the rest of our book. No integers. No strings, no booleans, no functions, no structs. There are just **names**, and a name is a channel that carries names. Go's recursive type definitions let us write it down very neatly, which is lucky.

Your reaction is probably that it isn't enough. I remember that feeling when I first saw the λ-calculus. Let's see if we *can* build up the whole of computer science from that tiny kernel.

Right, let's make a name:

```go
s := make(Name)
```

Crikey, that was easy.

------

## The only thing that happens

The π-calculus has but two operations in total, and a single simple rule to bring the two together. Let's see those operations first.

### Output

Sending something on a channel. In the π-calculus we write this as `x̄⟨y⟩`, which we read as "send the name `y` on the channel `x`". Well, that's not hard to translate into Go:

```go
x <- y
```

### Input

The other operation - receiving something on a channel. In the π-calculus we say `x(z)`, which we read as "receive on the channel `x`, and call whatever arrives `z`". Again, I think we know what that looks like in Go:

```go
z := <-x
```

Two things mark out the output: the overbar on the channel name - the `x̄` - saying "output on x", and the angle brackets around what's being sent. Input gets round brackets instead. It's belt and braces, but it's the notation Milner used and it's what you'll meet in all the literature, so you may as well get used to reading it now.[^letters]

[^letters]: The overbar is a combining character, so it only really sits over one letter - `c̄h` looks like a mistake rather than a channel called `ch`. Which is why π-calculus papers use single-letter names for everything, and why we will too, in the calculus. In the Go, names can be as long as they need to be.

### Where does any of this happen?

So we've got our type - the `Name`. And we've got two operations - output and input, send and receive. That's all fine and dandy but it still needs somebody to *do* them - and, well, not just one person. This is all about communication, so we actually need at least one other thing to handle the other side. One sends, one receives. This is what we're working towards here - two independent things, running at the same time, meeting. Almost all the notation from here on exists to describe that, rather than to describe names.

So we call that thing that *does* the operation a **process**. A process is a thing that runs and interacts. That's the whole definition, and the vagueness is deliberate - we never say what a process is *made of*, only what it can do. We write them (in the syntax we're using) with capital letters, `P`, `Q`, `R`, the way we've been writing `x` and `y` for names.

We need three pieces of very specific process notation to get us off the ground, and then finally we can state the one magic rule.

#### zero

`0` is the inert process. Finished. Does nothing, offers nothing, nobody can interact with it. Exited. Slipped off this mortal coil. Dead. Unresponsive. You fill in your favourite metaphors.

What does that look like in Go? You know at the end of a function when it stops doing something? That.

#### a dot

A dot, a `.`, is "and then". `x̄⟨y⟩.P` reads *output `y` on `x`, and then carry on like `P`*. Nothing to the right of the dot will actually happen until that output has actually happened. It's "blocking".

If "and then `P`" makes it look like we're *naming* the process `P`, that's the notation being unhelpful and your eyes are working fine. `P` is a variable - akshually a **metavariable**, which is to say it belongs to the language we use to talk about terms, not to the terms themselves. It stands in for however much more of a process there is, and in a real process you'd just fill it in:

```
x̄⟨y⟩.P            where P is w̄⟨v⟩.0 ... so we just say
x̄⟨y⟩.w̄⟨v⟩.0       the same term, written out
```

There's no `P` left in the finished thing.[^defs]

[^defs]: Worth keeping this separate from something you may meet elsewhere. Some presentations *do* let you name a process, with definitions like `A(x) ≜ x(z).ā⟨z⟩.0`, and then `A` genuinely is a name for a process and genuinely does appear inside terms. It's a convenience for writing recursive behaviour down without going mad. We won't need it - we get repetition from `!` in chapter 6 - and `P` and `Q` are never that.

This is the only sequencing in the language. If you want two things in order, you chain prefixes: `x(z).y(w).P` gets an input on `x`, and *then* gets one on `y`. The second one cannot possibly go first.

What does it look like in Go? Well, you know how you write one line of Go, and then you write another line of Go underneath it? And you know how the first one happens before the second one? That. It looks like that. This is not rocket science.

And look - chain a couple of prefixes onto a `0` and you've written a whole process:

```
x(z).ȳ⟨z⟩.0
```

Take a name off `x`, pass it along on `y`, stop. An input, an output, and a full stop. No boilerplate, nothing declared, nothing hiding, behold! that's a complete program in the language, and you could read it. Well done you.

#### Parallel composition

Parallel composition - or the pipe, `|`. So when we write `P | Q`, what we mean is P and Q running at the same time. Both live, neither waiting on the other, and either one free to interact with anything else that's about. It'll get a chapter of its own, because "at the same time" has more in it than you'd think, but "at the same time" will do for now.

And how does that look in Go? C'mon, you've probably already worked it out. Sure, ok, let's do it. You know when you've got a process in Go - like a function. And you know when you start that little process going with a little `go` in front of it? And you know when you've done that more than once so you've got a few of the little guys off running at once? That. It's a bit harder than "the end of a function" or "a new line", but it really shouldn't be taxing you.

### THE ONE RULE ~~TO RULE THEM ALL~~

And now (finally) the rule:

```
x̄⟨y⟩.P  |  x(z).Q   →   P  |  Q{y/z}
```

Two processes, side by side. One does `P` things. One does `Q` things. The `P` flavour one is offering an output on `x`, the `Q` flavour one an input on `x`. They meet, ~~kiss briefly,~~ and the whole thing becomes something new: the two continuations, still side by side, with a substitution applied to the one that received. Now all the `z`s in `Q` are all `y`s. Wow did `P` just get `Q` pregnant?

> **`Q{y/z}`** is read "Q with y for z". Every occurrence of `z` in `Q` is replaced by `y`.
>
> It's easy to get backwards, and the notation is no help at all. The **new** name goes on top; the **old** one underneath. It's a fraction you're cancelling: `z` goes away, `y` arrives.
>
> A simpler (but less general) example makes it easy to see:
>
> ```
> x̄⟨y⟩.0  |  x(z).ā⟨z⟩.0   →   0  |  ā⟨y⟩.0
> ```
>
> The `z` became a `y`. Never the other way round.[^subst]

[^subst]: Some authors write `Q[y/z]`, and a few use an arrow, `Q[z ↦ y]`, which has the virtue of being impossible to misread. The `{y/z}` form is the most common and we'll stick with it. There's also a wrinkle I'm skating over: strictly it's every *free* occurrence of `z` that gets replaced, because `Q` might contain its own binder that happens to use the same letter, and that one is a different `z` entirely. Chapter 2.

You may have seen this substitution flavour before:

```
(λz.Q) y   →   Q{y/z}
```

Same rule, but in fancy lambda-land it's called a β-reduction, and it's what happens when function meets argument. Communication is what happens when an output meets an input, and *naming the same channel* is what decides they've met.

The π-calculus takes β-reduction and splits it into two halves that have to go and find each other first. Everything difficult and everything interesting falls out of that one change.

------

## The first test

Having fun yet? Let's watch one communication happen in Go.

```go
package pi

import "testing"

type Name chan Name

func TestOneCommunication(t *testing.T) {
	x := make(Name) // the channel
	y := make(Name) // the name we'll send

	go func() { x <- y }() // x̄⟨y⟩

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

------

## What a process is in Go

Mapping a process to Go is, as I've said, easy: it's just a function that takes names and returns nothing.

```go
func SomeProcess(x Name, y Name) { ... }
```

Three rules when we're playing around, to stay safe and not mess up our π:

**It only touches the names it was given.** No closures reaching out into the enclosing scope, no package-level state, no globals. If a process knows about a name, somebody handed it that name. That's the only way there is.

**No return values.** Not "usually not" - never. A process that wants to tell you something sends it on a name you gave it.

**`go` is parallel composition; a plain call is the dot.** Spawning is how you say `P | Q`. Calling is how you say "and then". A goroutine is a thread of causality, and `go` is the only way to get another one - everything reachable by plain calls from a single `go` is one process working through its prefixes in order.

```go
func Recv(x Name, y Name) {
	z := <-x   // x(z).
	Send(y, z) //      ȳ⟨z⟩   ← a call: this IS the continuation - the dot
}

func Both(x Name, y Name) {
	go P(x) // P | Q
	go Q(y) //   ← go: this is parallel composition - the pipe
}
```

A process blocked on a channel is a process sitting at a dot, waiting to interact. Not running, not finished - waiting. That's a perfectly respectable state and it's most of the state space.

Now notice what the signature *doesn't* tell you:

```go
func SomeProcess(x Name, y Name)
```

Which of those does it read from? Which does it write to? Both? Neither? You can't tell, and neither can the compiler, because there's one type and it hasn't got a direction. That should bother you a little, and we'll fix it with some types (but not how you're thinking). But first let's work on the untyped calculus.

------

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
// Send sends whatever the payload is on ch, and then stops.
//
// Which in π looks like:
//
//	x̄⟨y⟩.0
//
// where x is ch, and y is payload
func Send(ch Name, payload Name)

// Recv takes an input from ch, binding whatever arrives as z, then
// sends z on the Name out, and then stops.
//
// Which in π looks like:
//
//	x(z).ȳ⟨z⟩.0
//
// where x is ch, and y is out
func Recv(ch Name, out Name)
```

Each one is a line or two. Have fun!

------

## The race, in the flesh

Back at the start I said that Church-Rosser was gone. Well, let's watch it go.

Here's the term we're going to write:

```
x̄⟨y⟩.0  |  x(z).ā⟨z⟩.0  |  x(z).b̄⟨z⟩.0
```

Everything in there is something you've already met: the overbar, the dot, `0`, and `|` for two things running at the same time. So try and say what it means.

Now we do it together.

There's a process that outputs the name `y` on channel `x`, and then stops. Running at the same time, a second process that takes something off channel `x` and calls it `z`, and then it outputs that `z` on channel `a`, and then stops. And alongside both of those, a third process that takes something off channel `x` and calls *it* `z`, which it then outputs on channel `b`, and then stops.

Three processes. One message. Two of them want it. FIGHT!

In Go:

```go
func TestTwoInputsOneOutput(t *testing.T) {
	x := make(Name)
	y := make(Name)
	a := make(Name)
	b := make(Name)

	go Send(x, y) // x̄⟨y⟩.0
	go Recv(x, a) // x(z).ā⟨z⟩.0
	go Recv(x, b) // x(z).b̄⟨z⟩.0

	select {
	case <-a:
		t.Log("the second process won")
	case <-b:
		t.Log("the third process won")
	}
}
```

Run it with `go test -race -v -count=20` and read the log lines. You should see both.[^race]

[^race]: The `-race` is not decoration, and you should try it without. Left to itself the scheduler in a program this small is effectively deterministic - the goroutines park on `x` in the same order every time, the queue of waiting inputs is served first-in-first-out, and the same one wins on all twenty runs. Turning the race detector on instruments every memory access, which perturbs the timing enough to shake the scheduler out of its rut. Nothing about the term changed. Nothing about the program changed. Both outcomes were always there, and one of them simply never got picked until something entirely unrelated to the meaning of the program nudged it. Which is worth remembering the next time twenty green runs make you feel safe.

### Reducing it, twice

Two log lines means two behaviours. Let's do it on paper and see the same thing.

The rule, once more:

```
x̄⟨y⟩.P  |  x(z).Q   →   P  |  Q{y/z}
```

**First way.** The output pairs with the second process. Matching against the rule, `P` is `0` and `Q` is `ā⟨z⟩.0`, so the continuation becomes `ā⟨z⟩.0{y/z}`, which is `ā⟨y⟩.0`:

```
x̄⟨y⟩.0  |  x(z).ā⟨z⟩.0  |  x(z).b̄⟨z⟩.0
   →     0  |  ā⟨y⟩.0  |  x(z).b̄⟨z⟩.0
```

**Second way.** Same starting term. The output pairs with the third process instead. Now `Q` is `b̄⟨z⟩.0`, and the continuation becomes `b̄⟨y⟩.0`:

```
x̄⟨y⟩.0  |  x(z).ā⟨z⟩.0  |  x(z).b̄⟨z⟩.0
   →     0  |  x(z).ā⟨z⟩.0  |  b̄⟨y⟩.0
```

One term. Two reductions. Both perfectly legal, and nothing in the rules picks between them.

Now look at where we've ended up. In the first, there's an output waiting on `a` and the third process is stuck on `x` forever, because there is no output left on `x` to pair with. In the second, exactly the reverse.

Those two states are not the same state, and there's no sequence of reductions that turns either into the other. Let the observer consume the pending output and you're left with `x(z).b̄⟨z⟩.0` in one case and `x(z).ā⟨z⟩.0` in the other. Still different. Still stuck. Nothing to be done.

Which is the `t.Log` line you just watched flip.

You might be wondering where *we* are in that description. We'll talk about that later, promise.

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

------

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

(YES GO YES YOU ARE THE BEST LANGUAGE)

`synctest.Test` runs your test inside a **bubble** - it and every goroutine started within it. And inside the bubble:

```go
synctest.Wait()
```

`Wait` blocks until every *other* goroutine in the bubble is **durably blocked**: parked on a channel operation that only another goroutine in the same bubble could ever unblock.

Read that with your π-calculus hat on. When `Wait` returns, no reduction is possible. Nothing can move without the observer moving first. Everything has now been reduced to its simplest form. So for us, used like this, `synctest` is a π-calculus normal form detector, bundled in the Go standard library.

(YOU BEAUTIFUL GOPHER YOU YESSSSS GO ON MY SON)

```go
func TestReceiveIsStuckUntilSomethingArrives(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		x := make(Name)
		y := make(Name)
		out := make(Name)

		go Recv(x, out)

		synctest.Wait()

		select {
		case <-out:
			t.Fatal("Recv produced something with no sender")
		default:
		}

		x <- y // now give it what it was waiting for

		if got := <-out; got != y {
			t.Errorf("got %v, want %v", got, y)
		}
	})
}
```

No sleep, no timeout, no flake. And there are three moves in there, doing three different jobs.

**It's stuck.** That's `Wait`, and it's the only *positive* assertion available to us. `Wait` returns when nothing in the bubble can move, so reaching the line after it is proof - not a guess, not a hopeful pause - that no reduction is possible. The term is in normal form.

**Nothing's coming out.** That's the `select`/`default`. And now that we know nothing *can* move, an empty `out` means something. It isn't "nothing has happened *yet*", it's "nothing happened, and nothing was going to".

**Stuck because it needed \*this\*.** That's the send on `x`, and it's the one doing the real work. The first two only tell us the process didn't do anything. This one tells us *why* - it was waiting on `x`, specifically, and the proof is that a name on `x` is exactly what set it going again.

Which is a slightly sneaky way of establishing something we can't establish directly. We never inspect the process to see whether it's blocked - we work out what it was waiting for by finding the one thing that unsticks it. Hold that thought until chapter 10, where working out what a process *is* by seeing what it responds to turns out to be the whole game.

And note what we still can't say. We cannot assert that a process is stuck *forever*. Forever isn't a property of a state, it's a claim about every possible future, and no test is going to run for that long. Our own test makes the point rather well: the process looked utterly stuck, right up until we sent something on `x`.

### One hazard, and it will bite you

Try writing that test without the last three lines - just the `Wait` and the `select`, no output on `x` at all. It looks better. It's more obviously about the one thing we're claiming.

It panics.

When the function you handed to `synctest.Test` returns, `Test` waits for every goroutine in the bubble to finish. Our `Recv` is parked on `<-x` and will be parked there until the heat death of the universe, so it never finishes, and `Test` quite correctly calls that a deadlock.

Which is awkward, because a stuck process is legal in the π-calculus. `x(z).P` with nothing to pair with is a perfectly respectable term. It's just that a bubble insists on tidying up after itself: **every process you start inside one has to be able to reach `0` before the test ends.**

So we assert the stuckness in the middle and let the process finish afterwards. Slightly annoying, and probably the right constraint anyway - it's the same instinct as not leaking goroutines in real code, which is a thing we'll take seriously in chapter 6.

------

## Exercises

1. **Chain of three.** Pass a name through `Recv` and then onward to a second process before it reaches `out`. How many channels did you need? Why that many?
2. **Delete the `go`.** In `TestTwoProcessesCommunicate`, call `Send(x, y)` directly instead of spawning it. Predict the failure before you run it. Then run it. Was it the failure you predicted?
3. **Both inputs.** Extend `TestTwoInputsOneOutput` with a second `Send`. Now both inputs can be satisfied. Does the non-determinism go away? Write out the reductions and say precisely what has and hasn't changed.
4. **Break the good test.** Make `TestReceiveIsStuckUntilSomethingArrives` fail by adding exactly one line to the test body. What does that tell you about what it's really asserting?

------

## Recap

- `type Name chan Name`. Names are all there is, and they're compared by identity, because they have no contents.
- Output is `x̄⟨y⟩`, input is `x(z)`, and one of each on the same channel is the only thing that ever happens.
- The dot is a prefix: the only sequencing we have, and it freezes everything to its right.
- A process is a function that takes names and returns nothing. `go` is parallel composition; a plain call is the dot.
- Church-Rosser doesn't hold, and that's the point rather than a problem.
- `synctest.Wait()` plus `select`/`default` magics a normal-form checker into existence.

## Next

Chapter 2: the dot, properly. What is a continuation actually *for*, why is `x(z).P` a binder from the same family as `λz.P`, and what breaks if we take the `P` off the end.
