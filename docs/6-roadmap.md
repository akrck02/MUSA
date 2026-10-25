# Roadmap
---

MUSA is built one module at a time. Inventory, events and sales are done; the rest follow in roughly this order.

## Version 1, fully local

- [x] **Inventory**: products, variants, stock, prices and tags.
- [x] **Sales**: orders from every channel, with shipping amounts and margin.
- [ ] **Backups**: automatic periodic copies of the local database.
- [ ] **Accounts**: income and expenses.
- [ ] **Marketplace**: listings and orders from the places the artist sells.
- [ ] **Shipping**: parcels and their state.
- [x] **Events**: fairs and markets, with the stock taken, sales, expenses and profit.
- [ ] **Income tax**: records needed for the yearly tax return.

## Later

- Optional sync with a Kotlin and Spring Boot backend, built as a modular monolith.
  Conflicts are resolved by last write wins, using timestamps and the UUIDs the app already generates.
