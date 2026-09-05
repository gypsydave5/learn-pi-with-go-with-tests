# Chapter 3: Everything, all at once

## What we're covering

`|`. We've been writing it since chapter 1 and calling it "at the same time", which is true and incomplete.

By the end of this chapter our reduction rule will actually work, we'll be allowed to cross out those `0`s I keep leaving lying about, and `go test -race` will have caught us cheating.

## Setup

```
mkdir all-at-once && cd all-at-once
```

Bring `Send` and `Recv` with you.

---

## Our rule is broken

Here's the rule, from chapter 1:

```
x̄⟨y⟩.P  |  x(z).Q   →   P  |  Q{y/z}
```

And here's a perfectly ordinary term:

```
a(u).0  |  x̄⟨y⟩.0  |  b(q).0  |  x(z).0
```

Four processes. Two of them - the second and the fourth - are an output and an input on `x`, and they should obviously communicate. Any Go programmer can see it. Spawn those four goroutines and the `x` pair will find each other in a heartbeat, because goroutines don't care who was declared first.

But look at the rule. It matches an output *immediately next to* an input. Ours have two other processes wedged between them. There is no `→` we're allowed to write.

So the rule is wrong, or at least useless, and we need to fix it. Go is right; our notation is being precious.

---

## Order isn't real

The problem is that writing things down forces an order on them that the world doesn't have.

`P | Q` is a *soup*. Both processes are in there, floating about, and any two of them that offer matching input and output can react. There's no first, no left, no order of declaration. `a(u).0 | x̄⟨y⟩.0 | b(q).0 | x(z).0` isn't a list of four processes, it's a bag containing four processes.

We know this is true in Go, because there's no way to ask which goroutine came first. There's no goroutine ID, no ordering, no `runtime.WhichOneWasFirst()`. Go hides it because it isn't there.

So we need a way to say "these two ways of writing it down are the same term". That's **structural congruence**, and we write it `≡`:

```
P | Q       ≡  Q | P            you can swap them
(P | Q) | R ≡  P | (Q | R)      brackets don't matter
P | 0       ≡  P                nothing is nothing
```

Commutative, associative, with `0` as the unit. If you've met a monoid, that's one. If you haven't: it means you can shuffle and rebracket freely, and drop any `0`s you find.

That's it. Three lines, and they say "the way you wrote it down doesn't count".[^more]

[^more]: There are more laws to come. Chapter 4 adds a couple for `ν`, and chapter 6 adds the big one for replication. `≡` grows as the language does.

---

## Reduction, up to ≡

Now we can fix the rule. Not by changing it - by changing when we're allowed to *apply* it:

```
P ≡ P'     P' → Q'     Q' ≡ Q
─────────────────────────────
           P → Q
```

Read the top line left to right: shuffle `P` into some more convenient arrangement `P'`, do a reduction, then shuffle the result back however you like.

So our stuck term isn't stuck at all:

```
a(u).0 | x̄⟨y⟩.0 | b(q).0 | x(z).0
  ≡  x̄⟨y⟩.0 | x(z).0 | a(u).0 | b(q).0     shuffle the pair together
  →  0 | 0 | a(u).0 | b(q).0               now the rule fits
  ≡  a(u).0 | b(q).0                       and tidy up
```

Three steps, and only the middle one is a reduction. The other two are just rearranging, because rearranging is free.

**And there's the answer to something I fobbed you off about in chapter 1.** I kept leaving `0`s lying around in reductions and told you to cross them out anyway. `P | 0 ≡ P` is why. A finished process contributes nothing, offers nothing, and can be swept up at any time.

---

## Why so fussy?

Reasonable objection: we just spent three laws and an inference rule establishing something you could see by squinting. Those four processes obviously communicate. Spawn them and watch them do it.

Sure. But "obviously" doesn't scale, and nobody else can check it.

The reduction rule only matching things that are literally side by side is a *feature*. It means applying it takes no judgement at all - you either have the shape or you don't. And `≡` then says, just as mechanically, exactly which rearrangements are allowed. So a reduction becomes a sequence of steps that something with no understanding whatsoever could verify: a tired reader, a colleague, a proof assistant, you at 2am.

Running the program is a different kind of evidence, and a weaker one. It tells you what happened, once, on this machine - which chapter 1 laboured somewhat. The rules tell you what *must* happen, everywhere, always. That's the deal: strictness up front, certainty afterwards.

Which is a bargain you've already made once, because it's why you write tests instead of saying "I had a look and it seemed fine".

---

## What ≡ is *not*

Careful here, because it's an easy and expensive mistake.

`≡` says two *ways of writing* a term are the same term. It never says two terms *do* the same thing.

```
x̄⟨y⟩.0 | x(z).0        can communicate
0                       cannot
```

The first one reduces to something very like the second. They are absolutely not `≡`. Structural congruence shuffles; it doesn't run anything.

"Do these two processes behave the same?" is a completely different question, it's much harder, and it's chapter 10.

---

## A precedence gotcha

While we're here: the dot binds tighter than the bar.

```
x(z).P | Q      means      (x(z).P) | Q
```

So if you want the input to be followed by *both* processes, you need brackets:

```
x(z).(P | Q)
```

Which reads: receive on `x`, and then become two processes running side by side, both of which know `z`.

That's about to matter.

---

## Your turn

Here's a process that tells two people about one name:

```
x(z).(ȳ⟨z⟩.0 | w̄⟨z⟩.0)
```

Receive a name on `x`. Then, at the same time, send it on `y` and send it on `w`.

The test:

```go
func TestTellsBoth(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		x := make(Name)
		y := make(Name)
		w := make(Name)
		n := make(Name)

		go Tell(x, y, w)

		x <- n

		if got := <-y; got != n {
			t.Errorf("on y: got %v, want %v", got, n)
		}
		if got := <-w; got != n {
			t.Errorf("on w: got %v, want %v", got, n)
		}
	})
}
```

And what it wants:

```go
// Tell receives a name on x, then sends it on both y and w,
// concurrently.
//
//	x(z).(ȳ⟨z⟩.0 | w̄⟨z⟩.0)
func Tell(x Name, y Name, w Name)
```

Write it. It's four lines.

---

## The one rule, and why

Back in chapter 1 I asked you to take something on trust:

**A spawned process only touches the names it was passed.**

Here's the reason. The π-calculus has no memory. There is no variable, no heap, no place two processes can both look. If a process knows a name, it is because somebody handed it that name - and the term records exactly who handed what to whom. That's not a limitation of the notation, it's the entire content of the notation. Connectivity *is* the program.

So the moment your Go coordinates through anything other than passing names, your term has stopped describing your program.

Now, you were probably never going to declare a shared `var` and mutate it from three goroutines. What you *will* do, because everybody does, is this:

```go
func TellSneaky(x Name, y Name, w Name) {
	z := <-x

	go func() { y <- z }()
	go func() { w <- z }()
}
```

Perfectly ordinary Go. No mutation, no shared state, no data race - `go test -race` is delighted with it, and the test passes.

And you can no longer read the program off the page.

Look at `go func() { ... }()`. What did that process get? Nothing, apparently - the spawn site is empty. To find out you have to read the body, spot `y` and `z`, and then look *outwards* to work out where they came from. The connectivity is still there, but it's implicit, and the term you'd write on a whiteboard has stopped matching the code.

You have met this before, and you have probably been told off about it. The worker loop:

```go
for _, j := range jobs {
	go func() {
		work(j) // j comes from out there somewhere
	}()
}
```

For years that was a real, biting bug. Every goroutine shared the one loop variable, and by the time they got round to running it was sitting on the last value. Whole afternoons went into that one. And the fix everybody learned was to pass the thing in:

```go
go func(j Job) { work(j) }(j)
```

Go 1.22 made loop variables per-iteration, so the bug is gone and the original version is fine now. But the habit was right for a better reason than the one that taught it to us. Passing the argument doesn't only fix a lifetime problem - it says, at the spawn site, what the new process was handed. That's the part worth keeping now that the bug has gone.

The rule isn't "don't use closures". It's: **say what the process gets.**

```go
go Send(y, z)                    // fine
go func(y, z Name) { y <- z }(y, z)  // also fine, if you like anonymity
```

Both of those name the arguments at the spawn site. Read either one and you know exactly which names crossed into the new process, which is precisely what `x(z).(ȳ⟨z⟩.0 | w̄⟨z⟩.0)` tells you and `go func(){...}()` doesn't.

### `-race` isn't going to save you here

I should be straight about this, because it would be convenient if it weren't true.

The race detector catches shared *mutable* state. Passing channels about is neither shared nor mutable, so a captured `z` sails through, and so would a captured anything else. The tool has nothing to complain about, because there's no bug - there's a modelling failure, and no compiler in the world checks whether your program matches the thing you drew on the whiteboard.

So this one is a discipline. Keep it and the shape of your Go is the shape of your term. Don't, and the term is decoration.

---

## Exercises

1. **Shuffle it yourself.** Reduce `a(u).0 | x̄⟨y⟩.0 | x(z).z̄⟨u⟩.0` to a normal form, writing out every `≡` step as well as the `→`. Where did you have to use associativity as well as commutativity?

2. **Precedence.** What's the difference between `x(z).ȳ⟨z⟩.0 | w̄⟨z⟩.0` and `x(z).(ȳ⟨z⟩.0 | w̄⟨z⟩.0)`? One of them has a free `z` in it. Which, and why is that a bug?

3. **Not congruent.** Give two terms that reduce to the same thing but are not `≡`. Then give two that *are* `≡` but look completely different at a glance.

4. **Sneaky in the wild.** Rewrite your `Tell` in the `TellSneaky` style, then again with an anonymous function that takes its names as arguments. All three pass. Line up the three spawn sites and say what each one tells a reader.

5. **Three-way.** Write `x(z).(ȳ⟨z⟩.0 | w̄⟨z⟩.0 | q̄⟨z⟩.0)` and a test for it. Does the order the three outputs complete in matter? Should your test care?

---

## Recap

- `P | Q` is a bag, not a list. There is no first.
- **Structural congruence** `≡` says which ways of writing a term count as the same term: commutative, associative, `0` is the unit.
- Reduction is defined *up to* `≡` - shuffle, reduce, shuffle back. That's what makes the rule usable.
- `P | 0 ≡ P`, which is why we may cross out finished processes.
- `≡` is not "behaves the same". That's a much harder question and it's chapter 10.
- The dot binds tighter than the bar, so `x(z).(P | Q)` needs its brackets.
- No shared state, ever - connectivity is the whole content of a term.
- Say what a spawned process gets, at the spawn site. `-race` will not check this for you.

## Next

Chapter 4: `ν`. We've been making names with `make(Name)` since the first page and I've said nothing about what that actually means. It turns out to be the most interesting thing in the calculus, and Go gives it to us almost for free.
