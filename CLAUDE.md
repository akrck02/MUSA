# MUSA

Mobile and desktop app for a digital artist who also sells physical stock: inventory, accounts, marketplace, shipping, events and income-tax records. Built module by module, **starting with inventory**.

Bigger changes go through a proposal in `docs/proposals/`: spec → plan → approval → implementation.

## Decisions already taken

- **Stack**: Go, with Go UI for the screens, Bubble 4 for the look and Go UI for Android to package it for phones. SQLite through `modernc.org/sqlite` (pure Go). Module `github.com/Itros97/MUSA`. The first version was Kotlin Multiplatform; it was replaced and lives only in git history.
- **Target device**: an iPad; Android and desktop are supported too. Every control is sized for a finger.
- **v1 is fully local**: no server, one device per artist, a single user. Data in a local `.db` file with automatic periodic backups.
- **Later (not v1)**: optional sync with a Kotlin + Spring Boot backend as a modular monolith (not Quarkus). Sync will be last-write-wins with timestamps and client-generated UUIDs, so use UUIDs as ids from the start.
- **Domain**: `Product` with two kinds, `DigitalProduct` and `PhysicalProduct`.
- **Platforms**: develop on Linux; iOS builds on a Mac.

## Architecture

One Go package per thing the artist works with (`catalogue`, `events`, `sales`, `stand`, `cash`, `calendar`, `money`), each holding its model, its store and its screens. `storage` opens the database and migrates it; `look` holds the shared pieces of the screens; `app` is the frame and the navigation; `cmd/musa` is the program. Tests sit beside the code.
