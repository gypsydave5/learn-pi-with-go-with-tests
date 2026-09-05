# Chapter 4: Where names come from

## What we're covering

Restriction - written `(νz)P`, where that's a Greek **nu** and not a wonky `v`. The other binder.

We've been calling `make(Name)` since the first page of chapter 1 and I have said absolutely nothing about what it means. Time to fix that, because it turns out to be the most interesting thing in the language.

## Setup

```
mkdir where-names-come-from && cd where-names-come-from
```

`Send` and `Recv` again.

---

## Everything so far has been somebody else's problem

Look back at every process we've written. `Send`, `Recv`, `Store`, `Tell`. Every single name in them is a parameter - free, in chapter 2's terms, which means *the world has to supply it*.

```
x(z).ȳ⟨z⟩.0        needs an x. needs a y.
```

Fine. But somebody, somewhere, has to actually make one. You can't have a system built entirely of processes waiting for names that nobody ever created. The buck has to stop.

So here's the buck:

```
(νz)P
```

Read it as "new `z`, in `P`". It creates a name `z`, brand new, unlike any name that has ever existed or ever will, and `z` is known only inside `P`.

That Greek letter is a nu, pronounced "new", which is either a superb pun or a coincidence that everybody has decided to enjoy.[^nu] Some languages built on the calculus write it out longhand - Pict says `new x in P` - and honestly that's how you should read it every time.

[^nu]: I've never found Milner saying he picked ν because it sounds like "new", and it may just be that ν was the next Greek letter going spare near λ and μ. But the reading is too good to give up, and it's exactly right about the behaviour, so everyone uses it.

And in Go?

```go
func P() {
	z := make(Name)
	// z is alive in here
}
```

That's it. `make(Name)` is `ν`. You have been writing restriction since page one.

---

## Made, not received

We've now got two binders, and it's worth being clear about how they differ, because they look alike and they aren't.

Both create a local name. In Go, both are the same shape:

```go
func P(x Name) {
	z := <-x         // bound by the input prefix
	s := make(Name)  // bound by ν
}
```

`z` and `s` are both locals. Neither is anybody else's business, both can be renamed freely, both vanish at the closing brace.

The difference is where the thing came from. **`z` was received. `s` was made.**

Which has a consequence you can see in the rules. An input prefix is *waiting* - `x(z).P` sits there until something arrives, and when it does, the arriving name is substituted for `z` throughout `P` and the prefix is gone. It got what it was waiting for and its job is done.

`(νs)P` isn't waiting for anything. There is no rule anywhere in the π-calculus that eliminates a `ν` by handing it a value, because it hasn't asked for one. Nothing is ever substituted for a restricted name.

So what *does* happen to a `ν`? Almost nothing. It sits there being a boundary, for the whole life of the process. The only interesting thing it can ever do is move, and that's the next chapter.

---

## Unforgeable

Here's the property that makes `ν` worth having, and Go hands it to us free.

**There is no way to construct a reference to a channel you weren't given.** None. There's no `chan.FromAddress(0x...)`, no literal syntax for a specific channel, no way to guess one. Channels come from `make`, or from somebody handing you one, and that's the complete list.

Which means `(νs)P` really does mean what it says. The set of processes that can use `s` is exactly: the ones inside `P`, plus anyone they subsequently hand it to. Nobody can sneak in from outside, because there's nothing to sneak in *with*.

This is rarer than it sounds. Most languages leak an escape hatch - reflection, a global registry, pointer arithmetic, some debug interface. Go's channels genuinely don't.

Except, of course, that you can always cheat:

```go
var everyChannelIveEverMade = map[string]Name{}   // don't
```

Stash a name somewhere global and any process in the program can fish it out, and your `ν` means nothing at all. Which is chapter 3's rule again, from a different angle: **no shared state, no globals, names travel by being passed.** The reason restriction works in our Go is that we've agreed not to break it.

The compiler won't stop you. This one's on us.

---

## Two more laws

`≡` grows. Two new lines:

```
(νz)(νw)P  ≡  (νw)(νz)P      make them in either order
(νz)0      ≡  0              a private name nobody's using is nothing
```

And `ν` slots into chapter 2's free-names table exactly where you'd expect:

```
fn((νz)P) = fn(P) \ {z}
```

Same subtraction as the input prefix. `z` isn't the world's problem any more, because we made it ourselves.

There's a third law for `ν`, and it's the important one, and I'm not going to show it to you yet.

---

## Fresh means fresh

Two restrictions of the same letter are not the same name:

```
(νs)P  |  (νs)Q
```

Those are two different `s`es. They just happen to be spelled alike, and since bound names can be renamed at will - alpha-conversion, chapter 2, applies to `ν` exactly as it does to input - we could write it as:

```
(νs)P  |  (νt)Q{t/s}
```

and nothing has changed. If that feels obvious, it's because Go has been doing it for you forever:

```go
func Thing() {
	s := make(Name)
}
```

Call `Thing()` twice and you get two channels. Same source code, same variable name, two distinct things. Nobody has ever been confused by this.

---

## Your turn

A process that keeps itself to itself:

```
(νs)( s̄⟨y⟩.0 | s(z).w̄⟨z⟩.0 )
```

Make a private name `s`. Then, side by side: send `y` on `s`, and receive on `s` and forward whatever turns up to `w`.

So `y` gets from one half of this process to the other by travelling down a channel that nothing else in the universe can touch. The only thing anyone outside can see is something coming out of `w`.

The test:

```go
func TestPrivateChannel(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		y := make(Name)
		w := make(Name)

		go Private(y, w)

		if got := <-w; got != y {
			t.Errorf("got %v, want %v", got, y)
		}
	})
}
```

What it wants:

```go
// Private makes a name of its own, uses it to pass y from one half
// of itself to the other, and reports the result on w.
//
//	(νs)( s̄⟨y⟩.0 | s(z).w̄⟨z⟩.0 )
func Private(y Name, w Name)
```

Four lines. Go on.

### Now cheat

Done? Good. Now delete it and write this instead:

```go
func Private(y Name, w Name) {
	w <- y
}
```

That also passes. Obviously it does.

And before you decide I've wasted your afternoon: **the cheat isn't wrong.** It isn't a shortcut that happens to fool a weak test. The two are the same process, and we can very nearly prove it:

```
(νs)( s̄⟨y⟩.0 | s(z).w̄⟨z⟩.0 )
  →   (νs)( 0 | w̄⟨y⟩.0 )              the private communication happens
  ≡   (νs) w̄⟨y⟩.0                     drop the 0, chapter 3
```

And now we're looking at a `ν` wrapped around a process that doesn't mention `s` anywhere. It's plainly redundant. You can see it's redundant. But I can't cross it out for you, because the law that lets me is the one I said I was keeping back.

So here's the real lesson of this chapter, and I'm afraid it's a bit deflating:

**A private name that never leaves home might as well not exist.**

You cannot write a test that catches the cheat, because there is nothing to catch. There's no observation anybody outside `Private` could make that distinguishes them. `s` is unobservable by construction - that's what restriction *means* - and a thing nobody can observe cannot be the difference between two programs.

Which rather raises the question of what `ν` is *for*.

---

## And now a problem

One more, and this one I'll write, because it's the setup for the next chapter.

```
(νs)x̄⟨s⟩.0
```

Make a fresh name, and send it on `x`.

```go
// Mint makes a new name and hands it out on x.
//
//	(νs)x̄⟨s⟩.0
func Mint(x Name) {
	s := make(Name)
	x <- s
}
```

```go
func TestFreshEveryTime(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		x := make(Name)

		go Mint(x)
		go Mint(x)

		first := <-x
		second := <-x

		if first == second {
			t.Error("two νs made the same name")
		}
	})
}
```

It passes. Two `ν`s, two names, and the test never has to trust anybody about it - it just holds them both up and compares.

But look at what we did.

`s` was private to `Mint`. That was the whole point of `ν`. And then `Mint` sent it to the test, and now the test has got it, and can send on it, and receive on it, and pass it to anybody it fancies.

So either `ν` doesn't mean what I said it means, or something rather interesting happens to a scope when a private name walks out of it.

---

## Exercises

1. **Count them.** What are the free names of `(νs)( s̄⟨y⟩.0 | s(z).w̄⟨z⟩.0 )`? What would `Private`'s Go signature have to be if you got it wrong?

2. **Pointless restriction.** What does `(νs)ȳ⟨w⟩.0` do? Is it different from `ȳ⟨w⟩.0`? Write both in Go and see whether you can tell them apart from outside.

3. **Rename it.** Alpha-convert `(νs)( s̄⟨y⟩.0 | s(z).w̄⟨z⟩.0 )` so the private name is called `q`. Now do it so the private name is called `y`. Something goes wrong - what, and what's the rule that stops you?

4. **Two privates.** Write `(νs)(νt)( s̄⟨y⟩.0 | s(z).t̄⟨z⟩.0 | t(q).w̄⟨q⟩.0 )` and a test. How many `make`s? How many goroutines? Does the order of the two `ν`s matter, and which law says so?

5. **Break it on purpose.** Rewrite `Private` so that `s` is a package-level variable instead of a local. The test will still pass. Say precisely what you've destroyed.

---

## Recap

- `(νz)P` makes a brand new name `z`, known only inside `P`. Read it "new z, in P".
- In Go it's `make(Name)`. You've been doing it all along.
- Both binders make locals. The difference is that an input prefix's name is **received** and a `ν`'s name is **made** - so nothing is ever substituted for a restricted name, and a `ν` is never discharged.
- Go channels are **unforgeable** - you cannot construct one you weren't given - which is what makes restriction mean anything.
- That guarantee survives only as long as you keep the no-globals rule. The compiler won't help.
- `(νz)(νw)P ≡ (νw)(νz)P`, `(νz)0 ≡ 0`, and `fn((νz)P) = fn(P) \ {z}`.
- Fresh means fresh: two `ν`s of the same letter are two different names.

## Next

Chapter 5, and the question `Mint` just left us with. A private name has escaped its scope. Either that's a hole in the calculus, or the scope moved - and the answer is the reason the π-calculus exists at all.
