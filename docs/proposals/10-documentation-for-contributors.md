# 10. Documentation for the people who build MUSA

> Status: done · Author's approval: proposal ☑ (2026-10-25) plan ☑ (2026-10-25) · Lives in: `docs/`, the root's
> community files, `.github/`, and one style test
> Depends on: nothing; follows the layout of Go Renderer and Go Engine

## Why

At least three people will work on MUSA at once. Today the docs explain what the app does
(inventory, events, sales) and give one page of architecture, but nobody new can learn from them
how a feature is built here: where its pieces go, how a screen is drawn and tested, how to add a
migration without clashing with someone else's. Each of them would ask the same questions, or
guess differently, and the code would drift into three styles.

## What

| Id | Requirement | Check |
|---|---|---|
| R1 | The docs follow the layout the other Go projects use, numbered for three readers: who uses MUSA, who changes it, who learns from it. Links inside the repository are updated; links to the old published pages break once | a person reads the menu; a test finds no broken relative link |
| R2 | **The code map**: each package, what it holds, what it may import, and its usual files (`model`, `store`, `screens`, `words`, `errors`, tests) | a person finds where a given screen lives in under a minute |
| R3 | **How a feature is built**, as a walk-through of one real, small feature from proposal to merge: the model, the migration, the store, the words, the screen, the tests, the docs | a newcomer follows it and ends with a working change |
| R4 | **How the screens work**: immediate mode, the state a package keeps, `look` and Bubble 4, sheets, navigation through `app`, accessible labels | the page's examples compile (an `Example` test) |
| R5 | **How to test**: logic tests, store tests on a temporary database, flow tests with `uitest`, the results viewer (`cmd/uitest`), captures, and what still needs a look on screen | each command shown runs |
| R6 | **The style guide**: the author's guidelines as they apply to MUSA, each rule with a short right and wrong example; the rules a program can check are checked by a test | the tests in `internal/style` |
| R7 | **Working together**: branches, proposals, reviews, who approves, commit messages, and the places where parallel work clashes (migrations, `app` navigation, shared `look` pieces) with how to avoid each | read by the three people before starting |
| R8 | A short `CONTRIBUTING.md` at the root that points into the chapter, a pull request template, issue templates, and CI that runs the formatter, `go vet` and the tests on every push | the files exist; CI runs on a branch |
| R9 | **Troubleshooting**: the errors people meet when building and running (cgo, the `x11` tag, Metal, a locked database), by their message or code, with the fix | a person finds the fix for each error met so far |

## Out of scope

- The journey chapter (`9-journey/`): only the author knows those decisions; it is left with an
  outline to fill in their voice.
- Rewriting the user chapters (inventory, events, sales, stand, projects) beyond moving them.
- Translating the docs; they stay in English, the app's words in Spanish.

## Design

The new layout, after Go Engine's:

```
docs/
  1-overview.md
  2-get-started.md
  3-core-ideas.md            the few ideas behind everything: local first, one package per idea,
                             immediate-mode screens, money in cents, days without time zones, UUIDs
  4-guides/                  the artist's chapters as they are today: inventory, events, sales,
                             stand, projects
  5-reference/               every package's public names, generated (go doc), and the error codes
  7-troubleshooting.md
  8-contributing/
    1-the-code.md            the code map (R2) and the direction of dependencies
    2-a-feature-step-by-step.md   (R3)
    3-screens.md             (R4)
    4-testing.md             (R5)
    5-style-guide.md         (R6)
    6-working-together.md    (R7)
  9-journey/                 outline only
  proposals/
```

The walk-through in R3 builds a real but small feature on a branch of its own, so every step shown
is code that exists and compiles; the candidate is "a note on a customer", which touches every
layer and nothing anyone else is changing.

The style test (R6) is Go Engine's `internal/style` idea brought here: it walks the code and fails
on short names, missing comments on public names, `errors.New` outside an `errors.go`, and
`map[bool]string`.

R7 suggests one rule for migrations, the main place three people clash: migrations are appended
only on `main`, right before merging, never on a long-lived branch.

Alternatives not taken: one long `CONTRIBUTING.md` (too long to read and to keep true); a wiki
(lives apart from the code and goes stale).

## Tasks

Each task leaves the docs readable and the tests green; the user chapters move first so later
pages link to their final places.

- [x] T1 — Move the chapters to the new layout (`1-overview`, `2-get-started`, `4-guides/1-inventory`
  … `5-projects`), fix every link, the README's included · `docs/`, `README.md` · serves R1 · test:
  `TestEveryDocsLinkLeadsToAFile` (new, `docs/links_test.go`)
- [x] T2 — `3-core-ideas.md`: local first, one package per idea, immediate mode, cents, days, UUIDs ·
  serves R1 · read by a person
- [x] T3 — `8-contributing/1-the-code.md`: the map, from `03-architecture.md` (which it replaces),
  with each package's usual files and what it may import · serves R2 · test:
  `TestPackagesImportOnlyWhatTheMapAllows` (the direction of dependencies, checked)
- [x] T4 — The style test in `internal/style`, after Go Engine's: short names, comments on public
  names, `errors.New`/`fmt.Errorf` outside `errors.go`, `map[bool]string`; fix what it finds ·
  serves R6 · tests: the six in `internal/style/style_test.go`
- [x] T5 — `8-contributing/5-style-guide.md`: each rule, why, right and wrong, and which the test
  checks · serves R6
- [x] T6 — Build «a note on a customer» on its own branch, step by step, then write
  `8-contributing/2-a-feature-step-by-step.md` from its commits; the feature itself waits for the
  author's yes before merging · serves R3 · tests: the feature's own
- [x] T7 — `8-contributing/3-screens.md`, with an `Example` test in `look` for its code · serves R4 ·
  test: `ExampleCard` (or the piece shown)
- [x] T8 — `8-contributing/4-testing.md`: every kind of test, `cmd/uitest`, captures, what needs a
  screen · serves R5 · each command run here
- [x] T9 — `8-contributing/6-working-together.md`: branches, proposals, review, commits, the clash
  points (migrations, `app` navigation, `look`) · serves R7
- [x] T10 — `CONTRIBUTING.md`, `.github/pull_request_template.md`, issue templates (bug, feature),
  `.github/workflows/test.yaml` (gofmt, vet, test, with the `x11` tag and the GLFW headers) · serves R8
- [x] T11 — `7-troubleshooting.md` from the errors met so far; `5-reference/` generated with a
  script, with the error codes; `9-journey/` outline · serves R9, R1
- [x] T12 — Review with the `reviewer` agent, fill **Result** · all

## Result

Built: the new layout (R1) with a test of every link; the code map with a test of the
dependencies (R2); a feature step by step, from the branch `feature/customer-note` (R3); screens,
with an `Example` test (R4); testing (R5); the style guide, with six checks in `internal/style`,
and the code brought up to them (R6); working together (R7); `CONTRIBUTING.md`, the templates and
CI (R8); troubleshooting and a generated reference with every error code (R9).

Reviewed by the `reviewer` agent; its findings were fixed, except:

- CI checks out the sibling libraries by branch, not by commit, so a push there can turn MUSA's
  CI red. Pinning commits is the author's call.
- Code constants inside a commented group need no comment of their own; asking for one would
  touch every `errors.go`.
- `feature/customer-note` (the example feature) waits for the author's yes to merge.

Still for a person: reading the pages as a newcomer, and the CI's first run on GitHub.
