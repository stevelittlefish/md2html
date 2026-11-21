# md2html

A small, opinionated command-line tool that turns a Markdown file into a single, self-contained, dark-mode HTML page that is comfortable to read.

- **One binary, no dependencies.** Fonts, styles and syntax highlighting are compiled in.
- **One HTML file out.** No CDN, no external fonts; it works offline and can be emailed or archived as-is.
- **Fast.** A typical document converts in about 20 ms.
- **No configuration.** There is one look: an ink-green reading page set in Literata, with a margin table of contents for longer documents.

## Usage

```
md2html <input.md> [options]
md2html - [options]              read Markdown from stdin

  -o, --output <file>   Output path. Default: <input>.html next to the input
                        (stdout when reading stdin). "-" forces stdout.
  -t, --title <text>    Page title. Default: first "# H1", else the file name.
      --open            Open the result in the default browser.
  -h, --help            Show help (also shown when run with no arguments).
  -v, --version         Show the version.
```

On success, the absolute path of the written file is printed to stdout. Exit codes: `0` success, `1` conversion or I/O error, `2` bad arguments.

```sh
md2html notes.md                     # writes notes.html
md2html notes.md -t "Q3 plan" --open
cat notes.md | md2html - > notes.html
```

### What is supported

CommonMark plus GitHub extras: tables, task lists, strikethrough, autolinks and footnotes. Fenced code blocks with a language tag get syntax highlighting. Raw HTML passes through. Documents with three or more `##` headings get a table of contents: it sits in the left margin on wide screens and collapses under the title on narrow ones. Printing produces dark text on white.

## Install

### Linux and Windows

Download the binary for your platform from the [Releases](../../releases) page, rename it to `md2html` (or `md2html.exe`), and put it somewhere on your `PATH`.

| Platform           | File                         |
|--------------------|------------------------------|
| Linux x86-64       | `md2html-linux-amd64`        |
| Linux ARM64        | `md2html-linux-arm64`        |
| Windows x86-64     | `md2html-windows-amd64.exe`  |
| Windows ARM64      | `md2html-windows-arm64.exe`  |

On Linux, run `chmod +x` on it after downloading.

### macOS (build from source)

There are no prebuilt Mac binaries, because unsigned downloads are blocked by Gatekeeper. Building locally takes one command:

```sh
git clone <this repo> md2html && cd md2html
./build.sh
sudo mv md2html /usr/local/bin/
```

`build.sh` uses your installed Go if it is version 1.24 or newer. Otherwise it downloads the official Go toolchain into `./.toolchain`, which needs no sudo and installs nothing system-wide. Only `curl` and `tar` are required, and both ship with macOS. The same script works on Linux, and `build.ps1` does the same job on Windows.

## Development

```sh
go test ./...                                   # tests
go run . testdata/sample.md -o /tmp/s.html      # render the visual test page
```

| Path                     | What it is                                         |
|--------------------------|----------------------------------------------------|
| `main.go`                | Argument parsing, help text, file I/O              |
| `render.go`              | Markdown → HTML, title and contents extraction, syntax theme |
| `assets/page.html`       | Page template                                      |
| `assets/style.css`       | The theme                                          |
| `assets/fonts/`          | Literata (subset to Latin, woff2) and its licence  |
| `testdata/sample.md`     | A document that uses every supported feature       |

### Releasing

Push a tag that starts with `v`:

```sh
git tag v1.0.0 && git push origin v1.0.0
```

The GitHub Action runs the tests, cross-compiles the Linux and Windows binaries, and attaches them to a GitHub Release with a `SHA256SUMS` file. Every push to `main` also runs the tests and a smoke test.

## Licence

[WTFPL](LICENCE). The bundled Literata font is under the SIL Open Font License; see `assets/fonts/OFL.txt`.
