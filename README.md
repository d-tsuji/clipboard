# clipboard

[![Actions Status](https://github.com/d-tsuji/clipboard/workflows/test/badge.svg)](https://github.com/d-tsuji/clipboard/actions)
[![Doc](https://img.shields.io/badge/doc-reference-blue.svg)](https://pkg.go.dev/github.com/d-tsuji/clipboard)
[![Go Report Card](https://goreportcard.com/badge/github.com/d-tsuji/clipboard)](https://goreportcard.com/report/github.com/d-tsuji/clipboard)

This is a multi-platform clipboard library in Go.

## Abstract

- This is clipboard library in Go, which runs on multiple platforms.
- External clipboard package is not required.

## Supported Platforms

- Windows (Pure Go)
- macOS (required cgo)

### ⚠WIP⚠

- Linux, Unix (X11)

*Unfortunately, I don't think it's feasible for Linux to build clipboard library, because xclient needs to be referenced as a daemon in order to keep the clipboard data. This approach is the same for [xclip](https://github.com/astrand/xclip) and [xsel](https://github.com/kfish/xsel).*

*Go has an approach to running its own programs as external processes, such as [VividCortex/godaemon](https://github.com/VividCortex/godaemon) and [sevlyar/go-daemon](https://github.com/sevlyar/go-daemon). But these cannot be incorporated as a library, of course. xclip and xsel can also be achieved because they are completed as binaries, not libraries.*

*So it turns out that it is not possible to achieve clipboard in Linux as a library.*

## Installation

```
go get github.com/d-tsuji/clipboard
```

## API

```go
package clipboard

// Get returns the current text data of the clipboard.
func Get() (string, error)

// Set sets the current text data of the clipboard.
func Set(text string) error
```

