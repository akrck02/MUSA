# Contributing to MUSA

Thank you for helping. Everything you need is in
[docs/8-contributing](docs/8-contributing/1-the-code.md):

1. [The code](docs/8-contributing/1-the-code.md): where everything lives, and which package may use which.
2. [A feature, step by step](docs/8-contributing/2-a-feature-step-by-step.md): a real one, from the test to the pull request.
3. [Screens](docs/8-contributing/3-screens.md): how a page is drawn, keeps its state and opens another.
4. [Testing](docs/8-contributing/4-testing.md): which test to write, and how to run and watch it.
5. [Style guide](docs/8-contributing/5-style-guide.md): how the code is written, and which rules a test checks.
6. [Working together](docs/8-contributing/6-working-together.md): branches, proposals, reviews, and where parallel work clashes.

In short: `gofmt -l .` prints nothing and `go vet ./...` and `go test ./...` pass before a pull
request; every change in behaviour comes with its test and its docs; anything larger than a fix
starts as a proposal in [docs/proposals](docs/proposals/).

Found a bug, or have an idea? Open an issue: the templates ask for what helps.
