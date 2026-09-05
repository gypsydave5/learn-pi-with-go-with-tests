# Introduction

## This will not help you at work

Let's get that out of the way.

Nobody has ever shipped a fix faster because they knew what scope extension was. No incident has been averted by a solid grasp of structural congruence. If you are looking for a book that will make you better at your job by Friday, it isn't this one, and I'd rather say so on the first page than let you find out in chapter 4.

So here's what it's actually for.

## It changes what you notice

You have written this, or something very like it:

```go
type Request struct {
	Data  string
	Reply chan Response
}
```

A reply channel. Perfectly ordinary. You made a channel, put it in a message, and sent it to somebody who had no way of knowing it existed a moment earlier.

There's a name for what you did, it was invented in 1992, it is the single feature that distinguishes the π-calculus from everything that came before it, and it turns out to be the same idea as a capability, a session token, and a signed URL.

That's the offer. Not new powers - new eyes. By the end of this you will not be able to look at a channel being passed around without seeing what it is, and that's a permanent change and rather a nice one.

## You're going to write some proofs

I know. Bear with me.

There is a whole population of programmers who'd quite like to be able to prove things and have bounced off every attempt, because the textbooks assume a maths degree and the proof assistants make you fight the syntax before you've had a single idea. That's a real gap and this book is my attempt at filling it.

I should be clear that I am one of these people rather than a visitor from the other side. I love the *idea* of maths considerably more than I have ever loved the doing of any, and I have window-licked my way across a fair number of adjacent subjects over the years, nose pressed against the glass, admiring things I could not quite do. This is merely the latest. If that sounds like you, then we're going to get along, and you should take everything I say about how manageable it all is with the appropriate pinch of salt - I'm working it out as I go, same as you.

The π-calculus is a good vehicle for it because the entire formal system fits on a postcard. There are two operations, one reduction rule, and about six laws. The objects are channels and processes rather than rings and fields, so you already have intuitions about them. And - this is the useful bit - every claim you prove can also be *run*, so you get two instruments pointed at the same question and can watch them disagree.

They do disagree. Chapter 1 has a program with two possible outcomes where the scheduler will only ever show you one of them - you can only go down one leg of the trousers of time.[^pratchett] Chapter 4 has two programs that no test can distinguish. In both cases the pencil gets you the answer and the machine doesn't, and by the time that's happened twice you'll want the pencil.

[^pratchett]: Terry Pratchett's, and much the best way anyone has put it.

**So get a pencil.** Exercises come marked *paper* or *Go*, and the paper ones aren't optional extras. They're where most of the learning is, and they're shorter than you're afraid they are - a page of working, crossings-out, one or two false starts. Squared paper helps. Nobody's marking it.

## It's fun

This is the real reason and I'm not going to dress it up.

In chapter 8 we build the natural numbers out of processes. Not numbers *represented by* processes - numbers that **are** processes, where two is a thing that talks to you twice. It is completely useless and it is one of the most delightful things I know.

If you've ever done Church numerals in JavaScript at eleven at night for no reason, you already understand the appeal and you can skip the rest of this section. If you haven't: that's what's coming, and it's worth the trip.

## What you need

Go 1.25 or later, because we lean on `testing/synctest` and it landed in 1.25.

Knowing some Go. Not expert Go - if you're comfortable with goroutines and channels you're fine.

No λ-calculus. Chapter 1 leans on it for about four paragraphs to explain why the π-calculus had to exist, and I've tried to make those paragraphs work whether or not you've met it before. Everything after that is explained in Go, because that's the language we both actually speak.

No maths. Genuinely. There's some notation, we build it up one symbol at a time, and there's nothing in here that a careful reader can't follow.

## A note on the title

It's called *Learn the π-calculus with Go (with tests)* because it stands on the shoulders of [Learn Go with Tests](https://quii.gitbook.io/learn-go-with-tests), which is where a great many of us learned Go properly, and which taught me that a technical book can be written by somebody who is working it out as they go.

Chapter 4, in the course of things, discovers that the most interesting property in the whole calculus cannot be tested. At all. By anything.

I've decided to keep the title anyway.

## How this goes

Each chapter introduces one piece of the calculus, maps it onto Go, and hands you a failing test. You write the code. Then there are exercises, some on paper and some in the editor.

Nothing is held back for a sequel. By chapter 9 you'll have built the λ-calculus out of channels, and by chapter 10 you'll know what it means for two concurrent programs to be the same - which is a question you've been implicitly answering every time you've written a test, and which turns out to be much harder and much more interesting than it looks.

Right. Chapter 1.
