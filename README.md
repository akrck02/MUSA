<div align="center">

# MUSA

**A local-first inventory and business app for digital artists who also sell physical stock.**

It keeps products, stock, fairs, orders and the sales history in a local SQLite database on the device, and sells from a stand with a basket and «Cobrar». It needs no server or account. It is designed for an iPad, and runs on the desktop and Android.

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)
![Platforms](https://img.shields.io/badge/platforms-Desktop%20%7C%20Android-blue)
![Status](https://img.shields.io/badge/status-in%20development-orange)
![License](https://img.shields.io/badge/license-MIT-green)

</div>

> **Note:** written with an AI assistant and pending thorough review. See the [AI disclosure](#-ai-disclosure).

## 🤖 AI disclosure

I write most of my code by hand, so I want to be clear about this project: **it was written by an
AI assistant (Claude, by Anthropic) under my direction, in October 2026.** I decided what to build
and how it should behave; the AI wrote the code, the tests and this README. Every commit made this
way carries a `Co-Authored-By: Claude` line.

**Status: pending thorough review.** It builds and its tests pass on Linux. Android and the iPad
have not been tried with this version yet. I have not yet reviewed it line by line.
Treat it as a working draft, not as code I vouch for.

> ⚠️ **Warning:** in early development, **not suitable for production use**. The database is migrated when it opens, but the tables will still change.

---

## 🚀 Quick start

Requires Go 1.26 and checkouts of Go UI, Go Renderer and Bubble beside this one.

```bash
go run ./cmd/musa -demo   # the app, on a throwaway database with toy data
go test ./...             # the tests
```

[More on getting started →](docs/02-get-started.md)

## 🧭 What is in it

- **Mis productos**: physical and digital products, variants, stock, prices and tags.
- **Ferias**: stock taken to a fair, its sales and expenses, and the profit.
- **Puesto**: sell from a fair or the whole stock; pay in cash, with the change worked out, or by card.
- **Pedidos**: orders to prepare and sent on a board; delivered ones and local sales in the history.

## 📚 Documentation

| | |
|---|---|
| [Overview](docs/01-overview.md) | What MUSA does today and what it does not do. |
| [Get started](docs/02-get-started.md) | Build, run and test it, and where the data is stored. |
| [Architecture](docs/03-architecture.md) | One package per idea, and how they depend on each other. |
| [Inventory](docs/04-inventory.md) | Products, variants, validation rules, filters and duplication. |
| [Events](docs/06-events.md) | Stock taken to a fair, sales, expenses, profit and closing. |
| [Sales](docs/07-sales.md) | Orders, their status, the board and the sales history. |
| [Roadmap](docs/05-roadmap.md) | The modules still to come and the plan for sync. |

[Read the docs online →](https://itros97.github.io/MUSA/)

## License

MIT, see [LICENSE](LICENSE).
