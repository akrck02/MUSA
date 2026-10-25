module github.com/Itros97/MUSA

go 1.26.0

require (
	github.com/akrck02/bubble v0.0.0
	github.com/akrck02/bubble/goui v0.0.0
	github.com/akrck02/go-renderer v0.2.1-0.20261022213249-8ecc924ab3e5
	github.com/akrck02/go-ui v0.1.1
	modernc.org/sqlite v1.61.0
)

require (
	github.com/dustin/go-humanize v1.1.0 // indirect
	github.com/go-gl/gl v0.0.0-20260331235117-4566fea9a276 // indirect
	github.com/go-gl/glfw/v3.4/glfw v0.1.0-pre.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/ncruces/go-strftime v1.1.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/image v0.45.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	modernc.org/libc v1.79.0 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.12.1 // indirect
)

// Go UI, Go Renderer and Bubble's go-ui adapter are built from checkouts beside this one, as Go UI
// for Android and Bubble do, until they publish versions that work together (see
// docs/2-get-started.md).
replace (
	github.com/akrck02/bubble => ../bubble
	github.com/akrck02/bubble/goui => ../bubble/goui
	github.com/akrck02/go-renderer => ../go-renderer
	github.com/akrck02/go-ui => ../go-ui
)

