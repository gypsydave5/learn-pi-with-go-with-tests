# Chapter 5: The scope moved

## What we're covering

The third law of `ν`, which I've been withholding for a chapter and a half. It's the one that makes the π-calculus the π-calculus, and it's the answer to the question chapter 4 finished on.

## Setup

```
mkdir the-scope-moved && cd the-scope-moved
```

`Send` and `Recv`, as ever.

---

## Where we got to

Two things happened at the end of chapter 4, and they don't obviously fit together.

First: a private name that never leaves home might as well not exist. `(νs)( s̄⟨y⟩.0 | s(z).w̄⟨z⟩.0 )` does exactly what `w̄⟨y⟩.0` does, and no observation can tell them apart.

Then: `Mint` sent one out anyway.

```
(νs)x̄⟨s⟩.0
```

Make a fresh name, hand it to whoever's listening on `x`. And now the test is holding a name that was supposed to be private to `Mint`, and can send on it, receive on it, and pass it to anybody it likes.

So either `ν` doesn't mean what I said, or something happened to the scope.

---

## The scope moved

Here's the law.

```
(νz)P | Q  ≡  (νz)(P | Q)      provided z does not occur free in Q
```

Read it right to left first, because that direction looks harmless: if you've got a private `z` covering two processes, and one of them doesn't use `z` at all, you may as well shrink the boundary to cover only the one that does.

Now read it left to right, which is the same law and considerably more alarming. `Q` was outside the restriction. Now it's inside. The boundary has *grown* to swallow a process that was previously beyond it.

That's it. That's the whole difference between this and every calculus that came before.

---

## The side condition is the entire trick

`provided z does not occur free in Q`. Don't skip it.

And read it in chapter 2's terms, because that's what makes it obvious: **`Q` doesn't take a `z` as a parameter.** `Q` might well have a `z` of its own somewhere inside it - a local, bound by its own input prefix - and that's none of our business. What it mustn't have is a `z` that somebody outside is expected to supply, because that's a different `z` and we're about to pretend it's ours.

Suppose `Q` already had a free `z` of its own - some `z` from somewhere else entirely, that just happens to be spelled the same. Pull the boundary over it and that `z` gets captured, exactly as in chapter 2. Two unrelated names mashed into one because of a spelling coincidence.

And when `z` does occur free in `Q`? Alpha-convert. `z` is bound, we can rename it to anything we like:

```
(νz)P | Q      where z occurs free in Q
≡ (νt)P{t/z} | Q      rename the bound one out of the way
≡ (νt)(P{t/z} | Q)    now the side condition holds
```

Which is why the side condition costs us nothing. It's never a wall, only a detour.

But notice what it means. **Before the law can be applied, `Q` provably cannot name `z`.** That's not a technicality we work around - it's the guarantee. We're only ever allowed to extend a scope over a process that had no way of referring to the name in the first place.

---

## Watching it happen

Let's do `Mint` properly, with somebody on the other end.

```
(νs)x̄⟨s⟩.0  |  x(y).ȳ⟨n⟩.0
```

On the left, mint `s` and send it on `x`. On the right, receive on `x` and send `n` on whatever turned up.

The right-hand process has no `s` in it. It couldn't have - `s` didn't exist when we wrote it down. So the side condition holds, and:

```
(νs)x̄⟨s⟩.0 | x(y).ȳ⟨n⟩.0
  ≡  (νs)( x̄⟨s⟩.0 | x(y).ȳ⟨n⟩.0 )     scope extension
  →  (νs)( 0 | s̄⟨n⟩.0 )                communicate on x, {s/y}
  ≡  (νs) s̄⟨n⟩.0                       tidy up
```

Look at the last line. There's a process sending on `s`, and it is *inside* the restriction, and it was never written there. We put it there by moving the boundary.

The substitution did the work. `{s/y}` replaced the placeholder `y` with the actual private name, and the process that received it can now do things it demonstrably could not do one line earlier.

---

## Two words for it

The law is called **scope extension**. That's the rearrangement - a `≡`, not a `→`, because nothing has run.

What happens when a bound name is sent out of its restriction is called **scope extrusion**. That's the phenomenon; extension is the rule that describes it.

I'll try to keep them straight and will certainly fail at some point.

---

## It grew. It didn't break.

The obvious worry: haven't we just leaked the private name? What was the point of `ν` if any output can smuggle it out?

No. Count who can use `s` after all that. Exactly the processes inside the boundary - which is now two processes instead of one, because one of them was *handed* the name.

Nobody broke in. There's no third process anywhere in the term that can name `s`, and there is no way for one to acquire it except by being given it, by somebody who already had it. The set of processes that can use a restricted name only ever grows by explicit gift.

Which is a rather good description of a key. Or a session token. Or a signed URL. Or a capability, if you like that word - an unforgeable reference whose possession *is* the authority to use it.[^caps]

[^caps]: This is not a coincidence and not an analogy people invented afterwards. The capability-security literature and the π-calculus have been borrowing from each other for decades, and if you've ever argued that ambient authority is a design error, you've been arguing for `ν`.

---

## The thing CCS couldn't do

Chapter 1, if you'll cast your mind back: CCS handled concurrency perfectly well, but the shape of the system was fixed when you wrote the term down. You could hide names, rename them, compose processes - but the graph of who-can-talk-to-whom never changed.

Here it changed. Mid-reduction. Because a name moved.

That's **mobility**, that's the entire reason there's a π-calculus as well as a CCS, and it's why Milner's book is called *Communicating and Mobile Systems*.

And it's why `Name` is defined the way it is:

```go
type Name chan Name
```

We've been carrying that recursion since page one and this is the chapter where it earns its keep. A channel that carries channels is the only way connectivity can change at runtime. Take the recursion away - make it `chan struct{}`, `chan int`, anything that isn't itself a channel - and you've got CCS. A fixed diagram. Wires soldered in place.

---

## Your turn

Right. Here's a process that mints a name, hands it out, and then listens on it:

```
(νs)( x̄⟨s⟩.0 | s(z).w̄⟨z⟩.0 )
```

Make a private `s`. Then side by side: send `s` on `x`, and receive on `s` and forward whatever arrives to `w`.

And here's the test, which is doing something we haven't done before - **the observer is the one receiving the private name**:

```go
func TestScopeExtrusion(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		x := make(Name)
		w := make(Name)
		n := make(Name)

		go Gate(x, w)

		synctest.Wait()

		select {
		case <-w:
			t.Fatal("something came back before anyone had s")
		default:
		}

		s := <-x // and now we have it

		s <- n

		if got := <-w; got != n {
			t.Errorf("got %v, want %v", got, n)
		}
	})
}
```

What it wants:

```go
// Gate makes a private name, hands it out on x, and reports whatever
// comes back on it to w.
//
//	(νs)( x̄⟨s⟩.0 | s(z).w̄⟨z⟩.0 )
func Gate(x Name, w Name)
```

Four lines again.

Now look hard at that test, and specifically at this:

```go
s := <-x
s <- n
```

Before the first of those lines there is no `s`. Not "an `s` we're choosing not to use" - the identifier does not exist, and there is no expression you could write that would produce the channel it's about to hold. `Gate` made it, in there, and nothing outside can reach into a function and take a local.

After that line we can send on it.

**That is scope extrusion, and it is the difference between two lines of Go.**

---

## What we can and can't test now

Chapter 4 ended on a deflating note: privacy is unobservable, so no test can catch a cheat that skips the `ν`.

This chapter is the other side of it. Try cheating `Gate` - hand out `w` instead of minting something:

```go
func Gate(x Name, w Name) {
	x <- w
}
```

The test deadlocks. We send `n` on what we received, and nobody is listening on `w` any more, because `w` is the thing we were supposed to be listening *to*. There's no forwarder. The bubble notices and fails.

So the cheat doesn't work here, and it's worth being precise about why. It isn't that the test detected a missing `ν`. **The test detected mobility.** A name travelled, and afterwards something could interact that couldn't interact before, and if you don't build that you don't pass.

Freshness is still unobservable. If `Gate` handed out some other channel it had lying about, and forwarded properly, the test would be perfectly happy. As ever, that part is discipline.

---

## Exercises

1. **On paper.** Reduce `(νs)( x̄⟨s⟩.0 | s(z).w̄⟨z⟩.0 ) | x(y).ȳ⟨n⟩.0` all the way, showing every `≡` and every `→`. How many times did you use scope extension?

2. **Capture.** Try to apply scope extension to `(νs)s̄⟨y⟩.0 | s(z).w̄⟨z⟩.0`, where that second `s` is free and has nothing to do with the first. What goes wrong? Fix it, and say what the fix cost.

3. **Shrink it.** The law works right-to-left too. Take `(νs)( s̄⟨y⟩.0 | w̄⟨n⟩.0 )` and pull the second process *out* of the restriction. Why are you allowed to?

4. **Two callers.** Give `Gate` two customers, each of which gets its own fresh name. Write it, and write a test that would fail if both got the same name. What did you have to do to make the failure visible?

5. **Soldered in place.** Change `type Name chan Name` to `type Name chan struct{}` and try to write `Gate`. Get as far as you can. Describe precisely where you stop.

---

## Recap

- **Scope extension**: `(νz)P | Q ≡ (νz)(P | Q)` when `z` doesn't occur free in `Q` - that is, when `Q` doesn't take a `z` as a parameter. The boundary can grow.
- The side condition isn't red tape - it's the guarantee. You can only extend a scope over a process that couldn't have named the thing anyway.
- When it's in the way, alpha-convert. It's never a wall.
- **Extrusion** is what happens when a restricted name is sent out; **extension** is the law that says it's allowed.
- Nobody breaks in. The set of processes that can use a private name grows only by gift. That's a capability.
- Connectivity changes at runtime. This is **mobility**, it's what CCS couldn't do, and it's the reason `Name` is a channel of `Name`.
- Privacy still isn't testable. Mobility is.

## Next

Chapter 6: `!`. Everything we've written so far does its thing once and dies. Real systems don't, and it turns out there are two quite different ways to make a process go round again - one of which is a worker and one of which is a server, and telling them apart is most of the chapter.
