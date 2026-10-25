# 🩺 Troubleshooting
---

This page is for whoever meets an error building, running or testing MUSA: find the message or
the code you see, and the fix under it.

## Building

### `replacement directory ../go-ui does not exist`

MUSA builds against checkouts of Go UI, Go Renderer and Bubble **beside** it (see `go.mod`). Clone
them next to MUSA, on the branches it is built with:

```bash
cd ..
git clone --branch feature/stacked https://github.com/akrck02/go-ui.git
git clone --branch feature/system-fonts https://github.com/akrck02/go-renderer.git
git clone --branch feature/v4 https://github.com/akrck02/bubble.git
```

### `go: go.mod requires go >= 1.26.0`

Your Go is older. Install a newer one from [go.dev/dl](https://go.dev/dl/).

### `fatal error: X11/Xlib.h` (or `GL/gl.h`, `Xrandr.h`…): no such file

The window is built with GLFW through cgo, which needs the system's headers:

```bash
sudo apt install gcc libgl1-mesa-dev xorg-dev        # Debian, Ubuntu
xcode-select --install                               # macOS
```

### `wayland-client-core.h: No such file` (or another Wayland header), on Linux

Without a tag, GLFW on Linux is built for X11 **and** Wayland, and needs both sets of headers. MUSA
builds for X11 only, which also runs under XWayland: add the tag to every command.

```bash
go run -tags=x11 ./cmd/musa -demo
go vet -tags=x11 ./...
go test -tags=x11 ./...
```

### `cgo: C compiler "gcc" not found`, or `CGO_ENABLED=0`

The window needs cgo. Install a C compiler (above) and make sure `CGO_ENABLED` is not `0`.

## Running

### `storage.open-failed`

The database file could not be opened: its folder cannot be written, or the file is not a
database. Try another file with `-database try.db`, or `-demo`, which keeps everything in memory.

### `database is locked`

Two MUSAs have the same database open. Close one. Tests never meet this: each opens its own
database in memory.

### The window stretches while resizing it, on a Mac

Known, and being worked on in Go Renderer's Metal backend. The picture is right again as soon as
the mouse is let go.

## Testing

### `uitest.element-not-found`

The test looked for a label that is not in the frame. The error lists every text the frame
showed: compare. The usual causes are a page that is not open yet (call `settle(window)` after a
step that animates), a date written by hand (compute it from the clock, see
[testing](8-contributing/4-testing.md)), or a label that changed.

### `uitest.element-ambiguous`

Two controls share the label. Give them unique labels (`"Quitar uno##"+id`) or describe them
with `context.Describe` (see [screens](8-contributing/3-screens.md)).

### `TestNamesAreWholeWords` (or another test in `internal/style`) fails

The code breaks a rule of the [style guide](8-contributing/5-style-guide.md); the message names
the file, the line and the rule.

### `TestPackagesImportOnlyWhatTheMapAllows` fails

A package imports one the [code map](8-contributing/1-the-code.md) does not allow. Either use a
function passed in (as `events` does to open the stand), or, if the new arrow is right, add it to
`docs/dependencies_test.go` and say why in the pull request.

## Every error code

The [error codes](5-reference/14-error-codes.md) page lists every code MUSA returns, and what
the artist reads for each problem; failures have no words, because the artist cannot fix them.
