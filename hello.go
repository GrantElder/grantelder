package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
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
	mux := http.NewServeMux()
	mux.HandleFunc("/", serveMarkdownPage("/", "pages/index.md", "grantelder.com"))
	mux.HandleFunc("/other", serveMarkdownPage("/other", "pages/other.md", "other ideas"))
	mux.HandleFunc("/progress", serveMarkdownPage("/progress", "pages/progress.md", "progress"))

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

func serveMarkdownPage(route, path, title string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != route {
			http.NotFound(w, r)
			return
		}

		content, err := os.ReadFile(path)
		if err != nil {
			http.Error(w, "could not read page", http.StatusInternalServerError)
			return
		}

		body := renderSimpleMarkdown(string(content))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, pageTemplate, title, body)
	}
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