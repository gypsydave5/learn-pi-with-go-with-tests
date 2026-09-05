# Learn the π-calculus with Go (with tests) — plan

Working document. Notes to ourselves, not prose.

---

## What this book is

A gentle introduction to writing proofs, for programmers who've never written one, disguised as a book about concurrency in Go.

The π-calculus is the right vehicle because the formal system fits on a postcard, the objects are channels and processes rather than rings and fields, and every claim can also be *run*. Two instruments, answering different questions — which the book demonstrates rather than asserts:

- ch1: `-race` shows one outcome; the reduction shows all of them.
- ch4: no test can distinguish the cheat, so paper is the only instrument left.
- ch5: the test catches mobility but not freshness.

Pencil and paper is a first-class activity, not homework. **Mark exercises as *paper* or *Go*** — not done in any chapter yet. Say so in the introduction: readers who'd be put off by "proofs" are much less put off when it's declared up front rather than sprung on them in chapter 5.

**Priorities.** Teaching the π-calculus first. Proofs second. "With tests" is bottom of the list — don't manufacture a failing test for material that isn't testable. Ch3 has exactly one and that's fine.

**Register discipline.** The failure mode is drifting into theorem-definition-lemma prose. Every law should be *wanted* before it's given — withholding scope extension through ch4 is the model. Every proof should do a job the reader felt they needed. A lemma introduced because it's next in the textbook is a lemma to cut.

---

## Method

- Claude drafts a chapter, ending at a failing test plus enough exposition to attempt it.
- Dave writes all the Go. Claude never supplies the solution first.
- Dave reads, gets stuck, reports. Both kinds of stuck — "I don't follow" and "I follow but can't do it" — mean a rewrite; the label just changes what sort.
- Iterate.

**Exit criterion for a chapter:** Dave can write the *next* chapter's code from its failing test alone, without re-reading the previous chapter. Tests retention, not clarity.

**House pattern for exercises:** predict, then run.

---

## Conventions

**Notation** — Milner/Parrow/Walker.

| | |
|---|---|
| output | `x̄⟨y⟩.P` |
| input | `x(z).P` |
| restriction | `(νx)P` |
| replication | `!P` |
| choice | `P + Q` |
| parallel | `P \| Q` |
| inert | `0` |
| substitution | `Q{y/z}` — new over old |

**Never use `v` as a name.** Visually indistinguishable from `ν` (U+03BD) in most fonts. Use `q`. Same care with `u` near a `ν`.

Overbar is a combining character (U+0304), so single-letter names in all calculus terms. Precomposed forms (`ā` U+0101 etc.) must be normalised to base + U+0304 — mixed representations break grep and sed.

**Never open a section with a bare `ν`** — it just reads as a `v`. Name the letter on first use in any chapter.

Reserve "the bar" for the overbar. `|` is always "parallel composition". "Synchronise" is the house verb for what happens at a rendezvous — not meet, pair, or communicate-with.

Say "**occurs free in**", not "is free in".

**Observer vs process.** Processes use the process functions (`go Send(y, z)`); the test is the observer and uses raw channel operations (`x <- y`, `<-out`). Writing `go` inside a test means adding a process to the system, and should be deliberate.

**Spawn sites name what they pass.** `go Send(y, z)`, or `go func(y, z Name){...}(y, z)`. Not `go func(){...}()`. This is discipline — `-race` will not check it.

**λ-calculus references** — λ earns its place in ch1 (the reductive move the reader already knows; Church-Rosser and its failure) and ch9 (the encoding). **In between, explain with Go.** The audience definitely knows Go and only maybe knows λ. Ch1 tells readers not to worry if they haven't done much λ — so don't then lean on it.

**Words** — "specification" for a term, not "type" or "signature".

**Asides** — three tiers:
- *sidenote* (footnote): <40 words, attached to a point, skippable entirely
- *aside* (blockquote → `<aside>`): needs paragraphs, has a title, reader should read it but it isn't the argument
- *parenthetical*: a clause

Test: if deleting it breaks the next paragraph, it's the argument — put it back in the body.

**Code** — one module at the root, directory per chapter, no versioned subdirectories. Each chapter self-contained; redefine `Name` rather than importing. Extract shared `Send`/`Recv` *when it starts hurting*, not before.

---

## Named results

Proofs are functions. Once derived, name it and call it rather than inlining the same four steps. Numbered by chapter, listed in an appendix, citable from exercises.

Worth naming if used more than once, or if the derivation is long enough that repeating it obscures the argument it sits in. Not worth naming if it's a single axiom wearing a hat.

**L5.1 — Dead ν.** `(νz)P ≡ P` when `z` does not occur free in `P`.
Proved ch5 (DW). Pad with `0 | P`, scope extension right-to-left, then `(νz)0 ≡ 0` and `P | 0 ≡ P`.
Used by: ch5 ex.1 (the final step), and anywhere a restriction outlives its user.

*Expected later:* replication unfolding `!P ≡ P | !P` (axiom, but wants a name); recursion/replication interdefinability with the concurrency caveat; polyadic-is-sugar.

**The padding trick** wants a callout around ch3: to apply a law that needs a `|`, conjure one with `P ≡ 0 | P`. It's how most of these proofs open and a reader who hasn't seen it will stall on line one.

---

## Chapters

### 0. Introduction — not written
Why bother; what the book is; pencil and paper. Currently duplicated in ch1's "Why bother?", which will want trimming once this exists.

### 1. A name is a channel is a name — drafted, not run
Names, input/output, the reduction rule, substitution, the dot, `0`, `|`. First tests. Non-determinism via `-race`. `synctest.Wait()` and the bubble's must-reach-`0` rule.

### 2. And then what? — drafted
The dot's three jobs, the third being scope. Lexical scope. Free and bound, *relative to a term*. Alpha-conversion. Capture. `x(z).0` is a legal sink but won't compile in Go.

Dave's note: reads slower than ch1.

### 3. Everything, all at once — drafted
`≡`: commutative, associative, `P | 0 ≡ P`. Reduction up to `≡`. Why the strictness is the point. Precedence: `.` binds tighter than `|`. No shared state; spawn sites name what they pass.

**Corrected:** `-race` does *not* enforce the no-closures rule. It catches shared mutable state only. Read-only capture is race-free and idiomatic — that one's discipline. The worker-loop history (pre-1.22 capture bug, fixed in 1.22, habit still right) is the motivating example.

### 4. Where names come from — drafted
`(νz)P`. Made, not received — both binders make locals, the difference is provenance, and nothing is ever substituted for a restricted name. Unforgeability, and that it survives only while we keep the no-globals rule. `(νz)(νw)P ≡ (νw)(νz)P`, `(νz)0 ≡ 0`, `fn((νz)P) = fn(P) \ {z}`.

**The lesson:** a `ν` that never leaves home is unobservable, so `w <- y` passes the test and *is correct*. Ends by deriving `(νs)w̄⟨n⟩.0` and being unable to drop the `ν` — because scope extension is deliberately withheld.

Title candidate: **Nu names** (nu metal, nu rave — the affected misspelling that means "new"). Better than the current one, which is accurate and forgettable.

### 5. The scope moved — drafted
Scope extension. The side condition as the guarantee, glossed as "`Q` doesn't take a `z` as a parameter". Extension (the law) vs extrusion (the phenomenon). Capabilities. Mobility, and `type Name chan Name` finally earning its recursion. `Gate`, with the observer receiving the private name.

The cheat fails here — but because the test detects *mobility*, not freshness. Freshness is still discipline.

### 6. Replication
`!P ≡ P | !P`. `for` loops, goroutine leaks, `goleak`.

**Teach replicated input `!x(y).P`, not general `!P`** — its own axiom, the form the literature mostly uses, and the shape you'd actually write.

**Recursion is not replication:**

```
A ≜ x(z).ȳ⟨z⟩.A          sequential. One at a time. A worker.
!x(z).ȳ⟨z⟩.0             concurrent. Unboundedly many. A server.
A ≜ x(z).(ȳ⟨z⟩.0 | A)    recursion made concurrent, by hand.
```

Interdefinable, but the naive translation silently changes the concurrency — the bug worth making the reader write. In Go the recursive version has no TCO, so the stack grows forever: survives a test, dies in production, and pushes you towards `for`, which is the honest translation anyway.

Named process definitions (`A ≜ …`) get properly introduced here. Ch1 mentions them only in a footnote, to say `P` and `Q` are *not* that.

### 7. Choice
`P + Q`, `select`. Why you can't build it from race-style primitives — the losing offer must be de-registered atomically. The observer finally becomes writable as a term: `a(q).0 + b(q).0`. Ch1 promises this.

### 8. Data as behaviour
Booleans, numerals, lists as processes. Parallel with Church encodings.

Numerals are **deterministic**, so trace comparison is a sound test here — and ch10 explains why you got away with it.

Honest coda: unary numerals mean four billion rendezvous to count to 2³². Not a defect — the price of "everything is interaction".

### 9. Milner's encoding
λ into π. CBN vs CBV is *where you put the bang*. Evaluation strategy as an explicit syntactic choice. The CPS footnote from ch2 pays off here.

### 10. Bisimulation
"What have we been asserting all along?" Trace equivalence vs bisimulation — `x(z).(ȳ⟨n⟩.0 + w̄⟨n⟩.0)` vs `x(z).ȳ⟨n⟩.0 + x(z).w̄⟨n⟩.0`, same traces, different processes.

Teach it as a **game**: I move, you match, forever. `P ~ Q` means the defender never *loses* — not wins — which is what makes it coinduction. And it's constructive: to prove bisimilarity you *exhibit the relation*, write down the pairs, check closure. A table and an invariant.

Also inherits ch1's cut material: offers and observations, and that we never inspect a process, only interact with it. By ch10 the reader has written it eight times.

The arc: ch1–9 teach *doing is meaning*; ch10 shows meaning is coarser than doing.

---

## Captured, unplaced

- **Milner's Turing lecture** — the π-calculus as capturing the uniformity of values and processes found in the actor model. His own account, better than anything currently in "why bother". *Elements of interaction*, CACM 36(1), 1993.
- **Where the calculus stops being useful.** A closed term can't touch a socket; the environment is always a free name. Nothing enforces that a server replies. Worth a coda rather than only showing the beautiful bits.
- **occam-π** — occam plus mobile channels. Currently in ch1's CSP aside.
- **Session types** as the recurring "this is the gap" refrain: ch1 (signatures tell you nothing, not even direction), ch3 (nothing checks your program matches your whiteboard), ch5 (unforgeability is discipline, not a check), ch7 (protocols unenforced).
- **A photograph of real working out** — crossings-out and a false start. Says "this is supposed to look like this" better than any clean code block. LGWT has nothing like it.
- **Do not build a reduction tool.** It would remove exactly the friction that does the teaching.

## To verify before print

- The Milner quote is from the book's abstract, not the preface. Check the real text.
- CSP aside: does the 1978 paper name processes rather than channels? How far did occam allow channel passing?
- Go 1.22 loop-variable change — and that it's gated on the `go` directive in `go.mod`, not the toolchain version.
- `synctest`: does a `select` mixing bubbled and non-bubbled channels count as durably blocking? Matters for ch7.
- The deadlock panic message in ch5 — currently invented. Paste the real one.
- Does `-race` flip ch1's race *reliably*, or only sometimes? The text promises "you should see both".
- Overbar + angle brackets rendering in the real toolchain, body text and monospace.
- All code in all chapters actually compiles and passes.
