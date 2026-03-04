// grantelder.com v0 — markdown to site
// Reads markdown files and publishes to static HTML.
package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/yuin/goldmark"
)

var (
	contentDir = flag.String("content", ".", "directory containing markdown files")
	outputDir  = flag.String("output", "docs", "output directory for published HTML")
	serve      = flag.String("serve", "", "if set, serve on this address (e.g. :8080)")
)

var routes = map[string]string{
	"/":                    "grantelder.com.md",
	"/cv":                  "cv.md",
	"/portfolio":           "portfolio.md",
	"/skills-and-values":   "skills-and-values.md",
	"/connections":         "connections.md",
	"/literate-programming": "literate-programming.md",
	"/dependencies":       "dependencies.md",
	"/architecture":       "architecture.md",
}

func main() {
	flag.Parse()

	if *serve != "" {
		runServer()
		return
	}
	publish()
}

func publish() {
	md := goldmark.New()
	tmpl := template.Must(template.New("page").Parse(pageTemplate))

	for path, file := range routes {
		src := filepath.Join(*contentDir, file)
		mdBytes, err := os.ReadFile(src)
		if err != nil {
			log.Printf("skip %s: %v", file, err)
			continue
		}

		var htmlBody string
		if err := md.Convert(mdBytes, &stringWriter{&htmlBody}); err != nil {
			log.Printf("convert %s: %v", file, err)
			continue
		}

		outPath := filepath.Join(*outputDir, pathToFile(path))
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			log.Fatal(err)
		}

		f, err := os.Create(outPath)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		if err := tmpl.Execute(f, map[string]any{"Body": template.HTML(htmlBody)}); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("published %s -> %s\n", file, outPath)
	}
	fmt.Printf("done. output in %s/\n", *outputDir)
}

func pathToFile(path string) string {
	if path == "/" {
		return "index.html"
	}
	return path[1:] + ".html"
}

func runServer() {
	md := goldmark.New()
	tmpl := template.Must(template.New("page").Parse(pageTemplate))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path != "/" && path[len(path)-1] == '/' {
			path = path[:len(path)-1]
		}
		file, ok := routes[path]
		if !ok {
			http.NotFound(w, r)
			return
		}

		src := filepath.Join(*contentDir, file)
		mdBytes, err := os.ReadFile(src)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var htmlBody string
		if err := md.Convert(mdBytes, &stringWriter{&htmlBody}); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, map[string]any{"Body": template.HTML(htmlBody)})
	})

	fmt.Printf("serving at http://localhost%s\n", *serve)
	log.Fatal(http.ListenAndServe(*serve, nil))
}

type stringWriter struct{ s *string }

func (w *stringWriter) Write(p []byte) (n int, err error) {
	*w.s += string(p)
	return len(p), nil
}

const pageTemplate = `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>grantelder.com</title></head>
<body>
{{.Body}}
</body>
</html>`
