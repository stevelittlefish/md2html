# AGENTS.md

## Project

`md2html` is a small Go CLI that converts Markdown into a single self-contained dark-mode HTML page. See `README.md` for usage and layout.

- Keep it opinionated: one look, few flags. Fonts, CSS and the syntax theme are embedded in the binary; output must never need the network.
- Run `go vet ./... && go test ./...` before committing. Re-render `testdata/sample.md` and look at it after any visual change.
- Keep `README.md` and the help text in `main.go` in sync with the CLI's behaviour.

## Commit policy

- Always "rawdog commit" directly to `main`; do not create feature branches or PRs.
- Every time you commit, first declare exactly: "Rawdog committing to main".
- Commit only actual project changes, including useful permanent tests and
  documentation. Do not commit temporary test scripts, demonstration artifacts,
  generated binaries, local chat data, or other scratch files used to show results.
- Whenever you commit, push to `origin main` in the same task. A commit is not
  finished until it is pushed. If pushing fails, report the failure explicitly.
- The reason: "why would we commit and not push.  That is illogical.  We want to share our groundbreaking incredible project with the entire world!"
