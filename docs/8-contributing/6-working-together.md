# 🤝 Working together
---

This page is for the people who build MUSA at the same time: by the end you know how work moves
from an idea to `main`, who decides what, and how to keep out of each other's way.

## From an idea to `main`

```
idea ─▶ issue ─▶ (proposal) ─▶ branch ─▶ pull request ─▶ review ─▶ merge
```

1. **An issue** says what is wrong or wanted, with the templates in `.github/ISSUE_TEMPLATE/`.
   Before starting, say in the issue that you take it, so two people never build the same thing.
2. **A proposal**, for anything bigger than a fix: one file in `docs/proposals/`, numbered next
   in line, with why, what (each requirement with its check), what is out of scope, the design
   and the tasks. The [proposals](../proposals/) already there are examples. The author approves the
   proposal, then the plan, before anything is built.
3. **A branch** per feature: `feature/<what-it-does>`, from the latest base branch.
4. **A pull request** when the tests pass, filling the template.
5. **A review** by someone who did not write it. The reviewer checks the proposal's requirements,
   the [style guide](5-style-guide.md), the tests and the docs.
6. **The merge**, by the author of the project, once the review is happy.

Nobody pushes straight to the base branch, and nobody rewrites history that someone else may have
pulled: no force-push on shared branches.

## Commits

A commit message is `area: what changed`, in the imperative, where the area is the package or
`docs`:

```
stand: ask which version is sold when a product has several
docs: explain how to test a flow
```

Never only "update", "fix" or "wip". A commit should build and pass its tests, so `git bisect`
works. A commit written with an AI assistant ends with a `Co-Authored-By:` line for it.

## Who decides

| Decision | Who |
|---|---|
| What the app does and how it looks | The author, through proposals |
| A new dependency (a module in `go.mod`) | The author: ask first |
| A new arrow between packages | The reviewer, in the pull request (see [the code](1-the-code.md)) |
| A change to Go UI, Go Renderer or Bubble | Their own repositories, with their own proposals |
| Everything inside the style guide | Nobody: it is already decided |

## Where parallel work clashes

Three people touching one app will meet in a few places. Each has a rule:

| Place | Why it clashes | Rule |
|---|---|---|
| `storage/migrations.go` | Steps are counted in order; two branches each adding a step would give two databases different histories | Add your step **last, right before merging**, on top of the latest base. Never edit a step that is merged |
| `storage/schema.sql` | It makes the first version's tables | Never change it: every change is a migration |
| `app/app.go` | Every package is wired here: places, `buildScreens`, `Show` | Keep each change to a few lines; rebase often |
| `look/` | Every screen depends on it | A change to a shared piece goes in its own small pull request, before the feature that needs it |
| `demo/` | Everyone adds toy data | Add, do not reshape; flow tests depend on it |
| `docs/proposals/` | Numbers | Take the next number when you open the pull request, not before |

Rebase your branch on the base before asking for a review (`git pull --rebase origin
feature/go-ui`), and again if the base moved, so conflicts are solved by the person who knows the
change.

## Before asking for a review

- [ ] `gofmt -l .` prints nothing, and `go vet ./...` and `go test ./...` pass.
- [ ] Every change in behaviour has its test.
- [ ] The docs that describe the behaviour changed in the same commit.
- [ ] A visual change was looked at on screen, in light and dark; the pull request says how.
- [ ] Nothing private: no real names, places or numbers of anyone's projects; examples are made
      up.
