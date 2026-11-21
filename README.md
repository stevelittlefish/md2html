# md2html

A small, opinionated command-line tool that turns a Markdown file into a single, self-contained, dark-mode HTML page that is comfortable to read.

- **One binary, no dependencies.** Fonts, styles and syntax highlighting are compiled in.
- **One HTML file out.** No CDN, no external fonts; it works offline and can be emailed or archived as-is.
- **Fast.** A typical document converts in about 20 ms.
- **No configuration.** There is one look, "Field Manual": a deep-navy page set in IBM Plex Sans with JetBrains Mono for code, and cyan accents and a margin table of contents for longer documents.

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

All downloads are on the [Releases](https://github.com/stevelittlefish/md2html/releases) page. `X.Y.Z` below is the version number.

### Linux

Download `md2html-vX.Y.Z-linux-amd64.tar.gz`, or `-linux-arm64` for ARM. Then:

```sh
tar -xzf md2html-vX.Y.Z-linux-amd64.tar.gz
sudo install md2html /usr/local/bin/
```

### Windows

Download `md2html-vX.Y.Z-windows-amd64.zip`, or `-windows-arm64` for ARM. Unzip it and put `md2html.exe` in a folder on your `PATH`.

### macOS

There is no prebuilt Mac binary, because unsigned downloads get blocked by Gatekeeper. Instead, download `md2html-vX.Y.Z-macos.tar.gz`, which contains the source and an install script, and run:

```sh
tar -xzf md2html-vX.Y.Z-macos.tar.gz
cd md2html-vX.Y.Z-macos
./install.sh                 # builds and installs to /usr/local/bin (asks for sudo)
./install.sh ~/.local/bin    # or install anywhere you like
```

`install.sh` runs `build.sh`, which uses your installed Go if it is new enough. If it isn't, `build.sh` downloads the official Go toolchain into the unpacked folder. That needs no sudo and installs nothing system-wide. Only `curl` and `tar` are required, and both ship with macOS. The same steps work from a git clone, and on Linux. On Windows, `build.ps1` builds from source.

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
| `assets/fonts/`          | IBM Plex Sans and JetBrains Mono (subset to Latin, woff2) and their licences  |
| `testdata/sample.md`     | A document that uses every supported feature       |

### Releasing

```sh
./make_release.sh            # shows the latest tag and suggests the next one
./make_release.sh v1.2.0     # tags HEAD and pushes the tag
```

Pushing a `v*` tag triggers the **Build** workflow (`.github/workflows/build.yml`), which:

1. Checks formatting, then runs `go vet` and the tests, plus a smoke test.
2. Builds Linux and Windows binaries for amd64 and arm64 on native runners, smoke-tests each one, and packages it as a `.tar.gz` or `.zip`.
3. Packages the source with `install.sh` as `md2html-vX.Y.Z-macos.tar.gz`, then checks that the package builds and installs.
4. Creates the GitHub Release and attaches all the downloads plus a `SHA256SUMS` file.

Pushes to `main` and pull requests run the same checks and builds, but don't publish anything.

## Licence

[WTFPL](LICENCE). The bundled IBM Plex Sans and JetBrains Mono fonts are under the SIL Open Font License; see `assets/fonts/`.
