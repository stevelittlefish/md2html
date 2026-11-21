package main

import (
	"os"
	"strings"
	"testing"
)

func TestRenderSample(t *testing.T) {
	src, err := os.ReadFile("testdata/sample.md")
	if err != nil {
		t.Fatal(err)
	}
	out, err := Render(src, "", "sample")
	if err != nil {
		t.Fatal(err)
	}
	html := string(out)
	for _, want := range []string{
		"<title>Migrating the ingest pipeline to streaming</title>",
		`<nav class="toc"`,
		`class="anchor" href="#why-change"`,
		"<table>",
		`type="checkbox"`,
		`class="chroma"`,
		"font/woff2;base64,",
		`class="footnotes"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("output missing %q", want)
		}
	}
}

func TestTitleRules(t *testing.T) {
	cases := []struct {
		src, override, want string
		header              bool
	}{
		{"# Hello *world*\n\ntext", "", "Hello world", false},
		{"# Hello\n", "Custom", "Custom", false},
		{"just text", "", "fallback", true},
		{"just text", "Given", "Given", true},
	}
	for _, c := range cases {
		out, err := Render([]byte(c.src), c.override, "fallback")
		if err != nil {
			t.Fatal(err)
		}
		s := string(out)
		if !strings.Contains(s, "<title>"+c.want+"</title>") {
			t.Errorf("%q: want title %q", c.src, c.want)
		}
		if got := strings.Contains(s, "<h1>"+c.want+"</h1>"); got != c.header {
			t.Errorf("%q: generated header = %v, want %v", c.src, got, c.header)
		}
	}
}

func TestNoTOCForShortDocs(t *testing.T) {
	out, _ := Render([]byte("# T\n\n## A\n\n## B\n"), "", "x")
	if strings.Contains(string(out), `<nav class="toc"`) {
		t.Error("TOC should need 3+ sections")
	}
}
