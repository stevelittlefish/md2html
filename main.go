// md2html converts a Markdown file into a single, self-contained, dark-mode HTML page.
package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// version is overridden at build time with -ldflags "-X main.version=..."
var version = "dev"

const usage = `md2html %s - turn a Markdown file into a beautiful, self-contained dark-mode HTML page.

USAGE
  md2html <input.md> [options]
  md2html - [options]              read Markdown from stdin

OPTIONS
  -o, --output <file>   Write HTML to <file>. Default: <input>.html next to the
                        input file. With stdin input, the default is stdout.
                        Use "-" to force stdout.
  -t, --title <text>    Set the page title. Default: the first "# H1" heading,
                        or the input file name if there is none.
      --open            Open the result in the default browser afterwards.
  -h, --help            Show this help.
  -v, --version         Show the version.

OUTPUT
  One .html file with fonts, styles and syntax highlighting embedded: no
  network access needed to view it. On success the output path is printed
  to stdout (nothing is printed when writing HTML to stdout).

SUPPORTED MARKDOWN
  CommonMark plus GitHub extras: tables, task lists, strikethrough,
  autolinks, footnotes, fenced code with syntax highlighting (use a
  language tag, e.g. ` + "```go" + `). Raw HTML in the Markdown is passed through.
  A table of contents is generated when there are 3 or more "##" headings.

EXAMPLES
  md2html notes.md                      # -> notes.html
  md2html notes.md -o /tmp/notes.html
  md2html notes.md -t "Q3 plan" --open
  cat notes.md | md2html - > notes.html

EXIT CODES
  0 success, 1 conversion or I/O error, 2 bad arguments.
`

type options struct {
	input, output, title string
	open                 bool
}

func main() {
	opts, code := parseArgs(os.Args[1:])
	if code >= 0 {
		os.Exit(code)
	}
	if err := run(opts); err != nil {
		fmt.Fprintln(os.Stderr, "md2html:", err)
		os.Exit(1)
	}
}

// parseArgs returns options, or an exit code >= 0 if the program should stop.
func parseArgs(args []string) (options, int) {
	var o options
	if len(args) == 0 {
		fmt.Printf(usage, version)
		return o, 0
	}
	bad := func(format string, a ...any) (options, int) {
		fmt.Fprintf(os.Stderr, "md2html: "+format+"\nRun 'md2html --help' for usage.\n", a...)
		return o, 2
	}
	for i := 0; i < len(args); i++ {
		a := args[i]
		// Support --flag=value.
		name, val, hasVal := a, "", false
		if strings.HasPrefix(a, "--") {
			if k, v, ok := strings.Cut(a, "="); ok {
				name, val, hasVal = k, v, true
			}
		}
		needVal := func() (string, bool) {
			if hasVal {
				return val, true
			}
			if i+1 >= len(args) {
				return "", false
			}
			i++
			return args[i], true
		}
		switch name {
		case "-h", "--help":
			fmt.Printf(usage, version)
			return o, 0
		case "-v", "--version":
			fmt.Println("md2html", version)
			return o, 0
		case "-o", "--output":
			v, ok := needVal()
			if !ok {
				return bad("%s needs a file name", name)
			}
			o.output = v
		case "-t", "--title":
			v, ok := needVal()
			if !ok {
				return bad("%s needs some text", name)
			}
			o.title = v
		case "--open":
			o.open = true
		default:
			if strings.HasPrefix(a, "-") && a != "-" {
				return bad("unknown option %q", a)
			}
			if o.input != "" {
				return bad("only one input file is allowed (got %q and %q)", o.input, a)
			}
			o.input = a
		}
	}
	if o.input == "" {
		return bad("no input file given")
	}
	return o, -1
}

func run(o options) error {
	var src []byte
	var err error
	fallbackTitle := "Document"
	if o.input == "-" {
		src, err = io.ReadAll(os.Stdin)
	} else {
		src, err = os.ReadFile(o.input)
		base := filepath.Base(o.input)
		fallbackTitle = strings.TrimSuffix(base, filepath.Ext(base))
	}
	if err != nil {
		return err
	}

	html, err := Render(src, o.title, fallbackTitle)
	if err != nil {
		return err
	}

	out := o.output
	if out == "" {
		if o.input == "-" {
			out = "-"
		} else {
			out = strings.TrimSuffix(o.input, filepath.Ext(o.input)) + ".html"
		}
	}
	if out == "-" {
		_, err = os.Stdout.Write(html)
		if o.open {
			fmt.Fprintln(os.Stderr, "md2html: --open ignored when writing to stdout")
		}
		return err
	}
	if err := os.WriteFile(out, html, 0o644); err != nil {
		return err
	}
	abs, err := filepath.Abs(out)
	if err != nil {
		abs = out
	}
	fmt.Println(abs)
	if o.open {
		if err := openBrowser(abs); err != nil {
			fmt.Fprintln(os.Stderr, "md2html: could not open browser:", err)
		}
	}
	return nil
}

func openBrowser(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}
