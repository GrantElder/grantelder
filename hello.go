package main

import (
	"flag"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const pageTemplate = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>%s</title>
  <style>
    body { margin: 0; padding: 2rem 1rem; font-family: Georgia, "Times New Roman", serif; color: #111; background: #fff; }
    main { max-width: 760px; margin: 0 auto; line-height: 1.55; }
    nav { margin-bottom: 1.5rem; }
    nav a { margin-right: 1rem; color: #1a1a1a; }
    nav a.progress-link { color: #b00020; font-weight: 700; background: #ffe8ee; padding: 0.1rem 0.4rem; border-radius: 4px; }
    h1, h2, h3 { line-height: 1.2; margin-top: 1.3em; margin-bottom: 0.5em; }
    p { margin: 0.7em 0; }
    ul { margin: 0.5em 0 0.8em 1.4em; }
    blockquote { margin: 0.8em 0; padding-left: 0.8em; border-left: 3px solid #ddd; color: #444; }
  </style>
</head>
<body>
  <main>
    <nav>
      <a href="/">home</a>
      <a href="/other">other</a>
      <a class="progress-link" href="/progress">progress</a>
    </nav>
    %s
  </main>
</body>
</html>`

func main() {
	publish := flag.Bool("publish", false, "write static site output to docs/")
	outDir := flag.String("out", "docs", "output directory for static publish")
	domain := flag.String("domain", "grantelder.com", "custom domain written to CNAME when publishing")
	flag.Parse()

	routes := []sitePage{
		{Route: "/", SourcePath: "pages/index.md", Title: "grantelder.com"},
		{Route: "/other", SourcePath: "pages/other.md", Title: "other ideas"},
		{Route: "/progress", SourcePath: "pages/progress.md", Title: "progress"},
	}

	if *publish {
		if err := publishStatic(routes, *outDir, *domain); err != nil {
			log.Fatal(err)
		}
		log.Printf("published static site to %s", *outDir)
		return
	}

	mux := http.NewServeMux()
	for _, p := range routes {
		mux.HandleFunc(p.Route, serveMarkdownPage(p))
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("serving on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

type sitePage struct {
	Route      string
	SourcePath string
	Title      string
}

func serveMarkdownPage(page sitePage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != page.Route {
			http.NotFound(w, r)
			return
		}

		htmlPage, err := renderPage(page)
		if err != nil {
			http.Error(w, "could not read page", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, htmlPage)
	}
}

func renderPage(page sitePage) (string, error) {
	content, err := os.ReadFile(page.SourcePath)
	if err != nil {
		return "", err
	}
	body := renderSimpleMarkdown(string(content))
	return fmt.Sprintf(pageTemplate, page.Title, body), nil
}

func publishStatic(routes []sitePage, outDir, domain string) error {
	if err := os.RemoveAll(outDir); err != nil {
		return err
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	for _, page := range routes {
		htmlPage, err := renderPage(page)
		if err != nil {
			return err
		}
		targetPath := outputFilePath(outDir, page.Route)
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(targetPath, []byte(htmlPage), 0644); err != nil {
			return err
		}
	}

	// Keep custom domain mapping stable on GitHub Pages.
	return os.WriteFile(filepath.Join(outDir, "CNAME"), []byte(domain+"\n"), 0644)
}

func outputFilePath(outDir, route string) string {
	if route == "/" {
		return filepath.Join(outDir, "index.html")
	}
	trimmed := strings.TrimPrefix(route, "/")
	return filepath.Join(outDir, trimmed, "index.html")
}

func renderSimpleMarkdown(src string) string {
	lines := strings.Split(src, "\n")
	var b strings.Builder
	inList := false

	closeList := func() {
		if inList {
			b.WriteString("</ul>\n")
			inList = false
		}
	}

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			closeList()
			continue
		}

		switch {
		case strings.HasPrefix(line, "### "):
			closeList()
			b.WriteString("<h3>" + formatInline(line[4:]) + "</h3>\n")
		case strings.HasPrefix(line, "## "):
			closeList()
			b.WriteString("<h2>" + formatInline(line[3:]) + "</h2>\n")
		case strings.HasPrefix(line, "# "):
			closeList()
			b.WriteString("<h1>" + formatInline(line[2:]) + "</h1>\n")
		case strings.HasPrefix(line, "- "):
			if !inList {
				b.WriteString("<ul>\n")
				inList = true
			}
			b.WriteString("<li>" + formatInline(line[2:]) + "</li>\n")
		case strings.HasPrefix(line, "> "):
			closeList()
			b.WriteString("<blockquote>" + formatInline(line[2:]) + "</blockquote>\n")
		default:
			closeList()
			b.WriteString("<p>" + formatInline(line) + "</p>\n")
		}
	}

	closeList()
	return b.String()
}

func formatInline(s string) string {
	escaped := html.EscapeString(s)
	// support [text](url) style links
	for {
		openText := strings.Index(escaped, "[")
		midText := strings.Index(escaped, "](")
		closeText := strings.Index(escaped, ")")
		if openText == -1 || midText == -1 || closeText == -1 || !(openText < midText && midText < closeText) {
			break
		}

		label := escaped[openText+1 : midText]
		href := escaped[midText+2 : closeText]
		link := `<a href="` + href + `">` + label + `</a>`
		escaped = escaped[:openText] + link + escaped[closeText+1:]
	}
	return escaped
}