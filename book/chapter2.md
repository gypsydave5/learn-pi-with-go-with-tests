# Chapter 2: And then what?

## What we're covering

The dot. We met it in chapter 1 and I told you it meant "and then", which is true and isn't the interesting part.

The interesting part is that `x(z).P` doesn't just say what happens next. It says *where `z` means anything at all*.

## Setup

```
mkdir and-then && cd and-then
```

Same two files as before. You'll want `Send` and `Recv` again - copy them over, we're not precious about it.

---

## Three jobs

Here's the input prefix again:

```
x(z).P
```

The `P` on the end is doing three separate jobs, and I only told you about one of them.

**It sequences things.** There's no `;` in the π-calculus, so if you want `P` to happen after the input, the only way to say so is to put it after the dot. Chapter 1, covered.

**It stops the process being dead.** `x(z).0` receives one thing and then stops forever. If every input had to end in `0` you'd have a calculus in which nothing could do two things, which would be a short book.

**It's the scope of `z`.** And this is the one.

`x(z).P` doesn't mean "receive something and call it `z`, then do `P`". It means "receive something, call it `z`, **and `z` means that thing exactly as far as the end of `P` and not one character further**".

`z` doesn't exist outside `P`. It isn't a name the term has; it's a name `P` has. The input prefix *created* it.

You already believe all of this, by the way, because you write Go. You already know what lexical scope is, even if nobody's ever made you call it that:

```go
func Recv(x Name, y Name) {
	z := <-x   // z is born here
	Send(y, z) // and it's alive in here
}              // and it's gone
```

`z := <-x` is the binder. Everything from that line to the closing brace is `P`. Go calls it "the scope of the variable", the π-calculus calls it "the continuation", and they're the same idea wearing different hats.[^cps]

[^cps]: Not to be confused with the "continuation" in continuation-passing style, which is a *value* - a thing you build and hand to somebody so they can decide when to invoke it. Ours is just syntax: whatever happens to be written after the dot. You can't pass it anywhere, because it isn't a thing. The two do meet in chapter 9, where encoding the λ-calculus involves sending a channel that says where to put the answer and then using it in the prefix continuation. Until then, treat them as different ideas that got landed with the same word.

---

## What sets it off

One more thing about the binder before we move on, and it's the bit with no equivalent in ordinary code.

A local in Go gets its value the moment the line runs. `z := <-x` is a bit different - it gets its value when *somebody else* does something. The binder sits there, loaded, and what sets it off is another process making an output on `x`.

```
x̄⟨y⟩.Q  |  x(z).P   →   Q  |  P{y/z}
```

So `z` isn't filled in by this process. It's filled in by a negotiation between two of them, and neither one can do it alone. That's the only way a name ever gets into a continuation.

---

## Parameters and locals

If a binder makes a name local, then names come in two flavours. And you already know both, because you write functions.

```go
func Recv(x Name, y Name) { // parameters
	z := <-x                // a local
	Send(y, z)
}
```

`x` and `y` come from outside. The function can't do anything until somebody supplies them, and it doesn't get to decide what they are. `z` is the function's own business - it made it, it can call it whatever it likes, and nobody outside knows or cares.

Exactly the same split in a term:

```
x(z).ȳ⟨z⟩.0
```

`z` is a local: `x(z)` made it. `x` and `y` are parameters: nothing here says where they came from, so somebody has to hand them over.

**Every term is a function signature you can read off the page.** The names it doesn't explain are the ones it needs supplied.

### The jargon

Now the words, which you need because everybody else uses them and which are, I'm sorry to say, terrible.

A local is called a **bound** name. That one's fine - it's tied to a binder, and you can point at the knot.

A parameter is called a **free** name. And that is a menace, because "free" sounds like it should mean "absent" or "unused", and it means nothing of the kind.

It's free as in *dangling*. Loose. Nothing round here has hold of it. Think of a rope: a bound name is tied to a post you can point at, and a free name has a loose end that runs off the edge of the page, where somebody else is holding it.

So a name can appear six times in a term and not be free once - `x(z).z̄⟨y⟩.0` mentions `z` three times and it's tied down every single one.

### Where you're looking matters

Second trap, and it's the one that'll actually bite you. **Free and bound aren't properties of a name. They're properties of a name relative to whichever term you've got your eye on.**

Take that same term and look at the `z` in the output. Relative to just the bit after the first dot - `z̄⟨y⟩.0` on its own - it's dangling. Relative to the whole term, it's tied down, because `x(z)` has hold of it. Same character, same spot on the page, two correct answers.

You always have to say *free in what*, which is why the careful phrasing is "`z` **occurs free in** `P`" rather than "`z` is free in `P`".

If that sounds academic, it's the thing you rely on every time you write a closure. The captured variable is free in the function body and bound in the enclosing scope - one variable, two answers, depending on where you draw the box. Programmers say "captured" rather than "free" precisely because what matters is the relationship between the two boxes, and that word quietly hides the fact that the answer depends on which one you're standing in.

### Only one binder so far

There is exactly one binder in the π-calculus at this point, and it's the input prefix.[^binders] Output doesn't bind. Sending a name doesn't create it - you had to have it already.

[^binders]: There's one more coming in chapter 4 - `(νz)P`, which also binds `z` in `P`, but for entirely different reasons. Input binds a name it *received*; `ν` binds a name it *invented*. Two binders, and that's the lot.

Which is where the relativity bites again. On its own:

```
x̄⟨y⟩.0            x dangling, y dangling
```

Nothing has hold of either, so both must come from outside. Now put an input to the left:

```
x(y).x̄⟨y⟩.0       x dangling, y tied down
```

Identical output prefix, and now that `y` is bound - because it's sitting *inside* the body of a binder that made a `y`. The `x` is still free, because nothing made an `x`.

Same question as always: parameter, or local?

Written out properly, the free names of a term are:

```
fn(0)         = {}
fn(x̄⟨y⟩.P)    = {x, y} ∪ fn(P)
fn(x(z).P)    = {x} ∪ (fn(P) \ {z})
fn(P | Q)     = fn(P) ∪ fn(Q)
```

Look at the third line. `x` is free, because you need a channel to receive on. But `z` is *subtracted* from the free names of `P` - whatever `P` wanted a `z` for, the binder has seen to it, so it isn't something the world has to provide any more.

---

## Now the pointless one

What if we leave `P` off?

```
x(z).0
```

Receive something, call it `z`, and then... nothing. `z` is bound in `0`, and `0` has no names in it, so binding it achieved precisely nothing.

Go, magnificently, will not let you write this:

```go
func Sink(x Name) {
	z := <-x
}
```

```
./name.go:2:2: declared and not used: z
```

The compiler has spotted that you've bound a name with nowhere to use it and told you to stop wasting everyone's time. Which is a stricter rule than the π-calculus has - `x(z).0` is a legal term - and the compiler is, once again, right.

To write it you have to drop the binding altogether:

```go
func Sink(x Name) {
	<-x
}
```

And here's the thing: **that process is not useless.** It's a sink. It consumes. It can pair with an output and make it disappear, and a system containing one behaves differently from a system without one.

You've seen it before. In chapter 1, the losing input in our race ended up as a process that could never receive anything - but the *winner* was, in effect, one of these. Something that takes a message out of the world.

So the continuation isn't what makes a process worth having. It's what makes the *binding* worth having.

---

## Alpha-conversion, which you have believed since forever

Two processes:

```
x(z).ȳ⟨z⟩.0
x(w).ȳ⟨w⟩.0
```

Same process. Obviously the same process. The bound name is internal bookkeeping, and renaming it consistently changes nothing that anybody outside could ever detect.

This is called **alpha-conversion**, it has a Greek letter and a formal definition, and you have believed it unquestioningly since the first time you renamed a variable. Here it is in Go:

```go
func Recv(x Name, y Name) {
	z := <-x
	Send(y, z)
}

func Recv(x Name, y Name) {
	whatever := <-x
	Send(y, whatever)
}
```

Nobody thinks those are different functions. Congratulations, you're an alpha-convert.

What you *cannot* do is rename a free name. `x(z).ȳ⟨z⟩.0` and `x(z).b̄⟨z⟩.0` are entirely different processes, because `y` and `b` are the world's business, not the term's. Renaming a parameter changes what you're asking for.

---

## Why any of this matters: capture

Here's the payoff, and it's the reason chapter 1 had that mealy-mouthed footnote about "free occurrences".

Take this term:

```
x(z).z̄⟨w⟩.0
```

Receive on `x`, call it `z`, then send the free name `w` on `z`. Now suppose something substitutes `z` for `w` - `{z/w}`, remember, new over old.

Do it naively and you get:

```
x(z).z̄⟨z⟩.0
```

Which is a disaster. The `z` that arrived from outside has been swallowed by the binder. It's now pretending to be the local one. Two completely different names have been mashed into one because they happened to be spelled the same.

This is **capture**, and the fix is to rename the bound name first - which we're allowed to do, because alpha-conversion:

```
x(q).q̄⟨w⟩.0        rename the bound name
x(q).q̄⟨z⟩.0        now substitute {z/w}
```

Correct. The local is still local, the arrival is still an arrival, and nobody got confused.

So substitution replaces **free** occurrences only, and alpha-converts out of the way when a binder would capture. That's the whole rule, and it's why chapter 1 couldn't quite tell you the truth in one line.

Go, incidentally, has the same problem and solves it with shadowing:

```go
func Confusing(x Name, z Name) {
	{
		z := <-x  // a different z entirely
		_ = z
	}
	// out here, z is still the parameter
}
```

The inner `z` and the outer `z` are unrelated. The compiler tracks which is which by scope, not by spelling, and that's exactly what "free occurrences only" is asking you to do by hand.

---

## Your turn

Time to write a process that actually *needs* its continuation.

Here's the term:

```
x(z).y(w).z̄⟨w⟩.0
```

Say it out loud first. Then together: receive a name on `x` and call it `z`. Then receive a name on `y` and call it `w`. Then send `w` on `z`, and stop.

Notice what that requires. `z` arrived first, and it isn't used until after the second input. So it has to *survive* an interaction - it has to still mean something further down the chain. That's the continuation doing the job that isn't sequencing.

It's also a shape you've seen: something hands you a channel to reply on, then later something hands you the thing to reply with.

Here's the test.

```go
func TestStore(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		x := make(Name)
		y := make(Name)
		reply := make(Name)
		payload := make(Name)

		go Store(x, y)

		x <- reply // here's where to answer

		synctest.Wait()

		select {
		case <-reply:
			t.Fatal("Store answered before it had anything to say")
		default:
		}

		y <- payload // and here's what to say

		if got := <-reply; got != payload {
			t.Errorf("got %v, want %v", got, payload)
		}
	})
}
```

And here's what it wants:

```go
// Store takes a name on x and holds onto it, then takes a name on y
// and sends that second name on the first one.
//
// Which in π looks like:
//
//	x(z).y(w).z̄⟨w⟩.0
func Store(x Name, y Name)
```

Three lines. Off you go.

Note the middle bit of that test, which is the chapter 1 pattern again: `Wait` for everything that can happen to have happened, then show nothing came out of `reply`. That's what proves `z` was being *held* rather than used immediately.

---

## Exercises

1. **The first one, not the second.** Write `x(z).y(w).z̄⟨z⟩.0` in Go - it sends the first name it received, on itself. Does it compile? What does the compiler say about `w`, and what does that tell you about binders with unused bodies?

2. **Shadowing.** What does `x(z).x(z).ȳ⟨z⟩.0` do? Which `z` gets sent? Now write it in Go. You'll find Go makes you say it differently - work out what the difference is, and whether the two terms end up behaving the same anyway.

3. **Count the free names.** For each of these, list the free names and the bound names:
   - `x̄⟨y⟩.0`
   - `x(z).z̄⟨y⟩.0`
   - `x(z).z̄⟨z⟩.0`
   - `x(x).x̄⟨x⟩.0`

   That last one is legal and horrible. Say what it does.

4. **Capture, by hand.** Apply `{z/w}` to `x(z).y(w).z̄⟨w⟩.0`. Do it carelessly first and write down what you get. Then do it properly. What went wrong the first time, and which occurrence of `z` was the liar?

5. **Alpha-convert `Store`.** Rewrite your Go `Store` with every local renamed. Confirm the test still passes. Then rename a *parameter* and watch the test fail to compile - and be clear in your own head about why those two renamings are different.

---

## Recap

If you take three things out of this chapter, take these:

- **Bound names can be renamed.** A binder in the term made them, so the term owns them, and nobody outside is relying on the spelling.
- **Free names can't.** They're the interface. Somebody else supplies them, and renaming one changes what you're asking for.
- **Which is which depends on where you're looking.** "Free" means not caught by any binder between here and the edge of whatever term you've got your eye on. Move the edge, change the answer.

And the thing those three give you for nothing: **when two names share a letter and you can't tell whether they're the same name, rename the bound one.** If they were different, the term suddenly reads. If they were the same, you'll find you can't rename it - and that's your answer too.

The rest:

- The dot does three things: it sequences, it keeps the process alive, and - the one that matters - it delimits the scope of the name the input just bound.
- `x(z).P` binds `z` in `P`. `P` is a body, not a name - fill it in and the letter disappears.
- **Free names are parameters. Bound names are locals.** A term's free names are what the world has to supply.
- Input is the only binder we have so far. Output binds nothing.
- Substitution replaces *free* occurrences only, and alpha-converts out of the way to avoid capture.
- Go's compiler enforces a stricter version of all of this than the calculus does, and is right to.
- The letter is not the name. Two `s`es are only the same name if the same binder made them both, and you cannot tell by looking at the letter.

## Next

Chapter 3: parallel composition. We've been writing `|` since chapter 1 and calling it "at the same time", which is going to turn out to be about a third of the story. We'll find out why `P | Q` and `Q | P` have to be the same thing, why we're allowed to cross out those `0`s I've been leaving lying around, and what `go test -race` has to say about processes that reach into scopes they weren't given.
