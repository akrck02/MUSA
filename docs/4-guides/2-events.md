# Events
---

The events module tracks a fair or market: what stock goes there, what sells, and what the day really earned once costs and expenses are paid.

## An event

An event has a name, a place, a start date and an end date. It is **open** while it can take stock and sales, and **closed** once it is over.

## Taking stock

Stock **moves** from the catalogue to the event. Taking 5 prints lowers the catalogue by 5 and puts them on the event's list.

- You cannot take more than the catalogue has.
- While the event is open, the − and + buttons on a line move one unit back to the catalogue or bring one more.
- You can only give back units that were not sold.

Only physical products can be taken to an event.

## Selling

Each line has a **Sell** button. The sale starts at one unit and at the variant's price; both can be changed, for discounts or bundles.

A sale stores the product and variant names, the price and the cost at that moment, so editing the product later does not change past totals. Selling lowers what remains at the event; the catalogue does not change, because those units already left it.

A sale can be undone while the event is open. Its units then count as remaining again.

## Expenses

Expenses are what attending cost: stand, travel, lodging, food. Each has a concept and an amount.

## Summary

| Figure | How it is computed |
|--------|--------------------|
| Units sold | Sum of the quantities of every sale |
| Gross | Sum of quantity × price |
| Cost of goods | Sum of quantity × cost |
| Expenses | Sum of the expenses |
| **Profit** | Gross − cost of goods − expenses |

Profit is negative when the event lost money.

## Closing and deleting

Closing an event returns every unsold unit to the catalogue and makes the event read-only.

An event can only be deleted while it has no sales. Its stock goes back to the catalogue. An event with sales must be closed instead, so the sales are kept.

## Where sales are stored

Each sale at the stand is an order of one line, with the channel set to `EVENT` and a link to the event. It is stored with every other sale, so it also shows in the [sales list](3-sales.md), where it can be seen but not undone.
