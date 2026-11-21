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
	//go:embed assets/fonts/literata.woff2
	fontRegular []byte
	//go:embed assets/fonts/literata-italic.woff2
	fontItalic []byte
)

// codeStyle is the syntax theme, built from the page palette.
var codeStyle = chroma.MustNewStyle("md2html", chroma.StyleEntries{
	chroma.Background:           "#D8D3C4 bg:#0D1311",
	chroma.Text:                 "#D8D3C4",
	chroma.Error:                "#E08A8A",
	chroma.Comment:              "italic #7C9285",
	chroma.CommentPreproc:       "noitalic #8DA595",
	chroma.Keyword:              "#C7A35A",
	chroma.KeywordType:          "#9FB7B0",
	chroma.Operator:             "#B8B2A2",
	chroma.Punctuation:          "#8F8A7D",
	chroma.Name:                 "#D8D3C4",
	chroma.NameFunction:         "#ECE6D6",
	chroma.NameClass:            "#ECE6D6",
	chroma.NameBuiltin:          "#9FB7B0",
	chroma.NameTag:              "#C7A35A",
	chroma.NameAttribute:        "#8DA595",
	chroma.NameDecorator:        "#8DA595",
	chroma.LiteralString:        "#C99A8E",
	chroma.LiteralStringEscape:  "#C7A35A",
	chroma.LiteralNumber:        "#B9C18F",
	chroma.GenericDeleted:       "#E08A8A",
	chroma.GenericInserted:      "#9BC29B",
	chroma.GenericHeading:       "bold #ECE6D6",
	chroma.GenericSubheading:    "#C7A35A",
	chroma.GenericEmph:          "italic",
	chroma.GenericStrong:        "bold",
	chroma.LiteralStringRegex:   "#C7A35A",
	chroma.LiteralStringSymbol:  "#B9C18F",
	chroma.KeywordConstant:      "#B9C18F",
	chroma.NameVariableMagic:    "#9FB7B0",
	chroma.LiteralStringDoc:     "italic #7C9285",
	chroma.CommentSpecial:       "bold italic #8DA595",
	chroma.GenericPrompt:        "#7C9285",
	chroma.GenericOutput:        "#B8B2A2",
	chroma.GenericTraceback:     "#E08A8A",
	chroma.NameException:        "#E08A8A",
	chroma.NameLabel:            "#C7A35A",
	chroma.NameNamespace:        "#9FB7B0",
	chroma.NameEntity:           "#C7A35A",
	chroma.NameConstant:         "#B9C18F",
	chroma.NameOther:            "#D8D3C4",
	chroma.NameProperty:         "#D8D3C4",
	chroma.KeywordNamespace:     "#C7A35A",
	chroma.KeywordDeclaration:   "#C7A35A",
	chroma.KeywordPseudo:        "#C7A35A",
	chroma.KeywordReserved:      "#C7A35A",
	chroma.LiteralStringChar:    "#C99A8E",
	chroma.LiteralStringHeredoc: "#C99A8E",
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
	fontFace(&css, "normal", fontRegular)
	fontFace(&css, "italic", fontItalic)
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

func fontFace(b *strings.Builder, style string, data []byte) {
	fmt.Fprintf(b, "@font-face{font-family:\"Literata\";font-style:%s;font-weight:400 700;font-display:swap;"+
		"src:url(data:font/woff2;base64,%s) format(\"woff2\")}\n", style, base64.StdEncoding.EncodeToString(data))
}
