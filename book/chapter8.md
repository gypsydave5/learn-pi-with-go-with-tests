# Chapter 8: Numbers that talk back

## What we're covering

We have the whole grammar now. Names, input, output, `0`, `|`, `ν`, `!`, `+`. Eight things.

And no data. No integers, no booleans, no strings, nothing. I told you in chapter 1 that if you couldn't say something with names alone then that was the calculus telling you something, and I've been quietly dodging the bill ever since.

Time to pay it. We're going to build the natural numbers out of processes.

## Setup

```
mkdir numbers && cd numbers
```

Everything from previous chapters.

---

## The move

Church's trick with functions is well known: if all you've got is functions, then data has to *be* functions, and a number becomes "a thing that applies something to something else that many times".

Our situation is the same shape with a different primitive. All we've got is interaction. So data has to *be* interaction, and a number is going to be **a process that talks to you a certain number of times**.

Two isn't a value. Two is a thing that says "tick, tick, done".

That's the whole idea and the rest of this chapter is carpentry.

---

## First, sending more than one name

We need some plumbing, because everything we've written sends exactly one name and we're about to want two.

Write it like this:

```
x̄⟨a,b⟩.P        send both a and b on x
x(u,q).P        receive two names on x, call them u and q
```

That's **polyadic** communication, and there's no new rule for it - it's sugar. We can already do it with what we've got.

The question is how.

### The obvious way, which doesn't work

Send twice.

```
x̄⟨a,b⟩.P   ≝   x̄⟨a⟩.x̄⟨b⟩.P
x(u,q).P   ≝   x(u).x(q).P
```

Looks fine. Sender puts two names down the channel, receiver takes two off. What could possibly go wrong?

Chapter 1, is what could possibly go wrong.

```
x̄⟨a⟩.x̄⟨b⟩.0  |  x(u).x(q).P  |  x(u).x(q).Q
```

One sender, two receivers, and the channel has no idea that `a` and `b` were meant to arrive together. The first receiver takes `a`. Then the second receiver takes `b`. Now both of them are stuck waiting for a second component that isn't coming, and neither has a pair.

Or, just as bad, with two senders:

```
x̄⟨a⟩.x̄⟨b⟩.0  |  x̄⟨c⟩.x̄⟨d⟩.0  |  x(u).x(q).P
```

Perfectly legal reduction: receiver gets `a`, then gets `c`. It now holds the first half of one message and the first half of another, and is entirely satisfied that it has received a pair.

Nothing in the rules is being violated. That's the trouble. A channel is not a queue and it does not know that two outputs belonged together - as far as the calculus is concerned, `x̄⟨a⟩.x̄⟨b⟩.0` is just a process that does two unrelated things in order.

### The way that works, and it's nuts

Here's where I'd normally say "so obviously the fix is X". I'm not going to, because you will not guess it, and I didn't either.

You have no locks. No queues. No transactions, no sequence numbers, no way to tag two messages as belonging together, and no way to ask a channel to hold still for a second. There is nothing in this calculus that could possibly deliver two things atomically.

So how on earth do you get a guaranteed pair out of a system whose only primitive is "one name goes down one channel"?

You do the conversation somewhere private.

```
x̄⟨a,b⟩.P   ≝   (νc)( x̄⟨c⟩.c̄⟨a⟩.c̄⟨b⟩.P )
x(u,q).P   ≝   x(c).c(u).c(q).P
```

Make a fresh channel. Send *that* on `x` - one output, so exactly one receiver gets it - and then send both components down the fresh one.

Now nobody can interleave, because there's nothing to interleave with. Whoever won the race on `x` is the only process in the universe that can name `c`, so the two components have a wire to themselves.

That's scope extrusion in the service of atomicity, and it's why chapter 4 and chapter 5 were worth the trouble. **What makes a two-name message be one message is a name nobody else can touch.**

> **You have been doing this all day**
>
> If that trick felt like a rabbit out of a hat, here's the thing: it's how the internet works, and both the good version and the bad one are deployed at enormous scale right now.
>
> **`accept()` is the encoding.** A server listens on a well-known port - the public channel. A client connects. And `accept()` hands back *a brand new socket*, identified by the four-tuple, which only those two endpoints can name. Rendezvous in public, conversation in private. That's `(νc)x̄⟨c⟩` with a syscall.
>
> **HTTP/1.1 is the naive version.** Requests go down one connection and replies are matched *by order* - first response belongs to the first request. There's no way to say which reply goes with which request, so it relies on sequence. Which is exactly our broken encoding, and it fails in exactly the way you'd now predict: one slow response blocks everything queued behind it. Head-of-line blocking.
>
> **HTTP/2 fixed it the other way.** Every frame carries a stream ID saying which request it belongs to. That's the solution we couldn't write down - correlation identifiers - and it works fine.
>
> Except that all those streams still sit inside one TCP connection, so a single lost packet stalls every one of them at once. Head-of-line blocking, one layer down. QUIC's answer was to give each stream genuinely independent delivery. Which is to say: after twenty years, back to private channels.
>
> There's a general lesson in the layering, and it's worth more than the history. TCP hands HTTP a perfectly good private channel. HTTP then holds *many conversations inside it*, and at that point the channel isn't private any more - it's shared between everything in flight, and the problem comes straight back.
>
> **A private channel stops being private the moment you multiplex over it.** One conversation per wire is the whole guarantee.
>
> And notice what the calculus couldn't do. Stream IDs need *data* - a number to tag a frame with - and we haven't got any. The π-calculus is forced into the structural answer because the tagging answer isn't available to it. Which is a decent argument for reductive systems in general: take away the tools that let you paper over a design problem and you get pushed into solving it.


In Go, the sugar is the encoding, written out:

```go
func Send2(x Name, a Name, b Name) {
	c := make(Name)
	x <- c
	c <- a
	c <- b
}

func Recv2(x Name) (Name, Name) {
	c := <-x
	return <-c, <-c
}
```

### Sending nothing at all

The same trick covers arity zero, which we're about to need for "tick":

```
x̄⟨⟩.P   ≝   (νc)x̄⟨c⟩.P
x().P   ≝   x(c).P
```

Send a fresh name that nobody will ever use. All the information is in the fact that a communication happened - which, if you cast your mind back to chapter 1, is exactly what `chan struct{}` is for.

> **A side condition**
>
> In every definition above, `c` must be **fresh**: it must not occur free in `P`, or in `a` or `b`.
>
> Get that wrong and you know exactly what happens, because it's happened twice already. On the sending side, `(νc)` would capture a `c` that `P` was relying on from outside - chapter 5's side condition. On the receiving side, `x(c)` would shadow it - chapter 2's capture.
>
> Note that in `x().P ≝ x(c).P` the `c` is perfectly well bound. It's just never *used*, which is a different thing, and Go will tell you off about it: this is the `declared and not used` sink from chapter 2, so in the Go you write `<-x` and don't name it at all.
>
> As ever, the condition costs nothing. If your `P` happens to have a free `c`, alpha-convert and pick another letter. There are always more letters.

```go
func Signal(x Name) { x <- make(Name) }
func Await(x Name)  { <-x }
```

---

## Warming up: booleans

```
True(b)  ≜ !b(t,f). t̄⟨⟩.0
False(b) ≜ !b(t,f). f̄⟨⟩.0
```

A boolean sits on `b` waiting. You hand it two channels. It signals on one of them.

That's the entire content of "true" and "false" here: not a value, but *which of your two channels gets poked*. Truth is a routing decision.

Note the `!` on the front - chapter 6's inexhaustible supply. A boolean you could only ask once would be a poor sort of boolean.

And note what `if` becomes. There's no `if`. You hand the boolean the two things you might want to happen, and it picks. The conditional isn't a construct, it's a conversation.

---

## Your turn: the numbers

Here's the encoding. Two definitions.

```
Zero(c)   ≜ !c(s,z). z̄⟨⟩.0
Succ(c,n) ≜ !c(s,z). s̄⟨⟩. n̄⟨s,z⟩.0
```

Read `Zero` first: sit on `c`; when somebody hands you two channels called "successor" and "zero", signal on the zero one. That's it. That's the number nought - a thing that immediately says "done".

Now `Succ`. It has *two* channels: `c`, where it's asked, and `n`, which connects it to the number one smaller than it. When asked, it signals once on `s` - one tick - and then hands `s` and `z` straight on to `n`, which does whatever it does.

So the number three is a chain of three `Succ`s wired to a `Zero`, and asking it produces tick, tick, tick, done.

```
Three(c) ≜ (νa)(νb)( Zero(a) | Succ(b,a) | Succ(c,b) )
```

Each `ν` is a private wire between one numeral and the next. Nobody outside can reach into the middle of a number.

Here's what to write:

```go
// Zero signals on the zero channel and nothing else.
//
//	Zero(c) ≜ !c(s,z).z̄⟨⟩.0
func Zero(c Name)

// Succ signals once on the successor channel, then passes the
// question on to the numeral below it.
//
//	Succ(c,n) ≜ !c(s,z).s̄⟨⟩.n̄⟨s,z⟩.0
func Succ(c Name, n Name)

// Numeral builds the number k.
func Numeral(k int) Name
```

And here's the test:

```go
func TestNumerals(t *testing.T) {
	for k := 0; k < 5; k++ {
		if got := Count(Numeral(k)); got != k {
			t.Errorf("Numeral(%d) counted %d", k, got)
		}
	}
}
```

Which needs one more thing - something to do the counting.

---

## The observer is a choice

```go
func Count(c Name) int {
	s := make(Name)
	z := make(Name)

	Send2(c, s, z)

	n := 0
	for {
		select {
		case <-s:
			n++
		case <-z:
			return n
		}
	}
}
```

Look at that `select`. Chapter 7 said the observer, written down properly, is a choice - `a(q).0 + b(q).0`, offer two inputs and take whichever comes.

Here it is doing real work. The counter offers two inputs, `s` and `z`, and has no idea which will arrive. The *number* decides, by ticking or by finishing, and the counter just keeps score.

We could not have written this in chapter 6. There was no way to wait on two things at once.

---

## What you've actually built

Go and run it. `Numeral(3)` really does count as 3, and there is not a single integer anywhere in the implementation. Just channels, goroutines, and a chain of processes handing a question down the line.

Two things worth noticing about the shape.

**The number is spread out.** `Numeral(3)` isn't a thing sitting in one place - it's four processes and three private channels, and the "3" only exists in the length of the chain. You cannot point at where the three is.

**Counting it is destructive of nothing.** Ask again and you get 3 again, because of all those `!`s. That's chapter 6's inexhaustible supply doing exactly the job it was introduced for. Take the `!` off and you get a number that works once and then isn't a number any more.

Also: `goleak` will scream. All those replications are still sitting there waiting to be asked again, which is correct, and which is what `!` means. If you want them to stop you'd need to give each one a `+` and a stop channel, and then it isn't `!` any more. Chapter 7 again.

---

## Arithmetic is plumbing

Now we've got numbers, we should be able to do things to them. And the striking thing about doing things to them is that **none of it is calculation**. Every operation below is a matter of wiring the tick channel and the done channel to the right places. Addition is joining two chains end to end. Multiplication is nesting one inside the other. Nobody adds anything.

### Addition

```
Add(c,m,n) ≜ !c(s,z). (νq)( m̄⟨s,q⟩.0 | q(). n̄⟨s,z⟩.0 )
```

Somebody asks our sum for its ticks, handing us the real `s` and `z`.

We ask `m` first - but we give it the *real* tick channel and a **private** finish channel `q`. So `m`'s ticks go straight through to the caller, who counts them, and `m`'s "done" comes back to us instead of ending things prematurely.

Then, when `q` fires, we know `m` has run out. Now we hand the whole question to `n`, with both real channels this time. Its ticks go to the caller and its "done" ends the sum.

The caller sees `m` ticks, then `n` ticks, then done. Which is `m + n`, and nothing anywhere computed a sum. We intercepted an "I'm finished" and turned it into a "your turn".

**That `ν` is the entire trick.** Give `m` the real `z` and the sum stops after the first numeral. The private finish channel is what lets us splice two sequences together.

### Multiplication

Same idea, one level up. Multiplication is `m` copies of `n`, so we want to run a whole `n` for every tick of `m`.

```
Mult(c,m,n) ≜ !c(s,z). (νq)( m̄⟨q,z⟩.0 | R )
       R    ≜ q(). (νr)( n̄⟨s,r⟩.0 | r(). R )
```

Now `m` gets a **private tick channel** `q` and the real `z`. Its ticks come to us, and when it runs out it ends the whole thing directly.

Each time `q` fires, `R` runs a complete copy of `n` - real `s`, private finish `r` - and only when *that* has said it's done does `R` go back for `m`'s next tick.

Trace `Mult(2,3)`: `m` ticks, we run a 3, three ticks reach the caller. `m` ticks again, another 3, three more. `m` says done, and that goes straight to `z`. Six and finished.

**And `R` has to be recursion, not replication.** Write it as `!q().(...)` and `m`'s ticks can all be accepted at once, several copies of `n` run simultaneously, and `m` reaches its "done" while some of them are still going - so the sum ends early and the count is wrong. This is chapter 6's distinction, and this is the first time in the book that getting it wrong actually breaks something.

---

## Testing this is fair, for once

Chapter 4 couldn't test the interesting thing. Chapter 5 could only test half of it. This chapter, unusually, the test is completely honest, and it's worth knowing why.

**Numerals are deterministic.** There's no `+` in the encoding, no race, nothing that could go two ways. Ask `Numeral(3)` a question and exactly one sequence of events is possible: tick, tick, tick, done.

When a process is deterministic, comparing the sequence of things that came out really does tell you whether two processes are the same, because there's only ever the one sequence. `Count` sees everything there is to see.

Enjoy it, because it doesn't last. The moment `+` gets involved, "the same sequence came out" stops being enough - that's the puzzle at the end of chapter 7, and it's chapter 10's business.

---

## The bill

Let's be honest about what this costs.

`Numeral(3)` is four goroutines. `Numeral(1000000)` is a million and one, and counting it is a million sequential rendezvous. A 32-bit integer would be four billion goroutines and four billion handoffs, which is not a program, it's a threat.

This is unary arithmetic implemented in message passing, and it is the slowest way to count that anybody has ever devised.

But it isn't a *defect*, and this is the point I want to leave you with. It's the price of the reductive move. We said everything is interaction, and then we asked for a number, and the calculus gave us one built entirely out of interaction, and it costs what interaction costs. Church numerals in a real language are slow for the same reason and nobody holds it against them.

What you've bought for that price is that there's nothing else in the system. No primitives, no built-ins, no "and the runtime provides integers". Eight pieces of grammar and a chain of processes, and out comes arithmetic.

---

## Exercises

1. **(paper)** Reduce `True(b) | b̄⟨t,f⟩.0` all the way, expanding the polyadic sugar in full. How many actual communications happen? How many `ν`s did you create along the way?

2. **(paper)** Write out `Numeral(2)` as a single term with all the definitions expanded and all the `ν`s in place. Then reduce it against a counter that asks once. It's long. Do it anyway - it's the last time this book will ask you to expand everything, and it's worth seeing the machinery once.

3. **(paper)** What does `!c(s,z).s̄⟨⟩.0` count as? It's not a numeral. What is it?

4. **(Go)** Write `Bool`, `True`, `False` and a `Cond` that takes a boolean channel and two things to do. Test it.

5. **(Go)** Write `Add`. Test that `Add(Numeral(2), Numeral(3))` counts as 5. Then give `m` the real `z` instead of the private `q` and confirm you get 2.

6. **(Go)** Write `Mult`. Then deliberately write `R` as a replication instead of a recursion, and find an `m` and `n` for which it gives the wrong answer. How reliably does it go wrong?

7. **(Go)** Property test the lot. For all small `a` and `b`, `Count(Add(a,b)) == a+b` and `Count(Mult(a,b)) == a*b`. This is a real property test against a real oracle, and we're only allowed it because numerals are deterministic.

8. **(Go)** `IsZero`. Given a numeral, produce a boolean. Careful: you mustn't consume the ticks of a number somebody might want to count later.

9. **(paper, hard)** **Predecessor.** Given a numeral, produce one that's smaller by one. There's no obvious wiring for this - the chain runs one way, and there's no handle on the second-from-last link.

    It's famously the hard one. The story goes that Kleene worked out how to do it for Church's encoding while sitting in a dentist's chair having his wisdom teeth out, which tells you something about how it was going.[^kleene] Have a proper go before you look anything up.

    A hint, if you want one: you don't have to produce the answer while you're counting. You could count *twice*, and lag one of them.

10. **(Go, cruel)** Time `Count(Numeral(k))` for `k` up to about 100,000. Plot it if you like. Then work out how long a 32-bit integer would take and write the number down somewhere you can look at it.

[^kleene]: Nitrous oxide. Make of that what you will.

---

## Recap

- Data is behaviour. A number is a process that talks to you a certain number of times.
- **Polyadic communication is sugar**: send a fresh private channel and then the components down it. The `ν` is what stops two messages interleaving.
- Arity zero works the same way - a signal is a fresh name nobody uses.
- Booleans are routing decisions: hand over two channels, get one of them poked.
- `Zero(c) ≜ !c(s,z).z̄⟨⟩.0` and `Succ(c,n) ≜ !c(s,z).s̄⟨⟩.n̄⟨s,z⟩.0`. A number is a chain.
- Every numeral needs `!` or it's a number you can only use once.
- The counter is a `+`. Chapter 7 earning its keep.
- Numerals are deterministic, so comparing outputs is a sound test - which is a luxury, and chapter 10 explains why.
- Arithmetic is **plumbing**. Addition splices two chains, multiplication nests them. Nothing computes anything.
- The private finish channel is what lets you splice - intercept an "I'm done" and turn it into a "your turn".
- It's desperately slow, and that's the honest price of having no primitives at all.

## Next

Chapter 9. We've built numbers out of processes. Now we build *functions* out of processes - Milner's encoding of the λ-calculus into the π-calculus, which is the moment this stops looking like a nice model of concurrency and starts looking like a foundation. And the reply channel from chapter 1 turns out to have been the whole trick.
