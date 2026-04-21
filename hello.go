package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

var mdParser = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
)

const pageTemplate = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>%s</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Open+Sans:wght@400;600;700&display=swap" rel="stylesheet">
  <style>
    body { margin: 36px auto; max-width: 700px; padding: 0 14px; font: 17px/1.62 "Open Sans", -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif; color: #222; background: #fff; }
    main { max-width: 100%%; margin: 0 auto; }
    nav { margin-bottom: 1rem; font-size: 0.96em; }
    nav a { margin-right: 0.9rem; color: inherit; text-decoration: underline; }
    nav a.progress-link { color: #b00020; font-weight: 700; }
    h1, h2, h3 { line-height: 1.22; margin-top: 1.1em; margin-bottom: 0.4em; }
    h1 { font-size: 1.65em; }
    h2 { font-size: 1.28em; }
    h3 { font-size: 1.08em; }
    p, ul, ol, blockquote, pre { margin: 0.62em 0; }
    ul, ol { padding-left: 1.3em; }
    blockquote { margin-left: 0; padding-left: 0.75em; border-left: 2px solid #ccc; color: inherit; }
    pre { padding: 0.58em 0.78em; overflow-x: auto; border: 1px solid #ddd; font-size: 0.92em; }
    code { font-family: Consolas, "Courier New", monospace; font-size: 0.95em; }
    hr { border: 0; border-top: 1px solid #ddd; margin: 1.2em 0; }
    table { border-collapse: collapse; width: 100%%; margin: 0.9em 0; }
    th, td { border: 1px solid #ddd; padding: 0.35em 0.55em; text-align: left; }
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
		{Route: "/cv", SourcePath: "pages/cv.md", Title: "curriculum vitae"},
		{Route: "/portfolio", SourcePath: "pages/portfolio.md", Title: "portfolio"},
		{Route: "/skills", SourcePath: "pages/skills.md", Title: "skills"},
		{Route: "/socials", SourcePath: "pages/socials.md", Title: "socials"},
		{Route: "/literate-programming", SourcePath: "pages/literate-programming.md", Title: "literate programming"},
		{Route: "/microservices", SourcePath: "pages/microservices.md", Title: "microservices"},
		{Route: "/architecture", SourcePath: "pages/architecture.md", Title: "architecture"},
	}

	if *publish {
		if err := publishStatic(routes, *outDir, *domain); err != nil {
			log.Fatal(err)
		}
		log.Printf("published static site to %s", *outDir)
		return
	}

	byPath := make(map[string]sitePage, len(routes))
	for _, p := range routes {
		byPath[p.Route] = p
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("serving on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, siteHandler(byPath)); err != nil {
		log.Fatal(err)
	}
}

type sitePage struct {
	Route      string
	SourcePath string
	Title      string
}

// canonicalPath makes /other/ and /other the same, and normalizes "" to /.
func canonicalPath(p string) string {
	if p == "" || p == "/" {
		return "/"
	}
	return strings.TrimSuffix(p, "/")
}

func siteHandler(byPath map[string]sitePage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		page, ok := byPath[canonicalPath(r.URL.Path)]
		if !ok {
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
	body, err := markdownToHTML(content)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(pageTemplate, page.Title, body), nil
}

func markdownToHTML(src []byte) (string, error) {
	var buf bytes.Buffer
	if err := mdParser.Convert(src, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
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