package main

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"strings"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var (
	//go:embed assets/style.css
	styleCSS string
	//go:embed assets/page.html
	pageTemplate string
	//go:embed assets/fonts/ibm-plex-sans.woff2
	fontSans []byte
	//go:embed assets/fonts/ibm-plex-sans-italic.woff2
	fontSansItalic []byte
	//go:embed assets/fonts/jetbrains-mono.woff2
	fontMono []byte
	//go:embed assets/fonts/jetbrains-mono-italic.woff2
	fontMonoItalic []byte
)

// codeStyle is the syntax theme, built from the page palette.
var codeStyle = chroma.MustNewStyle("md2html", chroma.StyleEntries{
	chroma.Background:           "#C3CCDC bg:#0A0E17",
	chroma.Text:                 "#C3CCDC",
	chroma.Error:                "#E8848E",
	chroma.Comment:              "italic #6B7891",
	chroma.CommentPreproc:       "noitalic #8FA3C2",
	chroma.Keyword:              "#7FC8D8",
	chroma.KeywordType:          "#A9B7F0",
	chroma.Operator:             "#AEB8CA",
	chroma.Punctuation:          "#7D8AA3",
	chroma.Name:                 "#C3CCDC",
	chroma.NameFunction:         "#E8EEF8",
	chroma.NameClass:            "#E8EEF8",
	chroma.NameBuiltin:          "#A9B7F0",
	chroma.NameTag:              "#7FC8D8",
	chroma.NameAttribute:        "#8FA3C2",
	chroma.NameDecorator:        "#8FA3C2",
	chroma.LiteralString:        "#B8D99A",
	chroma.LiteralStringEscape:  "#7FC8D8",
	chroma.LiteralNumber:        "#E5B97A",
	chroma.GenericDeleted:       "#E8848E",
	chroma.GenericInserted:      "#9DD6A0",
	chroma.GenericHeading:       "bold #E8EEF8",
	chroma.GenericSubheading:    "#7FC8D8",
	chroma.GenericEmph:          "italic",
	chroma.GenericStrong:        "bold",
	chroma.LiteralStringRegex:   "#7FC8D8",
	chroma.LiteralStringSymbol:  "#E5B97A",
	chroma.KeywordConstant:      "#E5B97A",
	chroma.NameVariableMagic:    "#A9B7F0",
	chroma.LiteralStringDoc:     "italic #6B7891",
	chroma.CommentSpecial:       "bold italic #8FA3C2",
	chroma.GenericPrompt:        "#6B7891",
	chroma.GenericOutput:        "#AEB8CA",
	chroma.GenericTraceback:     "#E8848E",
	chroma.NameException:        "#E8848E",
	chroma.NameLabel:            "#7FC8D8",
	chroma.NameNamespace:        "#A9B7F0",
	chroma.NameEntity:           "#7FC8D8",
	chroma.NameConstant:         "#E5B97A",
	chroma.NameOther:            "#C3CCDC",
	chroma.NameProperty:         "#C3CCDC",
	chroma.KeywordNamespace:     "#7FC8D8",
	chroma.KeywordDeclaration:   "#7FC8D8",
	chroma.KeywordPseudo:        "#7FC8D8",
	chroma.KeywordReserved:      "#7FC8D8",
	chroma.LiteralStringChar:    "#B8D99A",
	chroma.LiteralStringHeredoc: "#B8D99A",
})

var md = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM, // tables, task lists, strikethrough, autolinks
		extension.Footnote,
		highlighting.NewHighlighting(
			highlighting.WithCustomStyle(codeStyle),
			highlighting.WithFormatOptions(chromahtml.WithClasses(true)),
		),
	),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	goldmark.WithRendererOptions(
		html.WithUnsafe(),
		renderer.WithNodeRenderers(util.Prioritized(&headingRenderer{}, 100)),
	),
)

type tocEntry struct {
	ID, Text string
}

type pageData struct {
	Title      string
	ShowHeader bool // true when the document has no H1 of its own
	TOC        []tocEntry
	Body       template.HTML
	CSS        template.CSS
}

var page = template.Must(template.New("page").Parse(pageTemplate))

// Render converts Markdown to a complete HTML page.
func Render(src []byte, titleOverride, fallbackTitle string) ([]byte, error) {
	doc := md.Parser().Parse(text.NewReader(src))

	var firstH1 string
	var toc []tocEntry
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		h, ok := n.(*ast.Heading)
		if !entering || !ok {
			return ast.WalkContinue, nil
		}
		switch h.Level {
		case 1:
			if firstH1 == "" {
				firstH1 = plainText(h, src)
			}
		case 2:
			if id, ok := h.AttributeString("id"); ok {
				toc = append(toc, tocEntry{ID: string(id.([]byte)), Text: plainText(h, src)})
			}
		}
		return ast.WalkSkipChildren, nil
	})
	if len(toc) < 3 {
		toc = nil
	}

	var body bytes.Buffer
	if err := md.Renderer().Render(&body, src, doc); err != nil {
		return nil, err
	}

	var css strings.Builder
	fontFace(&css, "IBM Plex Sans", "normal", fontSans)
	fontFace(&css, "IBM Plex Sans", "italic", fontSansItalic)
	fontFace(&css, "JetBrains Mono", "normal", fontMono)
	fontFace(&css, "JetBrains Mono", "italic", fontMonoItalic)
	css.WriteString(styleCSS)
	if err := chromahtml.New(chromahtml.WithClasses(true)).WriteCSS(&css, codeStyle); err != nil {
		return nil, err
	}

	title := titleOverride
	if title == "" {
		title = firstH1
	}
	if title == "" {
		title = fallbackTitle
	}

	var out bytes.Buffer
	err := page.Execute(&out, pageData{
		Title:      title,
		ShowHeader: firstH1 == "",
		TOC:        toc,
		Body:       template.HTML(body.String()),
		CSS:        template.CSS(css.String()),
	})
	return out.Bytes(), err
}

// plainText returns the visible text of a node, without markup.
func plainText(n ast.Node, src []byte) string {
	var b strings.Builder
	_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch t := c.(type) {
		case *ast.Text:
			b.Write(t.Segment.Value(src))
			if t.SoftLineBreak() || t.HardLineBreak() {
				b.WriteByte(' ')
			}
		case *ast.String:
			b.Write(t.Value)
		case *ast.CodeSpan:
			for cc := t.FirstChild(); cc != nil; cc = cc.NextSibling() {
				if tt, ok := cc.(*ast.Text); ok {
					b.Write(tt.Segment.Value(src))
				}
			}
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	return strings.TrimSpace(b.String())
}

// headingRenderer adds a § self-link to h2/h3 headings.
type headingRenderer struct{}

func (r *headingRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindHeading, r.render)
}

func (r *headingRenderer) render(w util.BufWriter, src []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	if entering {
		fmt.Fprintf(w, "<h%d", n.Level)
		if n.Attributes() != nil {
			html.RenderAttributes(w, node, html.HeadingAttributeFilter)
		}
		w.WriteByte('>')
		if id, ok := n.AttributeString("id"); ok && (n.Level == 2 || n.Level == 3) {
			fmt.Fprintf(w, `<a class="anchor" href="#%s" aria-label="Link to this section">§</a>`,
				util.EscapeHTML(id.([]byte)))
		}
	} else {
		fmt.Fprintf(w, "</h%d>\n", n.Level)
	}
	return ast.WalkContinue, nil
}

func fontFace(b *strings.Builder, family, style string, data []byte) {
	fmt.Fprintf(b, "@font-face{font-family:\"%s\";font-style:%s;font-weight:400 700;font-display:swap;"+
		"src:url(data:font/woff2;base64,%s) format(\"woff2\")}\n", family, style, base64.StdEncoding.EncodeToString(data))
}
