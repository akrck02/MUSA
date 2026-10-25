# 4. Orders that travel, payments and commissioned projects

> Status: done in part (R1–R6, R11, R12, R15); R7–R10, R13 and R14 wait for Etsy, SumUp and Trello · Author's approval: proposal ☑ plan ☑ · Lives in: `orders`, `payments` (new), `customers` (new), `projects` (new), `storage`
> Depends on: proposal 3 (definitions, bundles, `orders`)

## Why

The artist sells in three ways: at a stand or by hand (a **local sale**), through Etsy, and by
drawing what a customer asks for (a **project**). Today MUSA only writes sales down after the
fact. The author's design asks for more:

- **Pedidos** shows only what still has to travel (to prepare, sent). Delivered orders and local
  sales go straight to the **sales history**.
- Etsy orders arrive on their own, and marking one as sent tells Etsy.
- At the stand, «Cobrar» takes the money: in cash, with the change worked out, or by card through
  SumUp.
- Projects live on a board like Trello's (and talk to Trello), each with its customer and the
  customer's social networks, so the artist can reach them.

Etsy, SumUp and Trello are the first of each kind, not the last: each goes behind an interface
(the strategy pattern), so another shop, card reader or board is one more implementation.

## What

| Id | Requirement | Check |
|---|---|---|
| R1 | An order has a channel (local, Etsy, project) and a status (to prepare, sent, delivered) | test `a new online order starts to prepare` |
| R2 | A local sale is saved delivered, without shipping: it never shows on the Pedidos board, only in the history | test `a local sale goes straight to the history` |
| R3 | Marking an order delivered moves it from the board to the history; the status only moves forward | test `a delivered order leaves the board`; test `a sent order cannot go back to prepare` |
| R4 | The history lists every delivered order, filtered by channel, with the month's revenue, cost and what is left | test `the history adds up the month` |
| R5 | An order has one or more payments (method, amount, reference); paid in full means the payments add up to the total | test `an order is paid when its payments add up` |
| R6 | Paying in cash: the artist taps the notes and coins received; MUSA shows the change and saves only on «Confirmar»; too little money is said, not saved | test `cash shows the change`; test `too little cash is not accepted` |
| R7 | Paying by card goes through a `payments.Terminal` (SumUp first); stock leaves only when the payment succeeds | test with a fake terminal that fails: `a failed card payment keeps the stock` |
| R8 | Orders from a shop arrive through an `orders.Source` (Etsy first); each external order is imported once | test with a fake source: `fetching twice imports once` |
| R9 | Marking an imported order as sent tells its source, with the tracking number when there is one | test with a fake source: `sending tells the shop` |
| R10 | A shop product is linked to a MUSA product, so its sale takes MUSA's stock; an unlinked one is said, not guessed | test `an unlinked listing is reported` |
| R11 | A customer has a name, contact and any number of social profiles; the networks (Instagram, Bluesky, X…) are a list the artist edits, each with its link pattern | test `a profile becomes a link` |
| R12 | A project has a customer, a title, a status column, a deadline and a price, may name the digital product it makes, and becomes an order when delivered | test `delivering a project records its order` |
| R13 | Projects are shown on a board whose columns can mirror a Trello board through `projects.Board`; MUSA works the same without Trello | test with a fake board: `moving a card moves the project` |
| R14 | No secret (tokens, keys) is stored in plain text in the database or in the repository | review; test `tokens are not in the projects table` |
| R15 | Every new problem and failure has a code and a short sentence (table below) | the tests of proposal 3, extended |

### New error codes

| Code | Kind | When | What the artist reads |
|---|---|---|---|
| `payments.not-enough-cash` | problem | the cash received is less than the total | «Falta dinero» |
| `payments.card-declined` | failure | the terminal says no | «La tarjeta no ha pasado» |
| `payments.terminal-unreachable` | failure | no answer from SumUp | «No hay conexión con el lector» |
| `orders.status-backwards` | problem | moving an order back | «Un pedido no puede volver atrás» |
| `orders.source-unreachable` | failure | the shop does not answer | «No hay conexión con Etsy» |
| `orders.source-unauthorized` | failure | the shop's permission ended | «Vuelve a conectar Etsy» |
| `orders.listing-not-linked` | problem | a shop product with no MUSA product | «Este producto de Etsy no está enlazado» |
| `orders.duplicate-external-order` | failure | the same external order twice | «Este pedido ya está en MUSA» |
| `projects.product-not-digital` | problem | a project naming a physical product | «Un proyecto solo lleva productos digitales» |
| `projects.board-unreachable` | failure | Trello does not answer | «No hay conexión con Trello» |
| `customers.network-pattern` | problem | a network's link pattern without `{user}` | «El enlace tiene que llevar {user}» |

## Out of scope

- Shops other than Etsy, readers other than SumUp, boards other than Trello: the interfaces are
  ready for them, the implementations are not written.
- Refunds, partial deliveries, invoices and income tax.
- Stock movements as their own table (`STOCK_MOVEMENT`): considered, not needed yet.
- Syncing between two devices.

## Design

```
CUSTOMER        1:N CUSTOMER_SOCIAL N:1 SOCIAL_NETWORK   (name, link pattern "https://…/{user}")
CUSTOMER        1:N ORDER                                optional
CUSTOMER        1:N PROJECT
ORDER                                  channel, status, shipping, source, external_id (unique with source)
ORDER           1:N PAYMENT            method, amount, reference
PROJECT                                title, column, deadline, price, board_card_id
PROJECT         N:1 PRODUCT            optional, digital only
PROJECT         0..1 ORDER             made on delivery
SOURCE_SYNC                            source, last fetched at
PRODUCT_LISTING                        source, external listing id → PRODUCT
```

Three interfaces, each with a fake for tests and one real implementation:

```go
// orders
type Source interface {
	Name() string
	Fetch(since time.Time) ([]ExternalOrder, error)
	MarkShipped(externalID string, tracking Tracking) error
}

// payments
type Terminal interface {
	Charge(amount money.Money, reference string) (Receipt, error)
}

// projects
type Board interface {
	Columns() ([]Column, error)
	Move(cardID, columnID string) error
}
```

- **Etsy**: Open API v3, OAuth 2 with PKCE, `getShopReceipts` and `createReceiptShipment`.
- **SumUp**: its card reader API, through its SDK on iPad and Android.
- **Trello**: REST API with a key and a token.
- The tokens live in the system's keychain (iOS Keychain, Android Keystore, the desktop's secret
  store), never in the database (R14).

Packages: `payments`, `customers` and `projects` are new and sit beside `orders`. Each outside
service gets its own package (`etsy/`, `sumup/`, `trello/`) that implements the interface, so the
domain knows nothing of HTTP.

**Not taken:** a fourth status «local» (a local sale is simply delivered at once, R2); a separate table for the
history (it is a filter over orders); keeping projects only in Trello (MUSA must work offline and
without an account).

_Tasks to be written once the proposal is approved._

## Result

_To be filled at the end._
