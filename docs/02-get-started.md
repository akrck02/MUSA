# Get started
---

This page is for whoever wants to run MUSA on their computer: by the end it is open, with toy data in it.

## Requirements

- Go 1.26 or newer.
- Checkouts of Go UI, Go Renderer and Bubble beside this one (`../go-ui`, `../go-renderer`,
  `../bubble`), until they publish versions that work together. `go.mod` points at them.
- On Linux, the usual OpenGL and X11 development packages that GLFW needs.

## Run the app

```bash
go run ./cmd/musa                     # the app, over your own database
go run ./cmd/musa -demo               # a throwaway database in memory, filled with toy data
go run ./cmd/musa -database my.db     # another database file
```

On Linux, add `-tags=x11` (`go run -tags=x11 ./cmd/musa -demo`): GLFW there is built either for
X11 or for Wayland, and X11 also runs under XWayland.

The demo lives only in memory: close the window and it is gone, so you can break things freely.

For a phone, Go UI for Android's `go-ui-apk` packages the same program as an APK.

## Run the tests

```bash
gofmt -l .        # prints nothing when everything is formatted
go vet ./...
go test ./...
```

The screen tests drive the real app in a window that lives in memory, so they need no display.

## Development tools

The home screen shows a **Herramientas de prueba** card (turn it off with `-tools=false`):

- **Rellenar con datos de prueba** loads the demo's toy data. It only works on an empty database.
- **Vaciar base de datos** deletes every row after asking. It cannot be undone.

## Where the data lives

The database is one file, `musa.db`, in MUSA's folder inside the user's configuration folder:

| System  | Folder                                    |
|---------|-------------------------------------------|
| Linux   | `$XDG_CONFIG_HOME/musa` or `~/.config/musa` |
| macOS   | `~/Library/Application Support/musa`      |
| Windows | `%AppData%\musa`                          |

On Android it lives in the app's private storage. An older database is brought up to date when it
opens: each change to the tables is a numbered migration in `storage/migrations.go`.
