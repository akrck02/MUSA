# 💡 Core ideas
---

This page is for anyone about to read MUSA's code: by the end you know the six ideas that explain
why almost every file looks the way it does.

## 1. Local first

MUSA runs on one device, for one artist, with no server and no account. Everything lives in one
SQLite file (`musa.db`), opened by `storage`. There is no network code to wait for and nothing to
log in to, so every screen can read its data straight from the database, every time.

The cost is that two devices cannot share data yet. The [roadmap](6-roadmap.md) plans a sync; the
next idea is what keeps that door open.

## 2. Ids are made on the device

Every row gets a random UUID from `storage.NewID` when it is created, never a number the database
counts up. Two devices that later sync can both create products without ever choosing the same id,
so nothing will have to be renumbered.

Stores take the function that makes ids (`newID`) as an argument. The app passes `storage.NewID`;
a test passes one that counts, so what it checks is the same on every run.

## 3. One package per idea

A package is a thing the artist works with: `catalogue`, `events`, `sales`, `stand`, `projects`,
`customers`. It holds all of that thing: its model, its rules, the store that reads and writes it,
its screens and its words. To change how fairs work, you open `events/` and nothing else.

Think of a shop with one counter per department, rather than one counter for every form, one for
every till and one for every shelf. The alternative, folders by kind of file (`models/`,
`views/`, `services/`), would scatter one feature across the whole tree.

Packages depend one way: base ideas never import what is built on them. The
[code map](8-contributing/1-the-code.md) draws the arrows, and a test keeps them true.

## 4. Screens are drawn every frame

The screens use Go UI, which is **immediate mode**: there is no tree of widgets kept in memory.
Each frame, a function walks through what should be on screen and draws it, and the same call
that draws a button says whether it was clicked:

```go
if context.Button("Guardar") {
	screens.save()
}
```

So there are no view models and no events to subscribe to. A package's `Screens` keeps the little
state it needs between frames (what is open, what was typed, what it read from the store), and
reads the store again after each change. [Screens](8-contributing/3-screens.md) goes deeper.

## 5. Money is whole cents, days have no hours

`money.Money` is an `int64` of cents: `Money(1250)` is 12,50 €. Adding prices never rounds, which
floating-point numbers would do.

`calendar.Date` is a day with no hour and no time zone: a fair starts on a day, not at midnight in
some zone. Dates are stored as ISO text (`2026-04-18`), so text order is date order.

## 6. Errors have codes; words are Spanish and live apart

Every error a package returns carries a stable code, such as `catalogue.blank-name`, declared in
that package's `errors.go`. Programs and tests check the code, never the text.

What the artist reads is Spanish and lives in `words.go`, apart from the rules. Code, comments and
docs are in English, so the code reads the same for every contributor, while the app speaks the
artist's language.
