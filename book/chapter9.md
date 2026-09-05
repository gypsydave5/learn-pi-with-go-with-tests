# Chapter 9: Functions are processes too

## What we're covering

We built numbers out of processes. Now we build functions.

Specifically we're going to compile the entire λ-calculus into the π-calculus, which is Milner's result from 1992 and the moment this stops looking like a nice model of concurrency and starts looking like a foundation.

And the trick turns out to be the reply channel from chapter 1.

## Setup

```
mkdir functions && cd functions
```

---

## The λ-calculus, in one page

I've been careful not to require this, so here it is from scratch. It's smaller than anything else we've introduced.

Three ways to build a term:

```
M, N ::= z            a variable
       | λz.M         a function: takes z, returns M
       | M N          apply M to N
```

That's the whole grammar. No numbers, no booleans - same reductive move we've been making all book, different primitive.

One rule:

```
(λz.M) N   →   M{N/z}
```

Called **β-reduction**. A function meets an argument; substitute the argument for the parameter throughout the body. You've been doing this substitution since chapter 1 - it's the same `{new/old}`, and it has the same business with free occurrences and capture that chapter 2 went through.

And that's it. That's the λ-calculus. Anything computable can be written in those three lines, which is either wonderful or infuriating depending on the hour.

---

## The problem

Functions return values. Processes don't return anything at all - chapter 1, rule two, no return values ever.

So how do you compile something whose entire purpose is to hand a value back, into something that structurally cannot hand anything back?

You know the answer. You've known it since chapter 1.

```go
type Request struct {
	Data  string
	Reply chan Response
}
```

**You pass in somewhere to put the answer.**

Every λ-term gets compiled *relative to a result channel* - a name that says where to deliver. Written:

```
⟦M⟧u
```

Read: "the translation of `M`, reporting to `u`". The brackets are there because there are two languages in play and we'd be lost without them - inside is λ, outside is π.

It's a compiler. Three equations, one per case of the grammar, structural recursion over the syntax. You've written this function before; it usually emits something less exotic.

---

## The three rules

```
⟦z⟧u      =  z̄⟨u⟩.0
⟦λz.M⟧u   =  u(z,q). ⟦M⟧q
⟦M N⟧u    =  (νq)( ⟦M⟧q | (νy)( q̄⟨y,u⟩.0 | !y(r).⟦N⟧r ) )
```

Take them one at a time.

**A variable** `⟦z⟧u = z̄⟨u⟩.0`. Send the result channel *to the variable*. That looks backwards until you realise what a variable is here: it's a name somebody will hand you a value on. So you don't get the value out of `z`, you tell `z` where you want it delivered. A variable is a place you can put a request.

**A function** `⟦λz.M⟧u = u(z,q).⟦M⟧q`. A function is a process sitting on its result channel, waiting to be handed two things: a channel standing for the argument, and a *new* result channel for the body to report to. When they arrive, it becomes the body.

So a λ isn't a value that gets returned. It's a process that waits to be called, and `u` is the phone number.

**An application** is where the work is:

```
⟦M N⟧u  =  (νq)( ⟦M⟧q | (νy)( q̄⟨y,u⟩.0 | !y(r).⟦N⟧r ) )
```

Three moving parts. Run `M` on a private channel `q`, so we know where to reach it. Mint a private `y` to stand for the argument. Then send `M` the pair `⟨y,u⟩` - here's your argument, and here's where to put the answer.

And alongside that, `!y(r).⟦N⟧r` - a **server for the argument**. Anyone who wants `N` sends a result channel to `y`, and gets `N` compiled to report there.

Look at what all those `ν`s are doing. `q` is private so nobody else can call this instance of `M`. `y` is private so the argument belongs to this application and no other. Chapter 5's capabilities, doing exactly the job they were introduced for.

---

## Watch it happen

Let's do the identity function applied to a free variable: `(λz.z) w`.

```
⟦(λz.z) w⟧u
  =  (νq)( ⟦λz.z⟧q | (νy)( q̄⟨y,u⟩.0 | !y(r).⟦w⟧r ) )
  =  (νq)( q(z,p).z̄⟨p⟩.0 | (νy)( q̄⟨y,u⟩.0 | !y(r).w̄⟨r⟩.0 ) )
```

Now reduce. The communication on `q`, substituting `{y/z}` and `{u/p}`:

```
  →  (νq)(νy)( ȳ⟨u⟩.0 | !y(r).w̄⟨r⟩.0 )
```

Then the communication on `y`, substituting `{u/r}`:

```
  →  (νq)(νy)( !y(r).w̄⟨r⟩.0 | w̄⟨u⟩.0 )
```

And there it is: `w̄⟨u⟩.0`. An output on `w` carrying `u`.

Which is exactly right. `(λz.z) w` reduces to `w` in the λ-calculus, and "the answer is `w`" translates to "go and ask `w`, and tell it to report to `u`".

**Two communications did one β-reduction.** The substitution you've been performing on paper since chapter 1 has turned out to be the substitution that happens when two processes synchronise. That's the result.

---

## The bang is the evaluation strategy

Look again at the argument server:

```
!y(r).⟦N⟧r
```

That `!` means the argument can be asked for any number of times - and, crucially, that it isn't evaluated until somebody asks. `N` doesn't run until a request arrives on `y`.

That's **call-by-name**. The argument is passed unevaluated and re-evaluated at each use.

Now take the bang off, and instead run `N` once up front and pass the result. That's **call-by-value**.

Which is a genuinely startling thing to be able to say: *evaluation strategy is where you put the bang.* In the λ-calculus, call-by-name versus call-by-value is a whole separate apparatus - different reduction rules, different theorems, arguments about which is correct. Here it's a syntactic choice about one character in the encoding, sitting in plain sight.

The π-calculus made the decision visible because it made *sharing* visible. That's what a replication is: an explicit statement about how many times something may be used.

---

## Your turn

We're going to write the compiler.

```go
type Term interface{ term() }

type Var struct{ Name string }
type Lam struct {
	Param string
	Body  Term
}
type App struct{ Fn, Arg Term }
```

And:

```go
// Compile runs the process ⟦t⟧u, reporting on u.
// env maps λ-variable names to the π names standing for them.
func Compile(t Term, u Name, env map[string]Name)
```

Three cases, one per rule. `env` is how a `Var` finds the name it was bound to - the λ-calculus uses variable names, we use channels, and something has to bridge the two.

The test:

```go
func TestIdentity(t *testing.T) {
	// (λz.z) w
	term := App{
		Fn:  Lam{Param: "z", Body: Var{"z"}},
		Arg: Var{"w"},
	}

	w := make(Name)
	u := make(Name)

	go Compile(term, u, map[string]Name{"w": w})

	if got := <-w; got != u {
		t.Errorf("got %v, want %v", got, u)
	}
}
```

Note what we're observing. Not "the answer arrived on `u`" - the answer is `w`, and `w` is free, so what we see is the encoding going and asking `w` for it, with `u` as the return address. The observer plays the part of the environment.

You'll want `Send2`, `Recv2` and `Signal` from chapter 8.

---

## What this actually establishes

The π-calculus can do anything the λ-calculus can do, which is anything computable. It's Turing complete, and not by a grubby simulation - by a direct structural compilation where β-reduction becomes communication.

The converse fails, and it's worth being clear why. There's no encoding of the π-calculus into the λ-calculus, because there's nowhere to put the non-determinism. Chapter 1's race has no λ-term, because every λ-term has at most one normal form. Church-Rosser, still gone.

So the π-calculus is strictly bigger. Functions are a special case of processes - the deterministic, one-answer, call-and-return special case, which happens to be the one we build most of our languages out of.

---

## Exercises

1. **(paper)** Compile and fully reduce `⟦(λz.λw.z) a b⟧u`. It's the constant function - the answer should be `a`. Watch what happens to `b`'s argument server: it's still sitting there at the end, never asked. What would call-by-value have done differently?

2. **(paper)** Compile `⟦λz.z z⟧u`. Don't reduce it, just look at it. Where does the sharing show up?

3. **(paper)** In `⟦M N⟧u`, replace `(νy)` with a free name `y`. Describe an application that now goes wrong, and say which chapter's rule you've broken.

4. **(Go)** Write `Compile` and pass the identity test.

5. **(Go)** Test `(λz.λw.z) a b` and `(λz.λw.w) a b`. You'll need two free variables and a way to see which one gets asked.

6. **(Go)** Church numerals, the λ way: `λs.λz.s (s z)` is two. Compile it, and hand it the `s` and `z` channels from chapter 8's counter. Does it count as 2? You have now built the same number twice by two completely different routes, which is worth a moment.

7. **(Go, open)** Make it call-by-value. Change the argument server so `N` is evaluated once, up front. Find a term that behaves differently under the two strategies and show the difference with a test.

---

## Recap

- Every λ-term compiles relative to a **result channel** - somewhere to put the answer. It's the reply channel from chapter 1.
- `⟦ ⟧` are semantic brackets: inside is λ, outside is π. It's a compiler, defined by structural recursion on the grammar.
- A variable sends the result channel to itself. A function waits on its result channel for an argument and a new result channel. An application wires the two together with two private names.
- **β-reduction is two communications.** The substitution is the same substitution.
- The `!` on the argument server is call-by-name. **Evaluation strategy is where you put the bang.**
- π can encode λ. λ cannot encode π, because there's nowhere to put the non-determinism.

## Next

Chapter 10, and the question this book has been quietly asking since page one.

Every test you have written has claimed that two things are the same. Chapter 6 had two terms - a replication and a recursion - that were visibly different and behaved identically. Chapter 7 left you with two terms that produce identical histories and are *not* the same process.

So what does it mean for two processes to be the same? It turns out to be the deepest question in the subject, and the answer is a game.
