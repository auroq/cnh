// Package qr renders the code a volunteer tapes beside a door sign: scan it and Wide Door knows where you stand.
package qr

import (
	"bytes"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	qrcode "github.com/skip2/go-qrcode"

	"github.com/auroq/cnh/pkg/venue"
)

// Kinds are the places someone stands at or in; a code beside their sign makes sense.
var Kinds = []string{"room", "entrance", "landmark", "quiet-room", "prayer-room", "nursing-room", "first-aid", "info-desk", "coat-check"}

const (
	pngSize   = 512
	quietZone = 4
)

// Codes renders codes that open the app with the location pre-filled.
type Codes struct {
	BaseURL string
}

// URL is what the code encodes: the home page with from set to the place id.
func (c Codes) URL(id string) string {
	return strings.TrimRight(c.BaseURL, "/") + "/?from=" + url.QueryEscape(id)
}

// Wanted reports whether a place gets a code.
func Wanted(p venue.Place) bool {
	for _, k := range Kinds {
		if p.Kind == k {
			return true
		}
	}
	return false
}

// Write renders <id>.svg and <id>.png for every wanted place into dir, and sheet.html when sheet is set.
// It returns the places that got a code.
func (c Codes) Write(dir string, places []venue.Place, sheet bool) ([]venue.Place, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create %s: %w", dir, err)
	}
	var written []venue.Place
	var pages []sheetPage
	for _, p := range places {
		if !Wanted(p) {
			continue
		}
		code, err := qrcode.New(c.URL(p.ID), qrcode.Medium)
		if err != nil {
			return written, fmt.Errorf("encode %s: %w", p.ID, err)
		}
		svg := SVG(code.Bitmap(), p.Name)
		if err := os.WriteFile(filepath.Join(dir, p.ID+".svg"), svg, 0o644); err != nil {
			return written, fmt.Errorf("write %s.svg: %w", p.ID, err)
		}
		png, err := code.PNG(pngSize)
		if err != nil {
			return written, fmt.Errorf("render %s.png: %w", p.ID, err)
		}
		if err := os.WriteFile(filepath.Join(dir, p.ID+".png"), png, 0o644); err != nil {
			return written, fmt.Errorf("write %s.png: %w", p.ID, err)
		}
		written = append(written, p)
		pages = append(pages, sheetPage{Place: p, URL: c.URL(p.ID), SVG: template.HTML(svg)})
	}
	if !sheet {
		return written, nil
	}
	var buf bytes.Buffer
	if err := sheetTmpl.Execute(&buf, pages); err != nil {
		return written, fmt.Errorf("render sheet: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sheet.html"), buf.Bytes(), 0o644); err != nil {
		return written, fmt.Errorf("write sheet.html: %w", err)
	}
	return written, nil
}

// SVG draws the bitmap as one path of unit squares inside a quiet zone, titled for a screen reader.
// The library already pads the bitmap with its own quiet zone, so the viewBox is the bitmap's size.
func SVG(bitmap [][]bool, name string) []byte {
	n := len(bitmap)
	var path strings.Builder
	for y, row := range bitmap {
		for x, dark := range row {
			if dark {
				fmt.Fprintf(&path, "M%d %dh1v1h-1z", x, y)
			}
		}
	}
	var b bytes.Buffer
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" shape-rendering="crispEdges" role="img" aria-labelledby="title">`, n, n)
	fmt.Fprintf(&b, "<title>Wide Door code for %s</title>", template.HTMLEscapeString(name))
	fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="#fff"/><path d="%s" fill="#000"/></svg>`, n, n, path.String())
	return b.Bytes()
}

type sheetPage struct {
	Place venue.Place
	URL   string
	SVG   template.HTML
}

var sheetTmpl = template.Must(template.New("sheet").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>Wide Door door codes</title>
<style>
@page { size: A4; margin: 15mm; }
body { font-family: system-ui, -apple-system, "Segoe UI", Roboto, sans-serif; color: #000; background: #fff; margin: 0; }
.page { break-after: page; page-break-after: always; display: flex; flex-direction: column; align-items: center; text-align: center; min-height: 90vh; justify-content: center; }
.name { font-size: 48pt; line-height: 1.1; margin: 0 0 12pt; }
.where { font-size: 24pt; margin: 0 0 24pt; }
svg { width: 120mm; height: 120mm; }
.url { font-size: 16pt; font-family: ui-monospace, Menlo, Consolas, monospace; margin: 18pt 0 0; word-break: break-all; }
.how { font-size: 18pt; margin: 12pt 0 0; }
.braille { font-size: 14pt; color: #333; margin-top: 18pt; }
</style>
</head>
<body>
{{range .}}<section class="page">
<h1 class="name">{{.Place.Name}}</h1>
<p class="where">{{.Place.Venue}}{{if .Place.Level}}, {{.Place.Level}}{{end}}</p>
{{.SVG}}
<p class="url">{{.URL}}</p>
<p class="how">Scan to ask Wide Door where the nearest elevator, quiet room or all-gender restroom is from here.</p>
<p class="braille">Braille label text: {{.Place.Name}}. {{.Place.Venue}}{{if .Place.Level}} {{.Place.Level}}{{end}}. Wide Door code {{.Place.ID}}.</p>
</section>
{{end}}</body>
</html>
`))
