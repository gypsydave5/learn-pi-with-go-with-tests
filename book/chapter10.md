# Chapter 10: When are two things the same?

## What we're covering

The question the whole book has been asking without saying so.

Every test you've written has claimed two things were the same. `got != y` claimed two names were the same name. `Count(Numeral(3)) == 3` claimed a chain of processes was the same as the number three. Chapter 6 had a replication and a recursion that were plainly different terms and behaved identically. Chapter 7 left you with two terms that produce identical histories and are *not* the same process.

So: when are two processes the same?

It's the hardest question in the subject and the answer is genuinely lovely.

## Setup

```
mkdir the-same && cd the-same
```

---

## The obvious answer, and why it fails

Two processes are the same if they can do the same things. Write down everything that could come out of one, write down everything that could come out of the other, compare the lists.

That's **trace equivalence**, and it's what every test in this book has been doing. Run it, see what happens, check the answer.

It worked for numerals because numerals are deterministic - one possible history, so listing histories tells you everything. It fails the moment there's a choice, and here's the pair from chapter 7 to prove it:

```
P  =  x(z).( ȳ⟨n⟩.0 + w̄⟨n⟩.0 )
Q  =  x(z).ȳ⟨n⟩.0  +  x(z).w̄⟨n⟩.0
```

Both can do: input on `x`, then output on `y`.
Both can do: input on `x`, then output on `w`.

Same traces. Every one of them, exactly.

Now put yourself opposite them, holding a name, and send on `x`.

Against `P`, you've got a process that will now offer you `y` **or** `w`, whichever you fancy. Both are still available. It's your call.

Against `Q`, sending on `x` has already destroyed one of the branches. You don't know which. You might now find that only `y` is on offer, and if you wanted `w` you're out of luck - and there was nothing you could have done differently, because the choice was made by the `+` at the moment you communicated.

**`P` commits late. `Q` commits early.** If you were a customer and these were two shops, you would care enormously which one you'd walked into.

The traces can't see it, because a trace is a record of what *happened*, and the difference is in what *remained possible*.

---

## The idea: play a game

Here's the move that fixes it, and it's the reason this chapter exists.

Instead of comparing the *histories* of two processes, compare them **step by step, and require that they still match afterwards**.

Make it a game. Two players.

- The **attacker** wants to prove the two processes are different.
- The **defender** wants to prove they're the same.

On each turn, the attacker picks either process and makes it do one thing. The defender must make the *other* process do the same thing. If it can't, the defender loses.

If it can, both processes have moved on - and now you play again, from wherever they've got to. Forever.

**`P ~ Q` means the defender can always keep going.** Not "wins" - there's no winning, the game never ends. The defender just never loses.

That's **bisimulation**, and that last paragraph is where people's brains go funny, so it's worth sitting with. We're not defining sameness by something that happens at the end. There is no end. We're defining it by the fact that nothing ever goes wrong.

That's **coinduction**. Ordinary induction builds up from a base case; coinduction is about properties that hold forever because they keep reproducing themselves.[^coind]

[^coind]: If you want a slogan: induction is about the smallest thing closed under the rules, coinduction is about the largest. Induction proves things terminate; coinduction proves they don't fall over.

---

## Playing it on the chapter 7 pair

Attacker's move: make `Q` do the input on `x`.

`Q` was `x(z).ȳ⟨n⟩.0 + x(z).w̄⟨n⟩.0`. The attacker chooses the *left* branch. `Q` becomes `ȳ⟨n⟩.0`. The right branch is destroyed.

Defender must match with `P`. Only one thing `P` can do on `x`, so `P` becomes `ȳ⟨n⟩.0 + w̄⟨n⟩.0`.

Both moved. Same label. So far so good.

Second turn. Attacker now picks `P` - which still has both options - and does the output on `w`.

Defender must match with the thing `Q` became, which is `ȳ⟨n⟩.0`. It cannot do an output on `w`. There is no `w` left anywhere in it.

**Defender loses.** `P` and `Q` are not bisimilar.

And notice exactly where it went wrong: not on the first move, but on the second, *after* the traces had already agreed. The whole point of the game is that it keeps going after the point where trace comparison stops looking.

---

## Proving it the other way: exhibit a relation

Losing the game shows two processes are different. To show they're the *same* you can't play forever, so you do something better: you write down the whole strategy in advance.

A **bisimulation** is a relation `R` between processes such that whenever `(A,B)` is in `R`:

- every move `A` can make, `B` can match, and the results are in `R`
- every move `B` can make, `A` can match, and the results are in `R`

And `P ~ Q` if there's *some* bisimulation containing `(P,Q)`.

Which is a construction you can actually do. You write down a set of pairs and check a closure property. **A table and an invariant.** You have written this test.

Let's do chapter 6's pair - the replication and the recursion:

```
S  =  !x(z).ȳ⟨z⟩.0
A  ≜  x(z).( ȳ⟨z⟩.0 | A )
```

Proposed relation:

```
R = { ( S | B , A | B ) : B is any process }
```

Check it. From `S | B`, the moves are: something `B` does on its own, or an input on `x`.

If `B` moves, `A | B` matches with the same move, and we land in `R` again with a different `B`. Fine.

If `S` takes an input on `x` with some name `m`, it becomes `ȳ⟨m⟩.0 | S | B`. And `A` takes the same input and becomes `ȳ⟨m⟩.0 | A | B`. Which is in `R`, with `B` now being `ȳ⟨m⟩.0 | B`.

Same in the other direction. Nothing escapes the relation, so `R` is a bisimulation, so `S ~ A`.

**That's a proof.** No induction on anything, no base case, no "and so on". You exhibited a set, showed it was closed, and you were done. It's not much longer than the paragraph explaining it.

---

## The wrinkle nobody warns you about

There's a version of this that isn't quite good enough and it's worth knowing about, because it explains a lot of otherwise baffling literature.

You'd want sameness to be a **congruence**: if `P ~ Q`, then you should be able to swap one for the other *anywhere* and have nothing change. `P | R ~ Q | R`, for any `R` at all. Otherwise it isn't much use - you couldn't refactor with it.

In the π-calculus, plain bisimulation isn't always a congruence, and the trouble is name-passing. When a process receives a name, that name might be one it already had, or a fresh one, and whether two processes match can depend on which. Substitution can break a bisimulation that looked perfectly sound.

The fixes are why you'll see **early**, **late** and **open** bisimilarity in the literature - they differ over exactly when a received name gets instantiated, and they aren't all the same relation. There's also **barbed congruence**, which takes the honest route: define it as "bisimilar in every context", by brute force.

I'm not going to work through them. What matters is that you know the shape of the problem, so that when you meet the zoo you understand what it's a zoo *of*: several careful answers to "when exactly do we plug the names in".

---

## Your turn: a bisimulation checker

Final piece of code in the book, and it's the only place where the machine gets to check a proof rather than run an experiment.

For finite processes - no `!`, or the state space is infinite - you can just do it. Explore both, compare, refine.

```go
// A finite process, as an AST.
type Proc interface{ proc() }

type Nil struct{}
type Pre struct {
	Label string // "x" for input on x, "'x" for output on x
	Next  Proc
}
type Par struct{ L, R Proc }
type Sum struct{ L, R Proc }

// Transitions returns every (label, result) pair this process can do.
func Transitions(p Proc) []Step

// Bisimilar reports whether p and q are bisimilar.
func Bisimilar(p, q Proc) bool
```

The algorithm, if you want a nudge: start by assuming *everything* is related to everything, then repeatedly throw out any pair where one side can make a move the other can't match. Stop when nothing changes. Whatever's left is the largest bisimulation - and coinduction is exactly why "start with everything and remove" is the right shape, where an inductive definition would have said "start with nothing and add".

The tests write themselves:

```go
func TestChapterSevenPair(t *testing.T) {
	// x(z).(ȳ⟨n⟩ + w̄⟨n⟩)
	p := Pre{"x", Sum{Pre{"'y", Nil{}}, Pre{"'w", Nil{}}}}
	// x(z).ȳ⟨n⟩ + x(z).w̄⟨n⟩
	q := Sum{Pre{"x", Pre{"'y", Nil{}}}, Pre{"x", Pre{"'w", Nil{}}}}

	if Bisimilar(p, q) {
		t.Error("these should not be bisimilar")
	}
}
```

Run that and the machine tells you what you worked out with a pencil a few pages ago. Which is, finally, the two instruments agreeing.

---

## What we've been doing all along

Back to the beginning.

Every test in this book has been an observer: a process standing outside a system, offering it interactions and watching what comes back. Chapter 1's `select`. Chapter 5's `s := <-x`. Chapter 8's counter.

We never once looked inside a process. We couldn't - there's no way to ask a goroutine what it's doing, and there's no term in the calculus meaning "is `P` stuck?". Everything we ever learned, we learned by interacting.

Bisimulation is what that turns into when you take it seriously. It says: **a process is exactly what it does when you interact with it, and two processes are the same if no amount of interacting can tell them apart.**

Which is why chapter 7's puzzle was hard. Both processes did the same things. But "the same things" wasn't the right test, because an observer can notice more than *what happened* - it can notice *what was still on offer*, and that's a fact about the process, not about the history.

Nothing else is left. There's no hidden state, no identity, no essence. A process is its behaviour, all the way down.

---

## Exercises

1. **(paper)** Are `x(z).0 + x(z).0` and `x(z).0` bisimilar? Play the game. Now try `x(z).ȳ⟨n⟩.0 + x(z).w̄⟨n⟩.0` and `x(z).w̄⟨n⟩.0 + x(z).ȳ⟨n⟩.0`.

2. **(paper)** Exhibit a bisimulation showing `P | Q ~ Q | P`. How big is your relation?

3. **(paper)** Show `(νs)( s̄⟨y⟩.0 | s(z).w̄⟨z⟩.0 ) ~ w̄⟨y⟩.0` - which is the thing chapter 4 asserted and couldn't prove.

4. **(paper)** Two vending machines. One takes your coin and then lets you choose tea or coffee. The other decides for you when you put the coin in. Write both as terms, show they have the same traces, and prove they aren't bisimilar. This is the classic example and it's worth doing in the coffee version.

5. **(Go)** Implement `Transitions` for the finite AST.

6. **(Go)** Implement `Bisimilar`. Test it against every pair in this chapter, and against exercises 1 and 4.

7. **(Go, open)** Extend the AST with restriction. What has to change in `Transitions`, and what does that tell you about why the literature has a zoo of bisimilarities?

---

## Recap

- **Trace equivalence isn't enough.** Two processes can have identical histories and differ in what remained possible.
- Bisimulation compares processes **step by step**, requiring them to still match after every move.
- Read it as a game: the attacker moves either process, the defender matches with the other, forever. `P ~ Q` means the defender never loses.
- There's no winning, because there's no end. That's **coinduction** - and it's why the algorithm starts with everything and removes, rather than starting with nothing and adding.
- To prove two processes the same, **exhibit a relation**: a set of pairs closed under every move. A table and an invariant.
- Plain bisimilarity isn't automatically a congruence in the π-calculus, because of name-passing. Hence early, late, open, and barbed.
- A process is exactly what it does when you interact with it. There is nothing else there.

## And that's the book

You started with `type Name chan Name` and a rule that fits on one line.

You've built numbers out of processes, compiled the λ-calculus into channels, and worked out what it means for two concurrent programs to be the same - which is a question most working programmers never get a straight answer to, and which you can now answer precisely, with a pencil.

None of it will help you on Monday. I did warn you.

But you will never look at a reply channel the same way again, and that's worth more than it sounds.
